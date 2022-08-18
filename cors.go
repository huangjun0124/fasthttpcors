package fasthttpcors

import (
	"os"
	"strconv"
	"strings"

	"github.com/valyala/fasthttp"
)

// Options is struct that defined cors properties
type Options struct {
	AllowedOrigins   []string
	AllowedHeaders   []string
	AllowMaxAge      int
	AllowedMethods   []string
	ExposedHeaders   []string
	AllowCredentials bool
	Debug            bool // 是否开启默认日志
	Logger           Logger
}

type CorsHandler struct {
	allowedOriginsAll bool
	allowedHeadersAll bool
	// Normalized list of plain allowed origins
	allowedOrigins []string
	// List of allowed origins containing wildcards
	allowedWOrigins  []wildcard
	allowedHeaders   []string
	allowedMethods   []string
	exposedHeaders   []string
	allowCredentials bool
	maxAge           int
	logger           Logger
}

var defaultOptions = &Options{
	AllowedOrigins: []string{"*"},
	AllowedMethods: []string{"GET", "POST"},
	AllowedHeaders: []string{"Origin", "Accept", "Content-Type"},
}

func DefaultHandler() *CorsHandler {
	return NewCorsHandler(*defaultOptions)
}

func NewCorsHandler(options Options) *CorsHandler {
	logger := OffLogger()
	if options.Logger != nil {
		logger = options.Logger
	} else if options.Debug {
		logger = NewLogger(os.Stdout)
	}
	cors := &CorsHandler{
		allowedOrigins:   options.AllowedOrigins,
		allowedHeaders:   options.AllowedHeaders,
		allowCredentials: options.AllowCredentials,
		allowedMethods:   options.AllowedMethods,
		exposedHeaders:   options.ExposedHeaders,
		maxAge:           options.AllowMaxAge,
		logger:           logger,
	}
	cors.RefreshAllowOrigins(options.AllowedOrigins)
	if len(cors.allowedHeaders) == 0 {
		cors.allowedHeaders = defaultOptions.AllowedHeaders
		cors.allowedHeadersAll = true
	} else {
		for _, v := range options.AllowedHeaders {
			if v == "*" {
				cors.allowedHeadersAll = true
				break
			}
		}
	}
	if len(cors.allowedMethods) == 0 {
		cors.allowedMethods = defaultOptions.AllowedMethods
	}
	return cors
}

// RefreshAllowOrigins
//  @desc: 支持运行时动态刷新 allowedOrigins 配置
//  @receiver c
//  @para allowedOrigins
//
func (c *CorsHandler) RefreshAllowOrigins(allowedOrigins []string) {
	if len(allowedOrigins) == 0 {
		c.allowedOrigins = defaultOptions.AllowedOrigins
		c.allowedOriginsAll = true
	} else {
		c.allowedOrigins = []string{}
		c.allowedWOrigins = []wildcard{}
		for _, origin := range allowedOrigins {
			// Normalize
			origin = strings.ToLower(origin)
			if origin == "*" {
				// If "*" is present in the list, turn the whole list into a match all
				c.allowedOriginsAll = true
				c.allowedOrigins = nil
				c.allowedWOrigins = nil
				break
			} else if i := strings.IndexByte(origin, '*'); i >= 0 {
				// Split the origin in two: start and end string without the *
				w := wildcard{origin[0:i], origin[i+1:]}
				c.allowedWOrigins = append(c.allowedWOrigins, w)
			} else {
				c.allowedOrigins = append(c.allowedOrigins, origin)
			}
		}
	}
}

func (c *CorsHandler) CorsMiddleware(innerHandler fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		if string(ctx.Method()) == "OPTIONS" {
			c.handlePreflight(ctx)
			ctx.SetStatusCode(200)
		} else {
			c.handleActual(ctx)
			innerHandler(ctx)
		}
	}
}

func (c *CorsHandler) handlePreflight(ctx *fasthttp.RequestCtx) {
	originHeader := string(ctx.Request.Header.Peek("Origin"))
	if len(originHeader) == 0 || c.isAllowedOrigin(originHeader) == false {
		c.logger.Log("Origin ", originHeader, " is not in", c.allowedOrigins)
		return
	}
	method := string(ctx.Request.Header.Peek("Access-Control-Request-Method"))
	if !c.isAllowedMethod(method) {
		c.logger.Log("Method ", method, " is not in", c.allowedMethods)
		return
	}
	headers := []string{}
	if len(ctx.Request.Header.Peek("Access-Control-Request-Headers")) > 0 {
		headers = strings.Split(string(ctx.Request.Header.Peek("Access-Control-Request-Headers")), ",")
	}
	if !c.areHeadersAllowed(headers) {
		c.logger.Log("Headers ", headers, " is not in", c.allowedHeaders)
		return
	}

	ctx.Response.Header.Set("Access-Control-Allow-Origin", originHeader)
	ctx.Response.Header.Set("Access-Control-Allow-Methods", method)
	if len(headers) > 0 {
		ctx.Response.Header.Set("Access-Control-Allow-Headers", strings.Join(headers, ", "))
	}
	if c.allowCredentials {
		ctx.Response.Header.Set("Access-Control-Allow-Credentials", "true")
	}
	if c.maxAge > 0 {
		ctx.Response.Header.Set("Access-Control-Max-Age", strconv.Itoa(c.maxAge))
	}
}

func (c *CorsHandler) handleActual(ctx *fasthttp.RequestCtx) {
	originHeader := string(ctx.Request.Header.Peek("Origin"))
	if len(originHeader) == 0 || c.isAllowedOrigin(originHeader) == false {
		if len(originHeader) != 0 {
			c.logger.Log("Origin ", originHeader, " is not in", c.allowedOrigins)
		}
		return
	}
	ctx.Response.Header.Set("Access-Control-Allow-Origin", originHeader)
	if len(c.exposedHeaders) > 0 {
		ctx.Response.Header.Set("Access-Control-Expose-Headers", strings.Join(c.exposedHeaders, ", "))
	}
	if c.allowCredentials {
		ctx.Response.Header.Set("Access-Control-Allow-Credentials", "true")
	}
}

func (c *CorsHandler) isAllowedOrigin(originHeader string) bool {
	if c.allowedOriginsAll {
		return true
	}
	originHeader = strings.ToLower(originHeader)
	for _, o := range c.allowedOrigins {
		if o == originHeader {
			return true
		}
	}
	for _, w := range c.allowedWOrigins {
		if w.match(originHeader) {
			return true
		}
	}
	return false
}

func (c *CorsHandler) isAllowedMethod(methodHeader string) bool {
	if len(c.allowedMethods) == 0 {
		return false
	}
	if methodHeader == "OPTIONS" {
		return true
	}
	for _, m := range c.allowedMethods {
		if m == methodHeader {
			return true
		}
	}
	return false
}

func (c *CorsHandler) areHeadersAllowed(headers []string) bool {
	if c.allowedHeadersAll || len(headers) == 0 {
		return true
	}
	for _, header := range headers {
		found := false
		for _, h := range c.allowedHeaders {
			if h == header {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	return true
}
