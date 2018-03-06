package requests

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

// CopyHTTPRequest returns a copy of the original request instance.
func CopyHTTPRequest(original *http.Request) (http.Request, error) {
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

func FromHTTPRequest(r http.Request, readBody bool) Request {
	req := Request{
		Headers: make(map[string]string),
		Host:    r.Host,
		Method:  r.Method,
		Proto:   r.Proto,
		URI:     r.URL.RequestURI(),
	}

	for key, value := range r.Header {
		req.Headers[key] = strings.Join(value, "")
	}

	if !readBody {
		return req
	}

	bites, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	req.Body = bites

	return req
}

type Request struct {
	Body    []byte
	Headers map[string]string
	Host    string
	Method  string
	Proto   string
	URI     string
}

func (r Request) String() string {
	buff := new(bytes.Buffer)
	buff.WriteString(fmt.Sprintf("%s %s %s\n", r.Method, r.URI, r.Proto))
	buff.WriteString(fmt.Sprintf("Host: %s\n", r.Host))

	if len(r.Headers) > 0 {
		keys := r.HeaderKeys()
		sort.Strings(keys)

		for _, key := range keys {
			buff.WriteString(fmt.Sprintf("%s: %s\n", key, r.Headers[key]))
		}
	}

	if len(r.Body) > 0 {
		buff.WriteString("\n")
		buff.Write(r.Body)
		buff.WriteString("\n")
	}

	return buff.String()
}

func (r Request) HeaderKeys() []string {
	keys := make([]string, len(r.Headers))

	i := 0
	for key := range r.Headers {
		keys[i] = key
		i++
	}

	return keys
}
