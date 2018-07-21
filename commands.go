package main

import (
	"fmt"
	"log"
	"strings"

	tb "gopkg.in/tucnak/telebot.v2"
)

func MeteoFreyaCmd(bot *tb.Bot, m *tb.Message) {
	sensor := config.Defaults.Sensor

	if len(m.Payload) > 0 {
		sensor = strings.Fields(m.Payload)[0]
	}

	d, err := QueryLastMeteoFreyaData(sensor)

	if err != nil {
		log.Printf("[MeteoFreya Error]: %s", err)
		bot.Send(m.Chat, "Error querying InfluxDB, check logs.")
		return
	}

	if d.DHT.Temperature == 0 && d.DHT.Humidity == 0 && d.DHT.HeatIndex == 0 {
		bot.Send(m.Chat, "No recent data for that sensor.")
		return
	}

	message := fmt.Sprintf("_MeteoFreya:_ %s\n%v *ºC* | %v *%%RH* | %v *ºC HI* | %v *hPa*",
		sensor, d.DHT.Temperature, d.DHT.Humidity, d.DHT.HeatIndex, d.BMP180.Pressure)

	bot.Send(m.Chat, message, tb.ModeMarkdown)
}
