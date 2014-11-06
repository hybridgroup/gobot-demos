package main

import (
	"fmt"
	"time"

	"github.com/hybridgroup/gobot"
	"github.com/hybridgroup/gobot/api"
	"github.com/hybridgroup/gobot/platforms/gpio"
	"github.com/hybridgroup/gobot/platforms/intel-iot/edison"
	"github.com/hybridgroup/gobot/platforms/mqtt"
	"github.com/hybridgroup/gobot/platforms/pebble"
)

func main() {
	gbot := gobot.NewGobot()
	server := api.NewAPI(gbot)
	server.Port = "8080"
	server.Start()

	e := edison.NewEdisonAdaptor("edison")
	p := pebble.NewPebbleAdaptor("pebble")
	m := mqtt.NewMqttAdaptor("mqtt", "tcp://192.168.0.90:1883")

	lever := gpio.NewButtonDriver(e, "lever", "2")
	fault := gpio.NewButtonDriver(e, "fault", "4")
	pump := gpio.NewDirectPinDriver(e, "pump", "13")

	watch := pebble.NewPebbleDriver(p, "pebble")

	work := func() {
		pumping := false
		served := byte(0)
		gobot.On(lever.Event("push"), func(data interface{}) {
			if !pumping {
				pumping = true
				pump.DigitalWrite(1)
				served++
				m.Publish("pump", []byte{served})
				gobot.After(2*time.Second, func() {
					pump.DigitalWrite(0)
					pumping = false
				})
			}
		})

		gobot.On(fault.Event("push"), func(data interface{}) {
			m.Publish("fault", []byte{})
		})

		m.On("pump", func(data interface{}) {
			msg := fmt.Sprintf("Customers served: %v", data.([]byte)[0])
			fmt.Println(msg)
			watch.SendNotification(msg)
		})

		m.On("fault", func(data interface{}) {
			msg := "There was a fault!"
			fmt.Println(msg)
			watch.SendNotification(msg)
		})

		m.On("drone", func(data interface{}) {
			msg := fmt.Sprintf("Message from drone: %v", string(data.([]byte)))
			fmt.Println(msg)
			watch.SendNotification(msg)
		})
	}

	gbot.AddRobot(gobot.NewRobot("brewmachine",
		[]gobot.Connection{e, m, p},
		[]gobot.Device{lever, fault, pump, watch},
		work,
	))

	gbot.Start()
}
