// Package waybar
package waybar

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cheezecakee/mpris-daemon/internal/mpris"
)

type WaybarOutput struct {
	Text    string `json:"text"`
	Tooltip string `json:"tooltip"`
	Class   string `json:"class"`
	Alt     string `json:"alt,omitempty"`
}

func FormatForWaybar(player mpris.PlayerInfo) WaybarOutput {
	artistString := strings.Join(player.Metadata.Artist, ", ")
	tooltip := fmt.Sprintf("%s\nby %s\nfrom %s", player.Metadata.Title, artistString, player.Metadata.Album)
	text := fmt.Sprintf("%s - %s", artistString, player.Metadata.Title)
	class := getWaybarClass(player.Status.PlaybackStatus)

	waybarOut := WaybarOutput{
		Text:    text,
		Tooltip: tooltip,
		Class:   class,
	}

	return waybarOut
}

func (w WaybarOutput) ToJSON() (string, error) {
	response, err := json.Marshal(w)
	if err != nil {
		return "", err
	}
	return string(response), nil
}

func FormatTooltip(metadata mpris.TrackMetadata) string {
	artistString := strings.Join(metadata.Artist, ",")
	return fmt.Sprintf("%s\nby %s\nfrom %s", metadata.Title, artistString, metadata.Album)
}

func getWaybarClass(status string) string {
	switch strings.ToLower(status) {
	case "playing":
		return "playing"
	case "paused":
		return "paused"
	case "stopped":
		return "stopped"
	default:
		return "unknown"
	}
}
