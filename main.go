package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/gommon/color"

	"github.com/clagraff/pobox/endpoints"
)

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
	rr := make(chan http.Request)
	_, start := endpoints.CreateServer(rr, 8080)
	go func() { start() }()

	logRequests(rr)
}
