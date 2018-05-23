package apihound

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCatchAndThrow(t *testing.T) {
	c := &Catch{
		APIHoundURL: "https://postman-echo.com/post",
	}

	ts := httptest.NewServer(c.CatchRequest(http.HandlerFunc(okHandler)))
	defer ts.Close()
	cl := ts.Client()

	var body = struct {
		Key1 string
		Key2 []string
		Key3 int
	}{
		Key1: "key1",
		Key2: []string{"a", "b"},
		Key3: 0,
	}

	bs, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}

	// TODO: test cookie
	res, err := cl.Post(ts.URL, "application/json", bytes.NewReader(bs))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	bs, err = ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(bs) != "Hello, client\n" {
		t.Errorf("unexpected response body: %s", string(bs))
	}
}

func okHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(w, "Hello, client")
}
