package endpoints

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/nbari/violetear"
)

type DefinedRequest struct {
	URI    string
	Method string
}

type DefinedResponse struct {
	Body   json.RawMessage
	Status int
}

type DefinedRoute struct {
	Request  DefinedRequest
	Response DefinedResponse
}

// MappedRoutes -> map[URI Pattern]map[Method Pattern]DefinedResponse
type MappedRoutes map[*regexp.Regexp]map[*regexp.Regexp]DefinedResponse

func copyRequest(original *http.Request) (http.Request, error) {
	copy := *original

	// Use new URL Userinfo struct instead of the original pointer
	if original.URL.User != nil {
		// TODO: does this actually work?
		copy.URL.User = &url.Userinfo{}
		(*copy.URL.User) = *original.URL.User
	}

	buf, err := ioutil.ReadAll(original.Body)
	if err != nil {
		return copy, err
	}

	// User new Body and GetBody instead of original ones.
	copy.Body = ioutil.NopCloser(bytes.NewBuffer(buf))

	origBody := ioutil.NopCloser(bytes.NewBuffer(buf))
	original.Body = origBody
	original.GetBody = func() (io.ReadCloser, error) {
		return original.Body, nil
	}

	// We dont support the copying of this information.
	copy.MultipartForm = nil
	copy.TLS = nil
	copy.Response = nil

	return copy, nil
}

func createCatchAllRoute(routes MappedRoutes, receivedRequests chan http.Request) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		copiedReq, err := copyRequest(r)
		if err != nil {
			panic(err)
		}

		receivedRequests <- copiedReq

		for uriPattern, methodMapping := range routes {
			if uriPattern.MatchString(r.URL.RequestURI()) {
				for methodPattern, definedResponse := range methodMapping {
					if methodPattern.MatchString(r.Method) {
						if definedResponse.Status == 0 {
							definedResponse.Status = 200
						}
						w.WriteHeader(definedResponse.Status)
						w.Write([]byte(definedResponse.Body))
						return
					}
				}
			}
		}

		w.Write([]byte(""))
	}
}

func mapDefinedRoutes(routes []DefinedRoute) MappedRoutes {
	mapped := make(MappedRoutes)

	for _, route := range routes {
		if len(route.Request.URI) <= 0 {
			panic("defined route missing a URI")
		}

		if len(route.Request.Method) <= 0 {
			panic("defined route missing a Method")
		}

		uriRegex := regexp.MustCompile(route.Request.URI)
		methodRegex := regexp.MustCompile(route.Request.Method)

		if _, ok := mapped[uriRegex]; !ok {
			mapped[uriRegex] = make(map[*regexp.Regexp]DefinedResponse)
		}

		mapped[uriRegex][methodRegex] = route.Response
	}

	return mapped
}

func CreateServer(routes []DefinedRoute, receivedRequests chan http.Request, port int) (*http.Server, func() error) {
	listenAddr := fmt.Sprintf(":%d", port)
	mapped := mapDefinedRoutes(routes)

	router := violetear.New()
	router.LogRequests = false
	router.Verbose = false

	router.HandleFunc("*", createCatchAllRoute(mapped, receivedRequests))

	srv := &http.Server{
		Addr:           listenAddr,
		Handler:        router,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   7 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return srv, srv.ListenAndServe
}
