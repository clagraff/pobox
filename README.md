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
$ go install github.com/clagraff/pobox
$ pobox
```

After that, use cURL to see it in action:
```bash
$ curl localhost:8080
$ curl -X POST -d 'this is data!' localhost:8080/whatever/route/i/want
$ curl -X OPTIONS localhost:8080/fiiz/buzz?anything=here
```

## Routes
Using the `pobox <route-file>` argument, you can specify a `.json` or `.yaml`
file which supplies pre-defined routes.

These routes specify a regex pattern for which URIs and methods to react to,
and supply response information.

This is useful in the event a web-hook requires a specific response back.

Here is an example `.yaml` route file:
```yaml
# routes.yaml
-
    request:
        uri: /[fF]izz/buz{1,3}  # regex pattern for matching URIs
        method: get|post        # regex pattern for matching request methods
    response:
        body: |                 # Multi-line string response
            <html>
                <body>
                    <h1>My Response</h1>
                    <p>this is my response</p>
                </body>
            </html>
```

```bash
$ pobox routes.yaml &
$ curl http://localhost/hello/world
$ curl http://localhost/Fizz/buzz
<html>
    <body>
        <h1>My Response</h1>
        <p>this is my response</p>
    </body>
</html>
```
