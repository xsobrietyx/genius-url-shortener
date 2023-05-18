# url-shortener app (genius task)
Url shortener service aka https://www.bitly.com
## Run the tests
```shell
$ go test -v ./src
```
## Run service
```shell
$ go run ./src
```
## Or build service into executable and execute the binary
```shell
$ go build -o shortenerapp ./src/shortener.go && ./shortenerapp
```
## Instead of web interface
You can use Postman collection that you can import in your Postman app - to verify the application workflow.  
Postman collection can be found in ***./postman*** folder.
## Endpoints
1. **host:port/url** - endpoint used to hash your url provided. **http method - POST**
   1. *JSON example: { "url": "https://www.amazon.ca/" }*
   2. *Response example: { "shortenedUrl": "http://localhost:8123/fe9970" }*
2. **host:port/:hashcode** - endpoint used to redirect you to the url that will be "unshortened". **http method - GET**
   1. *Url example: host:port/fe9970*
3. **host:port/internal/ttl** - endpoint used to clean-up internal state and remove outdated entries. **http method - GET**
   1. *Url example: host:port/internal/ttl*
   2. *Response example: { "outdatedEntriesCount": 0 }*
## Info
* default reference ttl == 15 seconds, afterwards you can call the endpoint #3 for outdated data eviction. Data can only be deleted manually through the REST call.
* ***shortener.log*** will appear in folder from which binary will be called or in the root sources folder.
## Depends on
* [Golang](https://go.dev)
* [Gin](https://github.com/gin-gonic/gin)
* [Testify](https://github.com/stretchr/testify)
* Other indirect/transitive dependencies
