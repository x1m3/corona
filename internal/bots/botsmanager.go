package bots

import (
	"log"
	"time"

	"github.com/x1m3/corona/internal/corona"
)

type Manager struct {
	game *corona.Game
}

func NewManager(g *corona.Game) *Manager {
	return &Manager{game: g}
}

func (m *Manager) Init() {
	t := time.NewTicker(5 * time.Second)
	for {
		<-t.C
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
