package mpris

import (
	"fmt"
	"time"

	"github.com/godbus/dbus/v5"
)

type TrackMetadata struct {
	TrackID string        `json:"trackid"` // mpris:trackid
	Length  time.Duration `json:"length"`  // mpris:length (microseconds)
	ArtURL  string        `json:"artUrl"`  // mpris:artUrl
	Album   string        `json:"album"`   // xesam:album
	Artist  []string      `json:"artist"`  // xesam:artist (can be multiple)
	Title   string        `json:"title"`   // xesam:title
	URL     string        `json:"url"`     // xesam:url
}

type PlayerStatus struct {
	PlaybackStatus string        // "Playing", "Paused", "Stopped"
	Position       time.Duration // Current position
	Rate           float64       // Playback rate
	Volume         float64       // Volume level
	CanControl     bool          // Whether player accepts control
	CanPlay        bool
	CanPause       bool
	CanGoNext      bool
	CanGoPrevious  bool
}

type PlayerInfo struct {
	ServiceName  string // org.mpris.MediaPlayer2.spotify
	Identity     string // Player name (e.g., "Spotify")
	DesktopEntry string // Desktop file name
	Metadata     *TrackMetadata
	Status       *PlayerStatus
}

func (m *MPRISClient) GetPlayerInfo(serviceName string) (*PlayerInfo, error) {
	variant := m.GetPlayerProperties(serviceName)
	metaData, err := m.ParseMetadata(variant)
	if err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %s", err)
	}
	playerStatus, err := m.ParsePlayerStatus(variant)
	if err != nil {
		return nil, fmt.Errorf("failed to parse player status: %s", err)
	}
	baseInfo, err := m.GetBaseProperties(serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get base info: %s", err)
	}

	return &PlayerInfo{
		ServiceName:  serviceName,
		Identity:     baseInfo["Identity"].(string),
		DesktopEntry: baseInfo["DesktopEntry"].(string),
		Metadata:     metaData,
		Status:       playerStatus,
	}, nil
}

func (m *MPRISClient) ParsePlayerStatus(variants map[string]dbus.Variant) (*PlayerStatus, error) {
	position := time.Duration(variants["Position"].Value().(int64)) * time.Microsecond
	playerStatus := &PlayerStatus{
		PlaybackStatus: extractStringProperty(variants, "PlaybackStatus"),
		Position:       position,
		Rate:           variants["Rate"].Value().(float64),
		Volume:         variants["Volume"].Value().(float64),
		CanControl:     variants["CanControl"].Value().(bool),
		CanPlay:        variants["CanPlay"].Value().(bool),
		CanPause:       variants["CanPause"].Value().(bool),
		CanGoNext:      variants["CanGoNext"].Value().(bool),
		CanGoPrevious:  variants["CanGoPrevious"].Value().(bool),
	}

	// fmt.Printf("Player Status:\n%+v\n", playerStatus) // Debug
	return playerStatus, nil
}

func (m *MPRISClient) ParseMetadata(variants map[string]dbus.Variant) (*TrackMetadata, error) {
	var trackMetadata *TrackMetadata
	var metaData map[string]dbus.Variant
	err := variants["Metadata"].Store(&metaData)
	if err != nil {
		return nil, fmt.Errorf("error storing metadata: %s", err)
	}
	var artists []string
	if artistVariant, exists := metaData["xesam:artist"]; exists {
		err := artistVariant.Store(&artists)
		if err != nil {
			return nil, fmt.Errorf("error parsing artists: %s", err)
		}
	}

	var length time.Duration
	if lengthVariant, exists := metaData["mpris:length"]; exists {
		if lengthValue := lengthVariant.Value(); lengthValue != nil {
			length = time.Duration(metaData["mpris:length"].Value().(uint64)) * time.Microsecond
		}
	}

	trackMetadata = &TrackMetadata{
		TrackID: extractStringProperty(metaData, "mpris:trackid"),
		Length:  length,
		ArtURL:  extractStringProperty(metaData, "mpris:artUrl"),
		Album:   extractStringProperty(metaData, "xesam:album"),
		Artist:  artists,
		Title:   extractStringProperty(metaData, "xesam:title"),
		URL:     extractStringProperty(metaData, "xesam:url"),
	}
	// fmt.Printf("Track metadata:\n%+v\n", trackMetadata) // Debug
	return trackMetadata, nil
}

func (m *MPRISClient) GetPlayerProperties(serviceName string) map[string]dbus.Variant {
	call := m.Conn.Object(serviceName, MPRISObjectPath).Call(DbusPropertiesGetAll, 0, PlayerInterface)

	var props map[string]dbus.Variant
	err := call.Store(&props)
	if err != nil {
		fmt.Printf("failed to get properties: %s", err)
		return nil
	}

	// fmt.Println("PROPS:")
	// for key, value := range props {
	// 	fmt.Printf("%s: %v\n", key, value)
	// }
	// fmt.Println("===***===")

	return props
}

func (m *MPRISClient) GetBaseProperties(serviceName string) (map[string]any, error) {
	call := m.Conn.Object(serviceName, MPRISObjectPath).Call(DbusPropertiesGetAll, 0, MPRISInterface)

	var props map[string]any
	err := call.Store(&props)
	if err != nil {
		return nil, fmt.Errorf("failed to get properties: %s", err)
	}
	// fmt.Printf("%v", props) // Debug

	return props, nil
}

func extractStringProperty(props map[string]dbus.Variant, key string) string {
	if variant, exists := props[key]; exists {
		return variant.String()
	}
	return ""
}
