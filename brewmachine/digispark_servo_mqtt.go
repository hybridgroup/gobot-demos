package main

import (
	"github.com/hybridgroup/gobot"
	"github.com/hybridgroup/gobot/platforms/digispark"
	"github.com/hybridgroup/gobot/platforms/gpio"
	"github.com/hybridgroup/gobot/platforms/mqtt"
)

func main() {
	gbot := gobot.NewGobot()

	m := mqtt.NewMqttAdaptor("mqtt", "tcp://192.168.0.90:1883", "drone")
	digisparkAdaptor := digispark.NewDigisparkAdaptor("digispark")

	servo := gpio.NewServoDriver(digisparkAdaptor, "servo", "0")

	work := func() {
		servo.Move(10)
		m.On("drop", func(data []byte) {
			servo.Move(150)
			m.Publish("drone", []byte("Dropped"))
		})
	}

	robot := gobot.NewRobot("servoBot",
		[]gobot.Connection{digisparkAdaptor, m},
		[]gobot.Device{servo},
		work,
	)

	gbot.AddRobot(robot)

	gbot.Start()
}
