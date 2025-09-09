package main

import (
	"context"
	"fmt"
	"log"

	"github.com/cheezecakee/mpris-daemon/internal/mpris"
)

func main() {
	client, err := mpris.NewMPRISClient()
	if err != nil {
		log.Printf("failed to setup mpris client: %s", err)
	}
	defer client.Conn.Close()

	players, err := client.ListPlayers()
	if err != nil {
		log.Printf("failed to get players: %s", err)
	}

	for _, player := range players {
		playerInfo, err := client.GetPlayerInfo(player)
		if err != nil {
			fmt.Printf("error getting player info: %s", err)
		}
		client.Players[player] = playerInfo
	}

	updates := make(chan mpris.PlayerInfo, 10)
	ctx := context.Background()
	client.StartListening(ctx, updates)

	go func() {
		for playerInfo := range updates {
			fmt.Printf("Update: %s - %s\n", playerInfo.Metadata.Artist, playerInfo.Metadata.Title)
		}
	}()

	// Keep program running
	select {} // or time.Sleep(time.Hour) for testing
}
