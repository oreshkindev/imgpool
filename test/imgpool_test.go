package test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

const BASE_URL = "http://localhost:9000/api/v1"

// Download task
func TestDownload(t *testing.T) {

	tests := []struct {
		testName string
		input    string
		code     int
	}{
		{
			testName: "valid",
			input:    "tcuAxhxKQF100.jpeg",
			code:     http.StatusOK,
		},
		{
			testName: "invalid not found",
			input:    "fsdDvrsfcc1.jpeg",
			code:     http.StatusNotFound,
		},
		{
			testName: "invalid missed uri",
			input:    " ",
			code:     http.StatusNotFound,
		},
	}

	for _, test := range tests {
		t.Logf("running %v", test.testName)

		c := resty.New()
		r, e := c.R().Get(BASE_URL + "/image/download/" + test.input)
		if e != nil {
			assert.Equal(t, test.code, e.Error())
		} else {
			assert.Equal(t, test.code, r.StatusCode())
		}
	}
}

// Get task
func TestGet(t *testing.T) {

	tests := []struct {
		testName string
		input    string
		code     int
	}{
		{
			testName: "valid",
			input:    "98",
			code:     http.StatusOK,
		},
		{
			testName: "invalid not found",
			input:    "37",
			code:     http.StatusNotFound,
		},
		{
			testName: "invalid wrong type",
			input:    "string",
			code:     http.StatusBadRequest,
		},
		{
			testName: "invalid missed id",
			code:     http.StatusNotFound,
		},
	}

	for _, test := range tests {
		t.Logf("running %v", test.testName)

		c := resty.New()
		r, e := c.R().Get(BASE_URL + "/image/" + test.input)
		if e != nil {
			assert.Equal(t, test.code, e.Error())
		} else {
			assert.Equal(t, test.code, r.StatusCode())
		}
	}
}

// Post image & params
func TestPost(t *testing.T) {
	tests := []struct {
		testName string
		width    string
		height   string
		image    string
		code     int
	}{
		{
			testName: "valid",
			width:    "37",
			height:   "37",
			image:    "./tmp/go.png",
			code:     http.StatusOK,
		},
		{
			testName: "invalid missed width",
			height:   "37",
			image:    "./tmp/go.png",
			code:     http.StatusBadRequest,
		},
		{
			testName: "invalid missed height",
			width:    "37",
			image:    "./tmp/go.png",
			code:     http.StatusBadRequest,
		},
		{
			testName: "invalid missed image",
			width:    "37",
			height:   "37",
			code:     http.StatusBadRequest,
		},
		{
			testName: "invalid wrong type",
			width:    "37",
			height:   "37",
			image:    "./tmp/go.txt",
			code:     http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Logf("running %v", test.testName)

		imageBytes, e := ioutil.ReadFile(test.image)
		if e != nil {
			assert.Equal(t, test.code, test.code, e.Error())
			return
		}

		c := resty.New()
		r, e := c.R().
			SetFileReader("image", "go.png", bytes.NewReader(imageBytes)).
			SetFormData(map[string]string{"width": test.width, "height": test.height}).
			Post(BASE_URL + "/image")

		if e != nil {
			assert.Equal(t, test.code, e.Error())
		} else {
			assert.Equal(t, test.code, r.StatusCode())
		}
	}
}
