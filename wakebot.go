package main

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/starudream/wake-on-lan/wol"
	"golang.zx2c4.com/irc/hbot"
)

type Config struct {
	Channels []string `json:"channels"`
	Nick     string   `json:"nick"`
	Host     string   `json:"host"`
}

func getConfig(fileName string) Config {
	fileStream, err := os.Open(fileName)
	if err != nil {
		log.Fatal("Can't open file!")
	}
	bytes, err := io.ReadAll(fileStream)
	if err != nil {
		log.Fatal("Error reading file")
	}
	var config Config

	err = json.Unmarshal(bytes, &config)
	if err != nil {
		log.Fatal("Error parsing file")
	}

	return config
}

func main() {
	config := getConfig("go-wakebot.json")

	botConfig := hbot.Config{
		Host:     config.Host,
		Nick:     config.Nick,
		Realname: "Mr. Wakeman",
		User:     "wakeman",
		Channels: config.Channels,
		Logger:   hbot.Logger{Verbosef: log.Printf, Errorf: log.Printf},
	}

	bot := hbot.NewBot(&botConfig)

	bot.AddTrigger(hbot.Trigger{
		Condition: func(b *hbot.Bot, m *hbot.Message) bool {
			if m.Command != "PRIVMSG" || strings.HasPrefix(m.Prefix.User, hbot.CommonBotUserPrefix) || strings.HasPrefix(m.Prefix.User, "~"+hbot.CommonBotUserPrefix) {
				return false
			}
			return true
		},
		Action: func(b *hbot.Bot, m *hbot.Message) {
			message := ""
			target := m.Prefix.Name
			if strings.Contains(m.Param(0), "#") {
				target = m.Param(0)
				if target[0] == '@' || target[0] == '+' {
					target = target[1:]
				}
				if len(m.Params) == 2 {
					cmdArr := strings.SplitN(m.Param(1), " ", 2)
					if len(cmdArr) == 2 {
						cmd := cmdArr[0]
						param := cmdArr[1]

						if cmd == "!wol" {
							_, err := net.ParseMAC(param)
							if err != nil {
								message = "That's an invalid MAC-address I'm afraid."
							} else {
								wol.Send("255.255.255.255", "9", param)
								message = "Right away, sir!"
							}
						}
					}
				}
				//message = m.Prefix.Name + ": " + message
			}
			if len(message) > 0 {
				b.Msg(target, message)
			}
		},
	})

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for range c {
			bot.Close()
			os.Exit(0)
		}
	}()

	for {
		bot.Run()
		time.Sleep(time.Second * 5)
	}
}
