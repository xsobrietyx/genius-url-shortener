package main

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

/*
	TODO:
		1. REST server with 3 handlers (first is for hashing the string input, second for de-hashing it, third for cleanup purposes) - ?
		2. internal hashmap that holds hashed as a key and URLs as a value + ttl - √
		3. pick up proper hashing function - √ (md5)
		4. think about the unit tests, do we need them - ?
		5. docs - √
		6. md file
*/

// hashedUrl type alias for holding a hashed url string
type hashedUrl string

// entry struct for internal state. ttl - the time our url should be "persisted internally". value - full url
type entry struct {
	ttl   time.Time
	value string
}

// state alias for hashmap for internal state persistence. It's a key-value pair (shortUrl -> full url)
type applicationState map[hashedUrl]entry

// appState is not persistent in DB for sake of simplicity
var appState = make(applicationState)

// hashRequest struct for request with full url that should be shortened
type hashRequest struct {
	Url string `json:"url" binding:"required"`
}

// hashingHandler performs operations for url shortening and internal state persistence. Returns 200 or 400 otherwise
func hashingHandler(c *gin.Context) {
	var request hashRequest
	err := c.BindJSON(&request)
	if err != nil {
		log.Printf("Error json binding during hashRequest went wrong: %s\n", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}
	hash := md5.New()
	_, err = io.WriteString(hash, request.Url)
	hashed := hashedUrl(hex.EncodeToString(hash.Sum(nil)))
	if err == nil {
		appState[hashed] = entry{
			value: request.Url,
			ttl:   time.Now(),
		}
	} else {
		log.Printf("Error during the hashing: %s\n", err.Error())
	}
	shortened := ""
	for i, char := range string(hashed) {
		if i <= 5 {
			shortened += string(char)
		}
	}
	c.IndentedJSON(http.StatusOK, "http://localhost:8123/"+shortened)
}

// redirectHandler fetches full url from internal state and performs redirect. Returns 308 and redirect if ok and 400 otherwise
func redirectHandler(c *gin.Context) {
	key := c.Param("hash")
	for k, v := range appState {
		if strings.HasPrefix(string(k), key) {
			c.Redirect(http.StatusPermanentRedirect, v.value)
			return
		}
	}
	c.Status(http.StatusBadRequest)
}

// ttlCleanupHandler removes stale data from internal state
func ttlCleanupHandler(c *gin.Context) {
	updatedState := make(applicationState)
	outdatedEntriesCount := 0
	now := time.Now()
	for k, v := range appState {
		// TODO: ttl limit could be parametrized, potential feature
		if now.Sub(v.ttl).Seconds() < (24 * 5) {
			updatedState[k] = v
		} else {
			outdatedEntriesCount++
		}
	}

	appState = updatedState

	c.IndentedJSON(http.StatusOK, outdatedEntriesCount)
}

func main() {
	log.SetPrefix("[genius-url-shortener-app]")
	router := gin.Default()

	router.GET("/:hash", redirectHandler)
	router.GET("/internal/ttl", ttlCleanupHandler)
	router.POST("/url", hashingHandler)

	err := router.Run("localhost:8123")

	if err != nil {
		log.Fatalf("Error during the server start: %s\n", err.Error())
	}
}
