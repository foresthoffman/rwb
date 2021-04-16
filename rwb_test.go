/**
 * rwb_test.go
 *
 * Copyright (c) 2021 Forest Hoffman. All Rights Reserved.
 * License: MIT License (see the included LICENSE file) or download at
 *     https://raw.githubusercontent.com/foresthoffman/rwb/master/LICENSE
 */

package rwb

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

var writeTestCases = []struct {
	Name         string
	ResponseBody []byte
	ExpectedBody []byte
	Flush        bool
}{
	{
		Name:         "write",
		ResponseBody: []byte("hello 123"),
		ExpectedBody: []byte("hello 123"),
		Flush:        true,
	},
	{
		Name:         "write_no_flush",
		ResponseBody: []byte("hello 123"),
		ExpectedBody: []byte(""),
		Flush:        false,
	},
}

func TestResponseWriterBuffer_Write(t *testing.T) {
	for _, testCase := range writeTestCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				rwb := New(w)
				// Writing to ResponseWriterBuffer.
				n, err := rwb.Write(testCase.ResponseBody)
				if err != nil {
					t.Error(err)
				}
				if n != len(testCase.ResponseBody) {
					t.Errorf("expected: %d got: %d", len(testCase.ResponseBody), n)
				}
				if testCase.Flush {
					// Send the buffered data to the ResponseWriter.
					n, err := rwb.Flush()
					if err != nil {
						t.Error(err)
					}
					if n != len(testCase.ResponseBody) {
						t.Errorf("expected: %d got: %d", len(testCase.ResponseBody), n)
					}
				}
			}))
			defer ts.Close()

			res, err := http.Get(ts.URL)
			if err != nil {
				t.Error(err)
			}
			b, err := ioutil.ReadAll(res.Body)
			res.Body.Close()
			if err != nil {
				t.Error(err)
			}

			if string(b) != string(testCase.ExpectedBody) {
				t.Errorf("expected: %q got: %q", string(b), string(testCase.ExpectedBody))
			}
		})
	}
}

var (
	headerTestCases = []struct {
		Name              string
		StartingHeaders   http.Header
		AdditionalHeaders http.Header
		ExpectedHeaders   http.Header
		Flush             bool
	}{
		{
			Name:              "one_single_value",
			StartingHeaders:   http.Header{"Content-Type": []string{"application/json"}},
			AdditionalHeaders: http.Header{},
			ExpectedHeaders:   http.Header{"Content-Type": []string{"application/json"}},
			Flush:             true,
		},
		{
			Name:              "one_single_value_no_flush",
			StartingHeaders:   http.Header{"Content-Type": []string{"application/json"}},
			AdditionalHeaders: http.Header{},
			ExpectedHeaders:   http.Header{"Content-Type": []string{"application/json"}},
			Flush:             false,
		},
		{
			Name: "multiple_single_value",
			StartingHeaders: http.Header{
				"Content-Type": []string{"application/json"},
				"sandwich":     []string{"BLT"},
			},
			AdditionalHeaders: http.Header{},
			ExpectedHeaders: http.Header{
				"Content-Type": []string{"application/json"},
				"sandwich":     []string{"BLT"},
			},
			Flush: true,
		},
		{
			Name: "multiple_single_value_no_flush",
			StartingHeaders: http.Header{
				"Content-Type": []string{"application/json"},
				"sandwich":     []string{"BLT"},
			},
			AdditionalHeaders: http.Header{},
			ExpectedHeaders: http.Header{
				"Content-Type": []string{"application/json"},
				"sandwich":     []string{"BLT"},
			},
			Flush: false,
		},
		{
			Name:              "one_single_value_additional",
			StartingHeaders:   http.Header{},
			AdditionalHeaders: http.Header{"Content-Type": []string{"application/json"}},
			ExpectedHeaders:   http.Header{"Content-Type": []string{"application/json"}},
			Flush:             true,
		},
		{
			Name:              "one_single_value_additional_no_flush",
			StartingHeaders:   http.Header{},
			AdditionalHeaders: http.Header{"Content-Type": []string{"application/json"}},
			ExpectedHeaders:   http.Header{},
			Flush:             false,
		},
		{
			Name:            "multiple_single_value_additional",
			StartingHeaders: http.Header{},
			AdditionalHeaders: http.Header{
				"Content-Type": []string{"application/json"},
				"sandwich":     []string{"BLT"},
			},
			ExpectedHeaders: http.Header{
				"Content-Type": []string{"application/json"},
				"sandwich":     []string{"BLT"},
			},
			Flush: true,
		},
		{
			Name:            "multiple_single_value_additional_no_flush",
			StartingHeaders: http.Header{},
			AdditionalHeaders: http.Header{
				"Content-Type": []string{"application/json"},
				"sandwich":     []string{"BLT"},
			},
			ExpectedHeaders: http.Header{},
			Flush:           false,
		},
	}
)

func TestResponseWriterBuffer_Header(t *testing.T) {
	for _, testCase := range headerTestCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Header has some values before the ResponseWriterBuffer takes over.
				for key, values := range testCase.StartingHeaders {
					for _, value := range values {
						w.Header().Add(key, value)
					}
				}
				rwb := New(w)
				// Header has some new values.
				for key, values := range testCase.AdditionalHeaders {
					for _, value := range values {
						rwb.Header().Add(key, value)
					}
				}
				if testCase.Flush {
					// Send the buffered data to the ResponseWriter.
					_, err := rwb.Flush()
					if err != nil {
						t.Error(err)
					}
				}
			}))
			defer ts.Close()

			res, err := http.Get(ts.URL)
			if err != nil {
				t.Error(err)
			}
			for key, expectedValues := range testCase.ExpectedHeaders {
				actualValues := res.Header.Values(key)
				if len(actualValues) == 0 {
					t.Errorf("missing header: %q with value %v", key, expectedValues)
				}
				if len(actualValues) > len(expectedValues) {
					t.Errorf("expected: %v got: %v", expectedValues, actualValues)
				}
				for _, expectedValue := range expectedValues {
					found := false
					for _, actualValue := range actualValues {
						if actualValue == expectedValue {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected: %v got: %v", expectedValues, actualValues)
					}
				}
			}
		})
	}
}
