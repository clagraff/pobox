package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/nbari/violetear"
)

type ParsedRequest struct {
	Method  string
	Headers map[string]string
	Body    []byte
	URL     string
}

func MakeParsedRequest(req http.Request) ParsedRequest {
	r := ParsedRequest{
		Method:  req.Method,
		URL:     req.URL.String(),
		Headers: make(map[string]string),
	}

	for key, value := range req.Header {
		r.Headers[key] = strings.Join(value, "")
	}

	bytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}
	r.Body = bytes

	return r
}

var incomingRequests chan ParsedRequest = make(chan ParsedRequest)

func catchAll(w http.ResponseWriter, r *http.Request) {
	incomingRequests <- MakeParsedRequest(*r)
	w.Write([]byte(""))
}

func runRequestServer() {
	router := violetear.New()
	router.LogRequests = false
	router.Verbose = false

	router.HandleFunc("*", catchAll)

	srv := &http.Server{
		Addr:           ":8080",
		Handler:        router,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   7 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	srv.ListenAndServe()
}

func logRequests() {
	for {
		select {
		case r := <-incomingRequests:
			fmt.Printf("Method: %s\n", r.Method)
			fmt.Printf("URL: %s\n", r.URL)
			fmt.Println("Headers:")
			for key, value := range r.Headers {
				fmt.Printf("\t%s: %s\n", key, value)
			}
			fmt.Println("Body:")
			fmt.Printf("\t%s\n", r.Body)
		default:
		}
	}
}

func main() {
	go func() { runRequestServer() }()
	logRequests()
}
