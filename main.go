package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/hoisie/web"
)

type HTTPDump struct {
	resp     string
	respFile string
	respCode int
	redirect string
	tls      bool
	tlsCert  string
	tlsKey   string
	ip       string
	port     string
}

func (d *HTTPDump) start() {
	// Load response file into resp field if specified
	if len(d.respFile) > 0 {
		d.resp = string(d.getFile(d.respFile))
	}

	// Setup handling for all supported verbs
	web.Get("/(.*)", d.handleGet)
	web.Post("/(.*)", d.handleBody)
	web.Put("/(.*)", d.handleBody)
	web.Delete("/(.*)", d.handleBody)
	web.Match("OPTIONS", "/(.*)", d.handleCORS)

	// Setup TLS
	if d.tls {
		cert := d.getFile(d.tlsCert)
		key := d.getFile(d.tlsKey)

		config := tls.Config{
			Time: nil,
		}

		config.Certificates = make([]tls.Certificate, 1)

		var err error
		config.Certificates[0], err = tls.X509KeyPair([]byte(cert), []byte(key))
		if err != nil {
			println(err.Error())
			return
		}

		web.RunTLS(d.ip+":"+d.port, &config)
	} else {
		web.Run(d.ip + ":" + d.port)
	}
}

// Read file from disk
func (d HTTPDump) getFile(path string) []byte {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return file
}

// Get URL params
func (d HTTPDump) getParams(params map[string]string) string {
	b := new(bytes.Buffer)
	for key, value := range params {
		fmt.Fprintf(b, "%s=%s&", key, value)
	}
	return strings.Trim(b.String(), "&")
}

// Get request body
func (d HTTPDump) getBody(requestBody io.ReadCloser) string {
	body, err := ioutil.ReadAll(requestBody)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		return ""
	}

	return string(body)
}

// Get request headers
func (d HTTPDump) getHeaders(headers http.Header) string {
	b := new(bytes.Buffer)
	for key, value := range headers {
		fmt.Fprintf(b, "%s: %s\n", key, value)
	}
	return b.String()
}

// Set non-200 HTTP resonse codes
func (d HTTPDump) setResponseCode(ctx *web.Context) {
	switch d.respCode {
	case 301, 302:
		ctx.Redirect(d.respCode, d.redirect)
	case 404:
		ctx.NotFound(d.resp)
	case 304:
		ctx.NotModified()
	case 403:
		ctx.Forbidden()
	case 401:
		ctx.Unauthorized()
	default:
		ctx.Abort(d.respCode, d.resp)
	}
}

// GET request handler
func (d HTTPDump) handleGet(ctx *web.Context, val string) string {
	urlString := ctx.Request.RemoteAddr + " - " + ctx.Request.RequestURI
	params := d.getParams(ctx.Params)
	if len(params) > 0 {
		params = "?" + params
	}
	println(urlString + params)
	println(d.getHeaders(ctx.Request.Header))

	if d.respCode != 200 {
		d.setResponseCode(ctx)
	}
	return d.resp
}

// Request handler for anything expected to have a body
func (d HTTPDump) handleBody(ctx *web.Context, val string) string {
	d.handleGet(ctx, val)

	println(d.getBody(ctx.Request.Body))
	ctx.SetHeader("Content-Type", "application/json;charset=utf-8", true)

	if d.respCode != 200 {
		d.setResponseCode(ctx)
	}
	return d.resp
}

// CORS config
func (d HTTPDump) handleCORS(ctx *web.Context, val string) string {
	ctx.SetHeader("Access-Control-Allow-Origin", "*", true)
	ctx.SetHeader("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS", true)
	ctx.SetHeader("Access-Control-Max-Age", "86400", true)
	ctx.SetHeader("Access-Control-Allow-Credentials", "true", true)
	ctx.SetHeader("Access-Control-Allow-Headers", "*", true)

	return ""
}

func main() {
	respPtr := flag.String("response", "", "The HTTP response body to return")
	respFilePtr := flag.String("response-file", "", "The file path containing an HTTP response body to return")
	respCodePtr := flag.Int("response-code", 200, "The HTTP response code to return")
	redirectPtr := flag.String("redirect", "", "URL to redirect user's to")
	tlsPtr := flag.Bool("tls", false, "Enable TLS (must include tlskey and tlscert")
	tlsKeyPtr := flag.String("tls-key", "", "Path to TLS key file")
	tlsCertPtr := flag.String("tls-cert", "", "Path to TLS certificate file")
	ipPtr := flag.String("ip", "0.0.0.0", "IP to host the web service on (default: 0.0.0.0)")
	portPtr := flag.String("port", "9999", "TCP port to host the web service on (default: 9999)")

	flag.Parse()

	// Error out if missing TLS key or cert
	if *tlsPtr && (*tlsKeyPtr == "" || *tlsCertPtr == "") {
		println("Usage:")
		flag.PrintDefaults()
		println("")
		log.Fatal("Error: Unable to use TLS, the tlskey or tlscert flags are missing")
	}

	// Error out if using 301/302 without a redirect URL
	if (*respCodePtr == 301 || *respCodePtr == 302) && *redirectPtr == "" {
		println("Usage:")
		flag.PrintDefaults()
		println("")
		log.Fatal("Error: Must specify redirect URL with -redirect")
	}

	// Set resp code
	respCode := 200
	if *redirectPtr != "" {
		respCode = 302
	} else {
		respCode = *respCodePtr
	}

	d := HTTPDump{
		resp:     *respPtr,
		respFile: *respFilePtr,
		respCode: respCode,
		redirect: *redirectPtr,
		tls:      *tlsPtr,
		tlsKey:   *tlsKeyPtr,
		tlsCert:  *tlsCertPtr,
		ip:       *ipPtr,
		port:     *portPtr,
	}
	d.start()
}
