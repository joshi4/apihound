package apihound

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

type Catch struct {
	// APIHoundURL defaults to cloud endpoint
	APIHoundURL string
	// SampleRate must be between 0.0 and 1.0 incluscive.
	// SampleRate defaults to 0.01
	SampleRate float64
}

type catch struct {
	handler http.Handler
	config  *Catch
}

func (c *Catch) CatchRequest(origin http.Handler) http.Handler {
	if c == nil {
		return origin
	}

	return &catch{
		handler: origin,
		config:  c,
	}
}

type throw struct {
	Headers     []string
	Cookies     []string
	QueryParams []string
	Host        string
	Method      string
	// TODO: what is diff between RequestURI and Path
	Path string
	// Only works for JSON body for now
	Body map[string]interface{}
	Err  string
}

func (c *catch) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req == nil || !c.shouldSample() {
		c.handler.ServeHTTP(rw, req)
		return
	}

	t := &throw{
		Method:      req.Method,
		Headers:     getHeaders(req),
		Cookies:     getCookies(req),
		QueryParams: getQueryParams(req),
		Host:        req.Host,
		Path:        req.URL.Path,
		Body:        make(map[string]interface{}),
	}
	defer func() {
		// TODO: make it concurrent
		c.Throw(t)
	}()

	if t.Method == "" {
		t.Method = "GET"
	}

	buf := new(bytes.Buffer)
	if req.Body != nil {
		req.Body = ioutil.NopCloser(io.TeeReader(req.Body, buf))
	}
	c.handler.ServeHTTP(rw, req)

	if t.Method == "GET" {
		return
	}

	// decode buf into the map
	fmt.Println("Body:", string(buf.Bytes()))
	if err := json.Unmarshal(buf.Bytes(), &t.Body); err != nil {
		t.Err = err.Error()
	}
}

func (c *catch) Throw(t *throw) {
	body, err := json.Marshal(t)
	if err != nil {
		return
	}

	res, err := http.Post(c.config.APIHoundURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return
	}
	defer res.Body.Close()

	bs, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	log.Println(string(bs))
}

func (c *catch) shouldSample() bool {
	return true
}

func getHeaders(req *http.Request) []string {
	var hs []string
	for k, _ := range req.Header {
		hs = append(hs, k)
	}
	return hs
}

func getCookies(req *http.Request) []string {
	var cs []string
	for _, c := range req.Cookies() {
		cs = append(cs, c.Name)
	}
	return cs
}

func getQueryParams(req *http.Request) []string {
	var params []string
	for p, _ := range req.URL.Query() {
		params = append(params, p)
	}
	return params
}
