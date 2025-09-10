// Package mpris
package mpris

import (
	"context"
	"fmt"
	"strings"

	"github.com/godbus/dbus/v5"
)

type MPRISClient struct {
	Conn         *dbus.Conn
	Players      map[string]*PlayerInfo // service name -> player info
	ActivePlayer string                 // currently active player
	Subscribers  []chan<- PlayerInfo    // for notifications
}

type MPRISError struct {
	Service string
	Method  string
	Err     error
}

func NewMPRISClient() (*MPRISClient, error) {
	newConn, err := dbus.ConnectSessionBus()
	if err != nil {
		return nil, fmt.Errorf("failed to create dbus connection: %s", err)
	}

	// fmt.Println("DBUS connect: ", newConn.Connected())

	return &MPRISClient{
		Conn:    newConn,
		Players: make(map[string]*PlayerInfo),
	}, nil
}

func (m *MPRISClient) ListPlayers() ([]string, error) {
	var names []string
	err := m.Conn.Object(DbusDestination, DBusObjectPath).Call(DbusListNames, 0).Store(&names)
	if err != nil {
		return nil, fmt.Errorf("failed to get names: %s", err)
	}

	var mprisPlayers []string
	for _, name := range names {
		if strings.HasPrefix(name, MPRISNamespace) {
			mprisPlayers = append(mprisPlayers, name)
		}
	}

	if len(mprisPlayers) == 0 {
		return nil, fmt.Errorf("no MPRIS players found")
	}

	// fmt.Println(mprisPlayers) // Debug
	return mprisPlayers, nil
}

func (m *MPRISClient) Subscriber(subscriber chan<- PlayerInfo) {
	m.Subscribers = append(m.Subscribers, subscriber)
}

func (m *MPRISClient) StartListening(ctx context.Context, updates chan<- PlayerInfo) error {
	listener := make(chan *dbus.Signal, 10)

	m.Conn.Signal(listener)

	err := m.Conn.AddMatchSignal(
		dbus.WithMatchObjectPath(MPRISObjectPath),
		dbus.WithMatchInterface(PropertiesInterface),
		dbus.WithMatchMember("PropertiesChanged"),
	)
	if err != nil {
		return fmt.Errorf("failed to add signal match for PropertiesChanged: %w", err)
	}

	err = m.Conn.AddMatchSignal(
		dbus.WithMatchInterface(DbusInterface),
		dbus.WithMatchMember("NameOwnerChanged"),
	)
	if err != nil {
		return fmt.Errorf("failed to add signal match for NameOwnerChanged: %w", err)
	}

	// fmt.Println("Listening...")

	go func() {
		for {
			select {
			case signal := <-listener:
				switch signal.Name {
				case PropertiesChanged:
					playerInfo, err := m.GetPlayerInfo(signal.Sender)
					if err != nil {
						fmt.Printf("Error getting updated player info: %s\n", err)
						continue
					}

					m.Players[signal.Sender] = playerInfo

					if playerInfo.Status.PlaybackStatus == "Playing" {
						m.ActivePlayer = signal.Sender
					}

					updates <- *playerInfo
					// for _, subscriber := range m.Subscribers {
					// 	subscriber <- *playerInfo
					// }
				case NameOwnerChanged:
					if len(signal.Body) >= 3 {
						serviceName := signal.Body[0].(string)
						newOwner := signal.Body[2].(string)

						if strings.HasPrefix(serviceName, MPRISNamespace) {
							if newOwner == "" {
								fmt.Printf("Player disappeared: %s\n", serviceName)
								delete(m.Players, serviceName)
							} else {
								playerInfo, err := m.GetPlayerInfo(serviceName)
								if err == nil {
									fmt.Printf("Player appeared: %s\n", serviceName)
									m.Players[serviceName] = playerInfo
								}
								if playerInfo.Status.PlaybackStatus == "Playing" {
									m.ActivePlayer = signal.Sender
								}
							}
						}
					}

				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}
