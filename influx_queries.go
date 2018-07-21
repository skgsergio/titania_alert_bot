package main

import (
	"encoding/json"
	"fmt"

	"github.com/influxdata/influxdb/client/v2"
)

type MeteoFreya struct {
	DHT struct {
		Temperature float64
		Humidity    float64
		HeatIndex   float64
	}

	BMP180 struct {
		Pressure    float64
		Temperature float64
	}
}

func QueryDB(db string, cmd string) ([]client.Result, error) {
	host := fmt.Sprintf("%s://%s:%d", config.InfluxDB.Proto, config.InfluxDB.Host, config.InfluxDB.Port)

	con, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     host,
		Username: config.InfluxDB.Username,
		Password: config.InfluxDB.Password,
	})

	res, err := con.Query(client.Query{
		Command:  cmd,
		Database: db,
	})

	con.Close()

	if err != nil {
		return nil, err
	}

	if res.Error() != nil {
		return nil, res.Error()
	}

	return res.Results, nil
}

func QueryLastMeteoFreyaData(sensor string) (MeteoFreya, error) {
	data := MeteoFreya{}

	query := fmt.Sprintf("SELECT last(temperature) as t, last(humidity) as h, last(heatindex) as hi FROM dht WHERE time >= now() - 5m AND sensor='%s';", sensor)
	query += fmt.Sprintf("SELECT last(pressue) as p, last(temperature) as t FROM bmp180 WHERE time >= now() - 5m AND sensor='%s';", sensor)

	res, err := QueryDB("meteofreya", query)

	if err != nil {
		return data, err
	} else if len(res) != 2 {
		return data, fmt.Errorf("Expected 2 series, retrieved %d", len(res))
	} else {
		for _, r := range res {
			if len(r.Series) != 1 {
				return data, nil
			}

			switch r.Series[0].Name {
			case "dht":
				data.DHT.Temperature, _ = r.Series[0].Values[0][1].(json.Number).Float64()
				data.DHT.Humidity, _ = r.Series[0].Values[0][2].(json.Number).Float64()
				data.DHT.HeatIndex, _ = r.Series[0].Values[0][3].(json.Number).Float64()

			case "bmp180":
				data.BMP180.Pressure, _ = r.Series[0].Values[0][1].(json.Number).Float64()
				data.BMP180.Temperature, _ = r.Series[0].Values[0][2].(json.Number).Float64()
			}
		}
	}

	return data, nil
}
