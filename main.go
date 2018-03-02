package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/clagraff/pobox/endpoints"
)

func logRequests(receivedRequests chan http.Request) {
	for {
		select {
		case r := <-receivedRequests:
			fmt.Printf("UTC Time:\t%s\n", time.Now().UTC())
			fmt.Printf("Local Time:\t%s\n", time.Now().Local())
			fmt.Printf("Method:\t%s\n", r.Method)
			fmt.Printf("URL:\t%s\n", r.URL)
			fmt.Println("Headers:")
			for key, value := range r.Header {
				fmt.Printf("\t%s: %s\n", key, strings.Join(value, ""))
			}
			fmt.Println("Body:")

			buf, err := ioutil.ReadAll(r.Body)
			if err != nil {
				panic(err)
			}
			fmt.Printf("\t%s\n", string(buf))
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
