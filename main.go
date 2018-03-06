package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/labstack/gommon/color"
	"github.com/pkg/profile"
	uuid "github.com/satori/go.uuid"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/clagraff/pobox/endpoints"
	"github.com/clagraff/pobox/monitoring"
	"github.com/clagraff/pobox/requests"
)

var (
	app    = kingpin.New("pobox", "An app for logging requests from web-hooks")
	routes = app.Arg("routes-files", "JSON/Yaml file path which defines custom routes").String()
	prof   = app.Flag("prof", "Enable profiling").Hidden().Bool()
)

func init() {
	app.Version("0.0.0")
	app.Parse(os.Args[1:])
}

func logRequests(receivedRequests chan http.Request) {
	for {
		select {
		case r := <-receivedRequests:
			color.Println(
				color.Magenta("UTC Time:"),
				"\t",
				color.Yellow(time.Now().UTC().String()),
			)

			color.Println(
				color.Magenta("Local Time:"),
				"\t",
				color.Yellow(time.Now().Local().String()),
			)

			color.Println(
				color.Magenta("Method:"),
				"\t",
				color.Yellow(r.Method),
			)

			color.Println(
				color.Magenta("URL:"),
				"\t\t",
				color.Yellow(r.URL.String()),
			)

			color.Println(color.Magenta("Headers"))

			for key, value := range r.Header {
				color.Println(
					"\t",
					color.Yellow(fmt.Sprintf("%s: ", key)),
					color.Green(strings.TrimSpace(strings.Join(value, ""))),
				)
			}

			buf, err := ioutil.ReadAll(r.Body)
			if err != nil {
				panic(err)
			}

			if len(buf) > 0 {
				color.Println(color.Magenta("Body:"))
				color.Println("\t", color.Blue(string(buf)))
			}

			fmt.Println("\n")
		default:
		}
	}
}

func main() {
	if prof != nil && *prof {
		defer profile.Start().Stop()
	}

	endpointsPort := 8080
	monitoringPort := 8090
	apiUUID := uuid.Must(uuid.NewV4())

	color.Println(
		color.Green(
			fmt.Sprintf("POBox web-hook server running on port: %d", endpointsPort),
		),
	)
	color.Println(
		color.Blue(
			fmt.Sprintf("POBox API server running on port: %d with API Key: %s", monitoringPort, apiUUID),
		),
	)
	fmt.Println("")

	rr := make(chan requests.Request)
	_, startEndpointsServer := endpoints.CreateServer(rr, endpointsPort)
	_, startMonitoringServer := monitoring.CreateServer(apiUUID, rr, monitoringPort)

	go func() { startEndpointsServer() }()
	go func() { startMonitoringServer() }()

	select {} // block as the servers run. less CPU intensive than for-loop.
}
