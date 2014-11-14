package main

import (
	"fmt"
	"net/url"
	"time"

	"github.com/hybridgroup/gobot"
	"github.com/hybridgroup/gobot/platforms/gpio"
	"github.com/hybridgroup/gobot/platforms/intel-iot/edison"
	"github.com/hybridgroup/gobot/platforms/mqtt"
)

func main() {
	gbot := gobot.NewGobot()

	e := edison.NewEdisonAdaptor("edison")
	m := mqtt.NewMqttAdaptor("mqtt", "tcp://192.168.0.90:1883", "pump")

	lever := gpio.NewButtonDriver(e, "lever", "2")
	fault := gpio.NewButtonDriver(e, "fault", "4")
	pump := gpio.NewDirectPinDriver(e, "pump", "13")

	work := func() {
		dgram := url.Values{
			"name":         {"Four"},
			"dispenser_id": {"4"},
			"drink_id":     {"0"},
			"event":        {"online"},
			"details":      {"dispenser"},
		}
		pumping := false
		served := byte(0)

		m.On("startPump", func(data []byte) {
			if !pumping {
				pumping = true
				pump.DigitalWrite(1)
				served++
				dgram.Set("event", "online")
				dgram.Set("drink_id", fmt.Sprintf("%v", served))
				m.Publish("pumped", []byte(dgram.Encode()))
				gobot.After(2*time.Second, func() {
					pump.DigitalWrite(0)
					pumping = false
				})
			}
		})

		gobot.On(lever.Event("push"), func(data interface{}) {
			m.Publish("pump", []byte{})
		})

		m.On("startFault", func(data []byte) {
			dgram.Set("event", "error")
			m.Publish("fault", []byte(dgram.Encode()))
		})

		gobot.On(fault.Event("push"), func(data interface{}) {
			m.Publish("startFault", []byte{})
		})
	}

	gbot.AddRobot(gobot.NewRobot("brewmachine",
		[]gobot.Connection{e, m},
		[]gobot.Device{lever, fault, pump},
		work,
	))

	gbot.Start()
}
