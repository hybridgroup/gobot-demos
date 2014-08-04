package main

import (
	"net/http"
	"strings"

	"github.com/hybridgroup/gobot"
	"github.com/hybridgroup/gobot/api"
	"github.com/hybridgroup/gobot/platforms/neurosky"
)

func main() {
	gbot := gobot.NewGobot()
	a := api.NewAPI(gbot)

	a.Get("/brain/:a", func(res http.ResponseWriter, req *http.Request) {
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
	adaptor := neurosky.NewNeuroskyAdaptor("neurosky", "/dev/rfcomm0")
	neuro := neurosky.NewNeuroskyDriver(adaptor, "neurosky")

	gbot.AddRobot(gobot.NewRobot("brain",
		[]gobot.Connection{adaptor},
		[]gobot.Device{neuro},
	))

	a.Start()
	gbot.Start()
}
