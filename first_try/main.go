// Gazer is a simple reverse proxy.
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func main() {

	http.HandleFunc("/", handler)

	err := http.ListenAndServeTLS(":10443", "server.pem", "server.key", nil)
	if err != nil {
		logf("http.ListenAndServeTLS: %v", err.Error())
	}
}

func handler(w http.ResponseWriter, r *http.Request) {

	// --- r
	logf("r: %v", r)

	defer r.Body.Close()
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logf("err: ioutil.ReadAll: %v", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if len(b) > 0 {
		logf("b: %v", string(b))
	}

	nr := bytes.NewReader(b)

	// --- req
	client := &http.Client{}

	req, err := http.NewRequest(r.Method, "https://example.com"+r.URL.Path, nr)

	for k, v := range r.Header {
		for _, vv := range v {

			if k == "Referer" {
				vv = strings.Replace(vv, "127.0.0.1", "example.com", 1)
			}

			if k == "Origin" {
				vv = strings.Replace(vv, "127.0.0.1", "example.com", 1)
			}

			req.Header.Add(k, vv)
		}
	}

	// --- resp
	resp, err := client.Do(req)
	if err != nil {
		logf("err: http.Get: %v", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logf("err: ioutil.ReadAll: %v", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	logf("req: %v", req)

	logf("resp: %v", resp)

	// --- Duplicate and return
	for k, v := range resp.Header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}

	_, err = w.Write(body)
	if err != nil {
		logf("w.Write err: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func logf(f string, v ...interface{}) {
	s := fmt.Sprintf(f, v...)
	log.Printf(s)
}
