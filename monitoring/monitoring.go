package monitoring

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/clagraff/pobox/requests"
	"github.com/nbari/violetear"
	uuid "github.com/satori/go.uuid"
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

var cachedRequests = make(map[uuid.UUID]requests.Request)

func listRequests(w http.ResponseWriter, r *http.Request) {
	bites, err := json.Marshal(cachedRequests)
	if err != nil {
		panic(err)
	}

	w.Write(bites)
}

func retrieveRequest(w http.ResponseWriter, r *http.Request) {
	requestedUUID := violetear.GetParam("uuid", r)
	uuidToRetrieve := uuid.Must(uuid.FromString(requestedUUID))

	if req, ok := cachedRequests[uuidToRetrieve]; ok {
		bites, err := json.Marshal(req)
		if err != nil {
			panic(err)
		}

		w.Write(bites)

	} else {
		panic("request uuid not found")
	}
}

func clearRequests(w http.ResponseWriter, r *http.Request) {
	cachedRequests = make(map[uuid.UUID]requests.Request)
}

func deleteRequest(w http.ResponseWriter, r *http.Request) {
	requestedUUID := violetear.GetParam("uuid", r)

	uuidToDelete := uuid.Must(uuid.FromString(requestedUUID))
	delete(cachedRequests, uuidToDelete)
}

func storeRequests(receivedRequests chan requests.Request) {
	for {
		select {
		case r := <-receivedRequests:
			requestUUID := uuid.Must(uuid.NewV4())
			cachedRequests[requestUUID] = r
		default:
		}
	}
}

func CreateServer(receivedRequests chan requests.Request, port int) (*http.Server, func() error) {
	listenAddr := fmt.Sprintf(":%d", port)

	router := violetear.New()
	router.LogRequests = false
	router.Verbose = false

	router.AddRegex(
		":uuid",
		`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`,
	)

	router.HandleFunc("/requests", listRequests, "GET")
	router.HandleFunc("/requests", clearRequests, "DELETE")

	router.HandleFunc("/requests/:uuid", retrieveRequest, "GET")
	router.HandleFunc("/requests/:uuid", deleteRequest, "DELETE")

	srv := &http.Server{
		Addr:           listenAddr,
		Handler:        router,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   7 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() { storeRequests(receivedRequests) }()

	return srv, srv.ListenAndServe
}
