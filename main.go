package main

import (
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

var hopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te",
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
}

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "Application address")
	flag.Parse()

	handler := &proxy{}

	log.Println("Starting proxy server on", *addr)
	if err := http.ListenAndServe(*addr, handler); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

type proxy struct {
}

func (this *proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(r.RemoteAddr, " ", r.Method, " ", r.URL)

	if r.URL.Scheme != "http" && r.URL.Scheme != "htts" {
		msg := "Unsupported scheme" + r.URL.Scheme
		http.Error(w, msg, http.StatusBadRequest)
		log.Println(msg)
		return
	}

	client := &http.Client{}

	r.RequestURI = ""
	deleteHopHeaders(r.Header)

	if clientIp, _, err := net.SplitHostPort(r.RemoteAddr); err != nil {
		appendHostToXForwardHeader(r.Header, clientIp)
	}

	resp, err := client.Do(r)
	if err != nil {
		http.Error(w, "Server Error", http.StatusInternalServerError)
		log.Fatal("ServeHTTP:", err)
	}

	defer resp.Body.Close()

	log.Println(r.RemoteAddr, " ", resp.Status)

	deleteHopHeaders(resp.Header)
	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func deleteHopHeaders(header http.Header) {
	for _, val := range hopHeaders {
		header.Del(val)
	}
}

func appendHostToXForwardHeader(header http.Header, host string) {
	if prior, ok := header["X-Forwarded-For"]; ok {
		host = strings.Join(prior, ", ") + ", " + host
	}
	header.Set("X-Forwarded-For", host)
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
