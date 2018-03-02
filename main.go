package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/nbari/violetear"
	"github.com/olebedev/config"
)

var defaultConfigJSON string = `{
	"port": 8080,
	"verbose": false
}`

var help *bool = flag.Bool("h", false, "Show help text")
var configPath *string = flag.String("cfg", "", "Path to configuration JSON or Yaml.")
var appConfig *config.Config

func init() {
	flag.Parse()

	if help != nil && *help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	var err error
	appConfig, err = config.ParseJson(defaultConfigJSON)
	if err != nil {
		fmt.Println("failed to load default configuration settings")
		panic(err)
	}

	if configPath != nil && len(*configPath) > 0 {
		parts := strings.Split(*configPath, ".")
		ext := parts[len(parts)-1]

		if ext == "json" {
			appConfig, err = config.ParseJsonFile(*configPath)

		} else if ext == "yml" || ext == "yaml" {
			appConfig, err = config.ParseYamlFile(*configPath)
		} else {
			fmt.Println("invalid configuration file extension. must be JSON or Yaml")
			os.Exit(-1)
		}

		if err != nil {
			panic(err)
		}
	}
}

type ParsedRequest struct {
	Method    string
	Headers   map[string]string
	Body      []byte
	URL       string
	Timestamp time.Time
}

func MakeParsedRequest(req http.Request) ParsedRequest {
	r := ParsedRequest{
		Method:    req.Method,
		URL:       req.URL.String(),
		Headers:   make(map[string]string),
		Timestamp: time.Now().UTC(),
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
		Addr:           fmt.Sprintf(":%s", appConfig.UString("port", "8080")),
		Handler:        router,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   7 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if appConfig.UBool("verbose", false) {
		fmt.Printf("running server on %s\n", appConfig.UString("port", "8080"))
	}

	srv.ListenAndServe()
}

func logRequests() {
	for {
		select {
		case r := <-incomingRequests:
			fmt.Printf("UTC Timestamp:\t\t%s\n", r.Timestamp)
			fmt.Printf("Local Timestamp:\t%s\n", r.Timestamp.Local())
			fmt.Printf("Method:\t%s\n", r.Method)
			fmt.Printf("URL:\t%s\n", r.URL)
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
