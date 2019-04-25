package bots

import (
	"github.com/x1m3/elixir/games/cookies"
	"log"
	"time"
)

type BotsManager struct {
	game *cookies.Game
}
func NewBotsManager(g *cookies.Game) *BotsManager {
	return &BotsManager{game:g}
}

func (m *BotsManager) Init() {
	t := time.NewTicker(5 * time.Second)
	for {
		<- t.C
		go func() {
			bot := New(m.game, NewDummyBotAgent(200, 200))
			log.Println("Bot started")
			if err := bot.Run(); err != nil {
				log.Println(err)
				bot.Destroy()
				return
			}
		}()
	}
}
