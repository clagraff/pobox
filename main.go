package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/labstack/gommon/color"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
	yaml "gopkg.in/yaml.v2"

	"github.com/clagraff/pobox/endpoints"
)

var (
	app    = kingpin.New("pobox", "An app for logging requests from web-hooks")
	routes = app.Arg("routes-files", "JSON/Yaml file path which defines custom routes").String()
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

func parseRoutesFile() []endpoints.DefinedRoute {
	definedRoutes := make([]endpoints.DefinedRoute, 0)

	if routes != nil && len(*routes) > 0 {
		routePath := *routes
		ext := filepath.Ext(routePath)

		buf, err := ioutil.ReadFile(routePath)
		if err != nil {
			panic(err)
		}

		if ext == ".yaml" || ext == ".yml" {
			err := yaml.Unmarshal(buf, &definedRoutes)
			if err != nil {
				panic(err)
			}
		} else if ext == ".json" {
			err := json.Unmarshal(buf, &definedRoutes)
			if err != nil {
				panic(err)
			}
		} else {
			panic(fmt.Errorf("unsupported route file extension: %s", ext))
		}
	}

	return definedRoutes
}

func main() {
	rr := make(chan http.Request)

	routes := parseRoutesFile()
	_, start := endpoints.CreateServer(routes, rr, 8080)
	go func() { start() }()

	logRequests(rr)
}
