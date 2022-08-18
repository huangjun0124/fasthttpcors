# fasthttpcors
Cors handler for fasthttp server

Read First
------------
* This handler is forked from https://github.com/AdhityaRamadhanus/fasthttpcors
* Adding log input, add allowOrigin config * (eg: http://*.mydomain.com)
* This handler comply to w3c spec of CORS (even in case-insensitive comparison) https://www.w3.org/TR/cors/

Installation
------------
* go get github.com/huangjun0124/fasthttpcors

Usage
------------
```
package main

import (
	"log"

	cors "github.com/huangjun0124/fasthttpcors"
	"github.com/valyala/fasthttp"
)

func main() {
	withCors := cors.NewCorsHandler(cors.Options{
		// if you leave allowedOrigins empty then fasthttpcors will treat it as "*"
		AllowedOrigins: []string{"http://example.com"}, // Only allow example.com to access the resource
		// if you leave allowedHeaders empty then fasthttpcors will accept any non-simple headers
		AllowedHeaders: []string{"x-something-client", "Content-Type"}, // only allow x-something-client and Content-Type in actual request
		// if you leave this empty, only simple method will be accepted
		AllowedMethods:   []string{"GET", "POST"}, // only allow get or post to resource
		AllowCredentials: false,                   // resource doesn't support credentials
		AllowMaxAge:      5600,                    // cache the preflight result
		Debug:            true,
	})
	if err := fasthttp.ListenAndServe(":8080", withCors.CorsMiddleware(RequestHandler)); err != nil {
		log.Fatalf("Error in ListenAndServe: %s", err)
	}
}

func RequestHandler(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("plain/text")
	ctx.SetStatusCode(200)
	ctx.SetBodyString("OK")
}

```

TODO
-----
* add test