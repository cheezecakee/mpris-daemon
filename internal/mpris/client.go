// Package mpris
package mpris

import (
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

	fmt.Println("DBUS connect: ", newConn.Connected())

	return &MPRISClient{
		Conn: newConn,
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

	fmt.Println(mprisPlayers)
	return mprisPlayers, nil
}
