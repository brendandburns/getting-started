package main

import (
	"example"
	"fmt"
	"flag"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/golang/glog"
)

var (
	dir          = flag.String("www", "", "The directory to serve files from.")
	addr         = flag.String("addr", ":8080", "The address to listen on.  Default ':8080'")
	clientID     = flag.String("client-id", "", "OAuth 2.0 Client ID.  If non-empty, overrides --clientid_file")
	clientIDFile = flag.String("client-id-file", "clientid.dat",
		"Name of a file containing just the project's OAuth 2.0 Client ID from https://developers.google.com/console.")
	clientSecret     = flag.String("secret", "", "OAuth 2.0 Client Secret.  If non-empty, overrides --secret_file")
	clientSecretFile = flag.String("secret-file", "clientsecret.dat",
		"Name of a file containing just the project's OAuth 2.0 Client Secret from https://developers.google.com/console")

	client *http.Client
)

func loadFileOrString(val, file string) string {
	if len(val) > 0 {
		return val
	}
	data, err := ioutil.ReadFile(file)
	if err != nil {
		glog.Fatalf("Failed to load data from %s: %v", file, err)
	}
	return string(data)
}

func tokenHandler(res http.ResponseWriter, req *http.Request) {
     code := req.URL.Query()["code"]
     if len(code) == 0 {
     	res.WriteHeader(500)
	res.Write([]byte("Missing expected parameter"))
	return
	}

     id := loadFileOrString(*clientID, *clientIDFile)
     secret := loadFileOrString(*clientSecret, *clientSecretFile)
	config, ctx := example.NewClientConfigAndContext(id, secret)

     _, port, err := net.SplitHostPort(req.Host)
     if err != nil {
     	res.WriteHeader(500)
	response := fmt.Sprintf("Error: %v", err)
	res.Write([]byte(response))
	return
	}

     // This is the URL to return to
     config.RedirectURL = fmt.Sprintf("http://localhost:%s/token", port)

	client = example.NewOAuthClient(ctx, config, code[0])

	glog.Infof("Created client: %v", client)
	
}

func authHandler(res http.ResponseWriter, req *http.Request) {
     id := loadFileOrString(*clientID, *clientIDFile)
     secret := loadFileOrString(*clientSecret, *clientSecretFile)

     config, _ := example.NewClientConfigAndContext(id, secret)

     _, port, err := net.SplitHostPort(req.Host)
     if err != nil {
     	res.WriteHeader(500)
	response := fmt.Sprintf("Error: %v", err)
	res.Write([]byte(response))
	return
	}

     // This is the URL to return to
     config.RedirectURL = fmt.Sprintf("http://localhost:%s/token", port)

     // This is the URL to send people to
     redirectURL := example.SendTokenRequest(config)
     http.Redirect(res, req, redirectURL, 302)
}

func main() {
	flag.Parse()

	http.Handle("/", http.FileServer(http.Dir(*dir)))
	http.HandleFunc("/auth", authHandler)
	http.HandleFunc("/token", tokenHandler)
	http.ListenAndServe(*addr, nil)




}
