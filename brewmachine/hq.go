package main

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/hybridgroup/gobot"
	"github.com/hybridgroup/gobot/api"
	"github.com/hybridgroup/gobot/platforms/mqtt"
	"github.com/hybridgroup/gobot/platforms/pebble"
)

func pebbleRobot() *gobot.Robot {
	p := pebble.NewPebbleAdaptor("pebble")
	m := mqtt.NewMqttAdaptor("mqtt", "tcp://192.168.0.90:1883", "pebble")
	watch := pebble.NewPebbleDriver(p, "pebble")

	work := func() {
		m.On("watch", func(data []byte) {
			watch.SendNotification(string(data))
		})
	}

	return gobot.NewRobot("pebble",
		[]gobot.Connection{p, m},
		[]gobot.Device{watch},
		work,
	)
}

func hqRobot() *gobot.Robot {
	m := mqtt.NewMqttAdaptor("mqtt", "tcp://192.168.0.90:1883", "hq")

	work := func() {
		watch := func(msg string) {
			fmt.Println(msg)
			m.Publish("watch", []byte(msg))
		}

		post := func(val []byte) {
			v, _ := url.ParseQuery(string(val))
			_, err := http.PostForm("https://brewmachine.herokuapp.com/drinks",
				v,
			)
			if err != nil {
				fmt.Println(err)
			}
		}

		m.On("pump", func(data []byte) {
			post(data)
			v, _ := url.ParseQuery(string(data))
			watch(fmt.Sprintf("Customers served: %v", v.Get("drink_id")))
		})
		m.On("fault", func(data []byte) {
			post(data)
			watch("There was a fault!")
		})
		m.On("drone", func(data []byte) {
			post(data)
			watch(fmt.Sprintf("Message from drone: %v", string(data)))
		})
		m.On("gcs", func(data []byte) {
			fmt.Println(string(data))
			post(data)
		})
	}

	return gobot.NewRobot("mqtt",
		[]gobot.Connection{m},
		work,
	)
}

func main() {
	gbot := gobot.NewGobot()
	api.NewAPI(gbot).Start()

	gbot.AddRobot(hqRobot())
	gbot.AddRobot(pebbleRobot())

	gbot.Start()
}
