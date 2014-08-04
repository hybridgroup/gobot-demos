package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/hybridgroup/gobot"
	"github.com/hybridgroup/gobot/api"
	"github.com/hybridgroup/gobot/platforms/mavlink"
	common "github.com/hybridgroup/gobot/platforms/mavlink/common"
)

type telemetry struct {
	Status    string
	Roll      float32
	Yaw       float32
	Pitch     float32
	Latitude  float32
	Longitude float32
	Altitude  int32
	Heading   float32
}

func main() {
	if gobot.Version() != "0.7.dev" {
		panic("this requires the dev branch!")
	}

	summary := &telemetry{Status: "STANDBY"}
	gbot := gobot.NewGobot()

	a := api.NewAPI(gbot)

	a.Get("/mavlink/:a", func(res http.ResponseWriter, req *http.Request) {
		path := req.URL.Path
		t := strings.Split(path, "/")
		buf, err := Asset("assets/" + t[2])
		if err != nil {
			http.Error(res, err.Error(), http.StatusNotFound)
			return
		}
		t = strings.Split(path, ".")
		if t[len(t)-1] == "js" {
			res.Header().Set("Content-Type", "text/javascript; charset=utf-8")
		} else if t[len(t)-1] == "css" {
			res.Header().Set("Content-Type", "text/css; charset=utf-8")
		}
		res.Write(buf)
	})

	adaptor := mavlink.NewMavlinkAdaptor("iris", "/dev/ttyACM0")
	iris := mavlink.NewMavlinkDriver(adaptor, "iris")

	work := func() {

		iris.AddEvent("telemetry")

		gobot.Once(iris.Event("packet"), func(data interface{}) {
			fmt.Println(data)
			packet := data.(*common.MAVLinkPacket)

			dataStream := common.NewRequestDataStream(10,
				packet.SystemID,
				packet.ComponentID,
				0,
				1,
			)
			iris.SendPacket(common.CraftMAVLinkPacket(packet.SystemID,
				packet.ComponentID,
				dataStream,
			))
		})

		gobot.On(iris.Event("message"), func(data interface{}) {

			fmt.Println("message: ", data.(common.MAVLinkMessage).Id())
			if data.(common.MAVLinkMessage).Id() == 0 {
				statusCodes := map[uint8]string{
					1: "BOOT",
					2: "CALIBRATING",
					3: "STANDBY",
					4: "ACTIVE",
					5: "CRITICAL",
					6: "EMERGENCY",
					7: "POWEROFF",
					8: "ENUM_END",
				}
				summary.Status = statusCodes[data.(*common.Heartbeat).SYSTEM_STATUS]
				if summary.Status != "" {
					gobot.Publish(iris.Event("telemetry"), summary)
				}
			}

			if data.(common.MAVLinkMessage).Id() == 30 {
				roll := data.(*common.Attitude).ROLL
				pitch := data.(*common.Attitude).PITCH
				yaw := data.(*common.Attitude).YAW

				if roll < 4 || roll > -4 {
					summary.Roll = (roll * 180 / 3.14)
				}
				if yaw < 4 || roll > -4 {
					summary.Yaw = (yaw * 180 / 3.14)
				}
				if pitch < 4 || roll > -4 {
					summary.Pitch = (pitch * 180 / 3.14)
				}
				gobot.Publish(iris.Event("telemetry"), summary)
			}

			if data.(common.MAVLinkMessage).Id() == 33 {
				summary.Latitude = float32(data.(*common.GlobalPositionInt).LAT) / 10000000.0
				summary.Longitude = float32(data.(*common.GlobalPositionInt).LON) / 10000000.0
				summary.Altitude = data.(*common.GlobalPositionInt).ALT
				summary.Heading = float32(data.(*common.GlobalPositionInt).HDG) / 100
				gobot.Publish(iris.Event("telemetry"), summary)
			}
		})
	}

	gbot.AddRobot(gobot.NewRobot("irisBot",
		[]gobot.Connection{adaptor},
		[]gobot.Device{iris},
		work,
	))

	a.Start()
	gbot.Start()
}
