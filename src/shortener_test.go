package main

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

var router *gin.Engine

func init() {
	router = RouterSetup()
}

func TestHashHandler(t *testing.T) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/url", strings.NewReader("{\"url\":\"https://www.google.ca\"}"))
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	body, _ := ioutil.ReadAll(recorder.Body)
	assert.Equal(t, "{\"shortenedUrl\":\"http://localhost:8123/fe9970\"}", string(body))
}

func TestNegativeHashHandler(t *testing.T) {
	recorder := httptest.NewRecorder()
	// Url without protocol considered as incorrect
	req := httptest.NewRequest(http.MethodPost, "/url", strings.NewReader("{\"url\":\"www.google.ca\"}"))
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestNegativeRedirectHandler(t *testing.T) {
	recorder := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodPost, "/url", strings.NewReader("{\"url\":\"https://www.google.ca\"}"))
	req2 := httptest.NewRequest(http.MethodGet, "/fe9971", nil)

	router.ServeHTTP(recorder, req1)
	assert.Equal(t, http.StatusOK, recorder.Code)

	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, req2)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestRedirectHandler(t *testing.T) {
	recorder := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodPost, "/url", strings.NewReader("{\"url\":\"https://www.google.ca\"}"))
	req2 := httptest.NewRequest(http.MethodGet, "/fe9970", nil)

	router.ServeHTTP(recorder, req1)
	assert.Equal(t, http.StatusOK, recorder.Code)

	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, req2)
	assert.Equal(t, http.StatusPermanentRedirect, recorder.Code)
}
func TestTtlHandler(t *testing.T) {
	recorder := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodPost, "/url", strings.NewReader("{\"url\":\"https://www.google.ca\"}"))
	req2 := httptest.NewRequest(http.MethodGet, "/fe9970", nil)
	req3 := httptest.NewRequest(http.MethodGet, "/internal/ttl", nil)

	router.ServeHTTP(recorder, req1)
	assert.Equal(t, http.StatusOK, recorder.Code)

	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, req2)
	assert.Equal(t, http.StatusPermanentRedirect, recorder.Code)

	recorder = httptest.NewRecorder()
	time.Sleep(16 * time.Second) // ttl equals to 15 seconds, that's why we want to wait here
	router.ServeHTTP(recorder, req3)
	assert.Equal(t, http.StatusOK, recorder.Code)
	body, _ := ioutil.ReadAll(recorder.Body)
	assert.Equal(t, "{\"outdatedEntriesCount\":1}", string(body))
}

func TestLogsFilePresent(t *testing.T) {
	defer func() {
		/*
			shortener.log in src folder is created during the ::RouterSetup
			deferred function is used to clean up after tests
		*/
		err := os.Remove("shortener.log")
		if err != nil {
			t.FailNow()
		}
	}()
	testLogFileName := "shortener.log"
	// permissions are not necessary for opening an existing file, that's why we can put 0000
	f, err := os.OpenFile(testLogFileName, os.O_RDONLY, 0000)
	if err != nil {
		t.FailNow()
	}

	assert.Equal(t, testLogFileName, f.Name())
}
