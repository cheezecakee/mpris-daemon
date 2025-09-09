package main

import (
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
		client.GetPlayerInfo(player)
	}
}
