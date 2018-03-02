![](.github/postbox.png)

[![CircleCI](https://circleci.com/gh/clagraff/pobox.svg?style=svg)](https://circleci.com/gh/clagraff/pobox)
[![GoDoc](https://godoc.org/github.com/clagraff/pobox?status.svg)](https://godoc.org/github.com/clagraff/pobox)
[![Go Report Card](http://goreportcard.com/badge/clagraff/pobox)](http://goreportcard.com/report/clagraff/pobox)

# pobox
POBox is a eary-to-run web server which will accept and log _all_ incoming
requests.

POBox provides a simple solution for mocking API requests and responses, as well
as a method for loggin web hooks and integrations.

## Quickstart
POBox is a Golang application. As such, you will need to Go installed in order
to run the application. 

```bash
go install github.com/clagraff/pobox
pobox
```

After that, use cURL to see it in action:
```bash
curl localhost:8080
curl -X POST -d 'this is data!' localhost:8080/whatever/route/i/want
curl -X OPTIONS localhost:8080/fiiz/buzz?anything=here
```
