package main

import (
	"crypto/tls"
	"example"
	"flag"
	"net/http"
	"net/http/httputil"

	"github.com/golang/glog"
)

var (
	dir  = flag.String("www", "", "The directory to serve files from.")
	addr = flag.String("addr", ":8080", "The address to listen on.  Default ':8080'")
)

func main() {
	flag.Parse()

	mux := http.NewServeMux()
	//mux.Handle("/", http.FileServer(http.Dir(*dir)))
	//mux.Handle("/", example.StaticFileHandler())
	mux.HandleFunc("/", example.StaticFileHandler)

	serviceHandler := &example.ServiceHandler{Delegate: mux}

	mux.HandleFunc("/list", serviceHandler.ListClusterHandler)
	mux.HandleFunc("/create", serviceHandler.CreateClusterHandler)
	mux.HandleFunc("/select", serviceHandler.SelectClusterHandler)

	proxy := &httputil.ReverseProxy{
		Director:  serviceHandler.SelectDirector,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}
	mux.Handle("/api/", proxy)

	if err := http.ListenAndServe(*addr, serviceHandler); err != nil {
		glog.Fatalf("ListenAndServe: %v", err)
	}

}
