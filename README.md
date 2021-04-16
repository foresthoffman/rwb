## RWB - Response Writer Buffer

The rwb package exposes a wrapper for the `http.ResponseWriter`, which allows for intercepting HTTP(s) response bodies and response headers. The `ResponseWriterBuffer` is lightweight and can assist in HTTP(s) server request handling, as well as server response middleware.

Here's a guide to thinking in terms of *response* middleware and not just *request* middleware:

[Intercepting RESTful Responses with Middleware](https://dev.to/foresthoffman/intercepting-restful-responses-with-middleware-4b64)

### Installation

Run `go get -u github.com/foresthoffman/rwb`

### Importing

Import this package by including `github.com/foresthoffman/rwb` in your import block.

e.g.

```go
package main

import(
    ...
    "github.com/foresthoffman/rwb"
)
```

### Usage

Basic usage:

```go
package main

import (
	"github.com/foresthoffman/rwb"
	"log"
	"net/http"
)

func main() {
	// Hit http://localhost:9001/ in your browser, or cURL/wget it! It's up to you.
	http.ListenAndServe(":9001", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// You could update the header directly.
		w.Header().Set("Content-Type", "application/json")

		// New header exists.
		log.Println(w.Header().Get("Content-Type"))

		// Or you could write to the buffer, and flush it when you're done.
		writerBuf := rwb.New(w)
		writerBuf.Header().Set("potato", "russet")

		// New header doesn't exist yet. It's in the buffer!
		log.Println(w.Header().Get("potato"))

		_, err := writerBuf.Flush()
		if err != nil {
			panic(err)
		}

		// New header exists!
		log.Println(w.Header().Get("potato"))
	}))
}
```

_That's all, enjoy!_
