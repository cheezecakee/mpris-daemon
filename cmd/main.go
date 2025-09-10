package main

import (
	"context"
	"fmt"

	"github.com/cheezecakee/mpris-daemon/internal/mpris"
	"github.com/cheezecakee/mpris-daemon/internal/waybar"
)

func main() {
	client, err := mpris.NewMPRISClient()
	if err != nil {
		fmt.Printf("failed to setup mpris client: %s", err)
	}
	defer client.Conn.Close()

	updates := make(chan mpris.PlayerInfo, 10)
	ctx := context.Background()
	client.StartListening(ctx, updates)

	go func() {
		for playerInfo := range updates {
			// fmt.Printf("Update: %s - %s\n", playerInfo.Metadata.Artist, playerInfo.Metadata.Title) // Debug

			output := waybar.FormatForWaybar(playerInfo)
			json, _ := output.ToJSON()
			fmt.Println(json)
		}
	}()

	// Keep program running
	select {} // or time.Sleep(time.Hour) for testing
}
