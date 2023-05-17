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
*/
type hashedUrl string
type entry struct {
	ttl   time.Time
	value string
}
type state map[hashedUrl]entry

// appState is not persistent for sake of simplicity
var appState = make(state)

type hashRequest struct {
	Url string `json:"url" binding:"required"`
}

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

func main() {
	log.SetPrefix("[genius-url-shortener-app]")
	router := gin.Default()

	router.GET("/:hash", redirectHandler)
	router.POST("/url", hashingHandler)

	router.POST("/internal/duplicates", nil)

	err := router.Run("localhost:8123")

	if err != nil {
		log.Fatalf("Error during the server start: %s\n", err.Error())
	}
}
