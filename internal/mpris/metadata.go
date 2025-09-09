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
	Metadata     TrackMetadata
	Status       PlayerStatus
}

/**
[org.mpris.MediaPlayer2.spotify]
Position: @x 96666000
CanGoNext: true
Metadata: {"mpris:artUrl": <"https://i.scdn.co/image/ab67616d0000b2734dcb6c5df15cf74596ab25a4">, "mpris:length": <@t 194607000>, "mpris:trackid": <"/com/spotify/track/1CPZ5BxNNd0n0nF4Orb9JS">, "xesam:album": <"KPop Demon Hunters (Soundtrack from the Netflix Film)">, "xesam:albumArtist": <["KPop Demon Hunters Cast"]>, "xesam:artist": <["HUNTR/X"]>, "xesam:autoRating": <@d 1>, "xesam:discNumber": <1>, "xesam:title": <"Golden">, "xesam:trackNumber": <4>, "xesam:url": <"https://open.spotify.com/track/1CPZ5BxNNd0n0nF4Orb9JS">}
Volume: @d 1
MaximumRate: @d 1
PlaybackStatus: "Playing"
CanControl: true
Shuffle: false
MinimumRate: @d 1
CanPause: true
CanPlay: true
Rate: @d 1
CanSeek: true
CanGoPrevious: true
LoopStatus: "Track"
**/

func extractStringProperty(props map[string]dbus.Variant, key string) string {
	if variant, exists := props[key]; exists {
		return variant.String()
	}
	return ""
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

	length := time.Duration(metaData["mpris:length"].Value().(uint64)) * time.Microsecond

	trackMetadata = &TrackMetadata{
		TrackID: extractStringProperty(metaData, "mpris:trackid"),
		Length:  length,
		ArtURL:  extractStringProperty(metaData, "mpris:artUrl"),
		Album:   extractStringProperty(metaData, "xesam:album"),
		Artist:  artists,
		Title:   extractStringProperty(metaData, "xesam:title"),
		URL:     extractStringProperty(metaData, "xesam:url"),
	}
	fmt.Printf("Track metadata:\n%+v\n", trackMetadata)
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

	fmt.Printf("Player Status:\n%+v\n", playerStatus)
	return playerStatus, nil
}
