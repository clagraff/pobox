package endpoints

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/clagraff/pobox/requests"
	"github.com/labstack/gommon/color"
	"github.com/nbari/violetear"
)

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

func createCatchAllRoute(receivedRequests chan requests.Request) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer logHit(r.URL.String())
		copiedReq, err := copyRequest(r)
		if err != nil {
			panic(err)
		}

		receivedRequests <- requests.FromHTTPRequest(copiedReq, true)

		w.Write([]byte(""))
	}
}

func logHit(url string) {
	color.Println(
		color.Yellow(time.Now().Local().String()),
		"\t",
		color.Cyan("Served web-hook request:"),
		"\t",
		url,
	)
}

func CreateServer(receivedRequests chan requests.Request, port int) (*http.Server, func() error) {
	listenAddr := fmt.Sprintf(":%d", port)

	router := violetear.New()
	router.LogRequests = false
	router.Verbose = false

	router.HandleFunc("*", createCatchAllRoute(receivedRequests))

	srv := &http.Server{
		Addr:           listenAddr,
		Handler:        router,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   7 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return srv, srv.ListenAndServe
}
