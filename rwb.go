/**
 * rwb.go
 *
 * Copyright (c) 2021 Forest Hoffman. All Rights Reserved.
 * License: MIT License (see the included LICENSE file) or download at
 *     https://raw.githubusercontent.com/foresthoffman/rwb/master/LICENSE
 */

package rwb

import (
	"bytes"
	"errors"
	"net/http"
)

var ErrBufferClosed = errors.New("buffer closed")

// ResponseWriterBuffer simulates the functionality of the underlying ResponseWriter
// without sending headers or body bytes to the actual requesting client. Upon flushing
// the ResponseWriterBuffer, all captured header and body information is written to the
// underlying ResponseWriter.
//
// This allows the prepared response to be reviewed before sending the one-and-only
// response to the requesting client.
type ResponseWriterBuffer struct {
	http.ResponseWriter
	header     http.Header
	body       bytes.Buffer
	statusCode int
	closed     bool
}

// Header returns a copy of the ResponseWriter header map. It should be assumed that this
// version diverged from the original.
func (rw *ResponseWriterBuffer) Header() http.Header {
	return rw.header
}

// Write sends the provided bytes to a buffer instead of the requesting client.
func (rw *ResponseWriterBuffer) Write(body []byte) (int, error) {
	if rw.closed {
		return 0, ErrBufferClosed
	}
	rw.body.Reset()
	return rw.body.Write(body)
}

// WriteHeader stores a copy of the desired ResponseWriter header status code.
func (rw *ResponseWriterBuffer) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
}

// Flush takes all the buffered header and body values, and writes them to the underlying
// ResponseWriter. Returns the number of bytes written and any error.
func (rw *ResponseWriterBuffer) Flush() (int, error) {
	if rw.closed {
		return 0, ErrBufferClosed
	}
	// Remove keys that were deleted from the clone.
	actualHeader := rw.ResponseWriter.Header()
	for key := range actualHeader {
		if _, ok := rw.header[key]; !ok {
			actualHeader.Del(key)
		}
	}
	// Copy new header values from the clone.
	for key, values := range rw.header {
		for _, value := range values {
			found := false
			for _, actualValue := range actualHeader.Values(key) {
				if actualValue == value {
					found = true
					break
				}
			}
			if found {
				continue
			}
			actualHeader.Add(key, value)
		}
	}

	if rw.statusCode != 0 {
		rw.ResponseWriter.WriteHeader(rw.statusCode)
	}
	n, err := rw.ResponseWriter.Write(rw.body.Bytes())
	if err != nil {
		return 0, err
	}
	rw.closed = true
	return n, nil
}

// New creates a buffer for the provided ResponseWriter.
func New(w http.ResponseWriter) *ResponseWriterBuffer {
	return &ResponseWriterBuffer{
		ResponseWriter: w,
		header:         w.Header().Clone(),
		body:           bytes.Buffer{},
		statusCode:     0,
	}
}
