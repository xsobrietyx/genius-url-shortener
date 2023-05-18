package main

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// hashedUrl type alias for holding a hashed url string
type hashedUrl string

// entry struct for internal state. ttl - the time our url should be "persisted internally". value - full url
type entry struct {
	ttl   time.Time
	value string
}

// applicationState holds the internal app state and sync.RWMutex for proper state synchronization
type applicationState struct {
	data map[hashedUrl]entry
	mu   sync.RWMutex
}

// appState is not persistent in DB for sake of simplicity
var appState = applicationState{
	data: make(map[hashedUrl]entry),
	mu:   sync.RWMutex{},
}

// hashRequest struct for request with full url that should be shortened
type hashRequest struct {
	Url string `json:"url" binding:"required"`
}

/*
	hashingHandler performs operations for url shortening and internal state persistence.
	Returns 200 or 400 from url validation middleware
*/
func hashingHandler(c *gin.Context) {
	urlForHash := c.MustGet("url").(string)

	hash := md5.New()
	_, err := io.WriteString(hash, urlForHash)
	hashed := hex.EncodeToString(hash.Sum(nil))
	shortened := ""
	for i, char := range hashed {
		if i <= 5 {
			shortened += string(char)
		}
	}
	if err == nil {
		appState.mu.Lock()
		defer appState.mu.Unlock()
		appState.data[hashedUrl(shortened)] = entry{
			value: urlForHash,
			ttl:   time.Now(),
		}
	} else {
		log.Printf("Error during the hashingHandler invocation, writing to the hash: %s\n", err.Error())
	}

	jsonData := map[string]interface{}{
		//TODO: port could be parametrized, potential feature
		"shortenedUrl": "http://localhost:8123/" + shortened,
	}

	c.JSON(http.StatusOK, jsonData)
}

// redirectHandler fetches full url from internal state and performs redirect. Returns 308 and redirect if ok and 400 otherwise
func redirectHandler(c *gin.Context) {
	key := c.Param("hash")
	appState.mu.RLock()
	defer appState.mu.RUnlock()
	for k, v := range appState.data {
		if strings.HasPrefix(string(k), key) {
			c.Redirect(http.StatusPermanentRedirect, v.value)
			return
		}
	}
	c.Status(http.StatusBadRequest)
}

// ttlCleanupHandler removes stale data from internal state
func ttlCleanupHandler(c *gin.Context) {
	updatedState := make(map[hashedUrl]entry)
	outdatedEntriesCount := 0
	now := time.Now()
	appState.mu.RLock()
	for k, v := range appState.data {
		//TODO: ttl limit could be parametrized, potential feature
		if now.Sub(v.ttl).Seconds() < 15 { // 15 seconds is used for demo purposes
			updatedState[k] = v
		} else {
			outdatedEntriesCount++
		}
	}
	appState.mu.RUnlock()

	appState = applicationState{
		data: updatedState,
		mu:   sync.RWMutex{},
	}

	jsonData := map[string]interface{}{
		"outdatedEntriesCount": outdatedEntriesCount,
	}

	c.JSON(http.StatusOK, jsonData)
}

// urlValidator is a middleware function for request data validation purposes
func urlValidator() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Checking do we need to apply the validation function
		if c.Request.URL.Path == "/url" {
			var request hashRequest
			err := c.BindJSON(&request)
			if err != nil {
				log.Printf("Error during the hashingHandler invocation, JSON binding: %s\n", err.Error())
				c.Status(http.StatusBadRequest)
				c.Abort()
			}

			// url validation function
			validateUrl := func(u string) error {
				e := errors.New("error during the url validation. Invalid url")
				if _, err := url.ParseRequestURI(u); err == nil {
					e = nil
				}
				return e
			}

			// Apply url validation function
			err = validateUrl(request.Url)
			if err != nil {
				log.Printf("Error during the hashingHandler invocation: %s\n", err.Error())
				c.Status(http.StatusBadRequest)
				c.Abort()
			}

			// Url is valid, we can set it into the context
			c.Set("url", request.Url)

			// Url is valid, we can proceed to the handler
			c.Next()
		}
	}
}

// RouterSetup sets up router, used for testing purposes as well
func RouterSetup() *gin.Engine {
	f, err := os.Create("./shortener.log")
	if err != nil {
		log.Println(err.Error())
	}
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	_ = r.SetTrustedProxies(nil)
	r.Use(urlValidator())

	r.GET("/:hash", redirectHandler)
	r.GET("/internal/ttl", ttlCleanupHandler)
	r.POST("/url", hashingHandler)

	return r
}

// main - application entrypoint
func main() {
	log.SetPrefix("[genius-url-shortener-app]")

	router := RouterSetup()

	//TODO: application port could be parametrized, potential feature
	err := router.Run(":8123")

	if err != nil {
		log.Fatalf("Error during the main method invocation, server start: %s\n", err.Error())
	}
}
