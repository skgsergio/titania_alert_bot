package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

type Configuration struct {
	Token string `json:"token"`

	AuthorizedUsers []int `json:"authorized_user_ids"`

	InfluxDB struct {
		Proto    string `json:"proto"`
		Host     string `json:"host"`
		Port     uint16 `json:"port"`
		Username string `json:"user"`
		Password string `json:"password"`
	} `json:"influxdb"`

	Defaults struct {
		Sensor string `json:"sensor"`
		Host   string `json:"host"`
	} `json:"defaults"`
}

var config = Configuration{}

func LoadConfig(file string) error {
	cfgFile, err := os.Open(file)
	if err != nil {
		return err
	}

	defer cfgFile.Close()

	decoder := json.NewDecoder(cfgFile)
	if decoder.Decode(&config) != nil {
		return err
	}

	return nil
}

func main() {
	// Read config file
	err := LoadConfig("config.json")
	if err != nil {
		log.Fatal(err)
	}

	// Initialize the bot with a Middleware poller that checks if the user is authorized
	bot, err := tb.NewBot(tb.Settings{
		Token: config.Token,
		Poller: tb.NewMiddlewarePoller(
			&tb.LongPoller{Timeout: 15 * time.Second},
			func(u *tb.Update) bool {
				if u.Message != nil {
					log.Printf("%+v: %s", u.Message.Sender, u.Message.Text)

					for _, userId := range config.AuthorizedUsers {
						if u.Message.Sender.ID == userId {
							return true
						}
					}
					return false
				}
				return true
			},
		),
	})

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Logged in as %s <@%s>", bot.Me.FirstName, bot.Me.Username)

	/* BEGIN HANDLERS */

	var helpMsg []string
	bot.Handle("/help", func(m *tb.Message) {
		bot.Send(m.Sender, strings.Join(helpMsg, "\n"))
	})

	helpMsg = append(helpMsg, fmt.Sprintf("/meteo [sensor] - Retrieve meteo sensor info [MeteoFreya] (default: %s)", config.Defaults.Sensor))
	bot.Handle("/meteo", func(m *tb.Message) { MeteoFreyaCmd(bot, m) })

	/*
		helpMsg = append(helpMsg, fmt.Sprintf("/load [host] - Retrieve host load: 1m 5m 15m [telegraf] (default: %s)", config.Defaults.Host))
		//bot.Handle("/load", func(m *tb.Message) { HostLoadCmd(bot, m) })

		helpMsg = append(helpMsg, fmt.Sprintf("/net [host] - Retrieve last 5 min network stats [telegraf] (default: %s)", config.Defaults.Host))
		//bot.Handle("/net", func(m *tb.Message) { HostNetCmd(bot, m) })
	*/
	bot.Start()
}
