package main

import (
        "crypto/rand"
	"crypto/tls"
	"encoding/json"
	"errors"
	"example"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"

	"github.com/golang/glog"
	container "github.com/google/google-api-go-client/container/v1"
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

	client  *http.Client
	service *container.Service
	selectedCluster *container.Cluster
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
	containerService, err := container.New(client)
	if err != nil {
		res.WriteHeader(500)
		response := fmt.Sprintf("Error: %v", err)
		res.Write([]byte(response))
		return
	}
	service = containerService
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

// Validates that there is a service impl. and that all keys in the params map are filled in from req
func validateClientAndParameters(req *http.Request, params map[string]string) error {
	if service == nil {
		return errors.New("Service object is nil")
	}
	queryParams := req.URL.Query()
	for key := range params {
		value, ok := queryParams[key]
		if !ok || len(value) == 0 || len(value[0]) == 0 {
			return fmt.Errorf("Missing value for parameter: %s", key)
		}
		params[key] = value[0]
	}
	return nil
}

func randomPassword(size int) string {
    const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
    l := byte(len(charset))
    data := make([]byte, size)
    rand.Read(data)
    for ix, val := range data {
        data[ix] = charset[val % l]
    }
    return string(data)
}

func selectClusterHandler(res http.ResponseWriter, req *http.Request) {
     params := map[string]string{"project": "", "zone": "", "cluster": ""}
     	if err := validateClientAndParameters(req, params); err != nil {
		res.WriteHeader(500)
		res.Write([]byte(fmt.Sprintf("Failed to list: %v", err)))
		return
	}
	cluster, err := service.Projects.Zones.Clusters.Get(params["project"], params["zone"], params["cluster"]).Do()
	if err != nil {
		res.WriteHeader(500)
		res.Write([]byte(fmt.Sprintf("Failed to select: %v", err)))
		return
	}
	data, err := json.Marshal(cluster)
	if err != nil {
		res.WriteHeader(500)
		res.Write([]byte(fmt.Sprintf("Failed to select: %v", err)))
		return
	}
	selectedCluster = cluster
	res.WriteHeader(200)
	res.Write(data)
}

func createClusterHandler(res http.ResponseWriter, req *http.Request) {
	params := map[string]string{"project": "", "zone": "", "cluster": ""}
	if err := validateClientAndParameters(req, params); err != nil {
		res.WriteHeader(500)
		res.Write([]byte(fmt.Sprintf("Failed to list: %v", err)))
		return
	}
	cluster := &container.Cluster{
		Name:             params["cluster"],
		InitialNodeCount: 1,
		MasterAuth: &container.MasterAuth{
			    Username: "admin",
			    Password: randomPassword(16),
                },
	}
	createRequest := &container.CreateClusterRequest{cluster}
	clusterResponse, err := service.Projects.Zones.Clusters.Create(params["project"], params["zone"], createRequest).Do()
	if err != nil {
		res.WriteHeader(500)
		res.Write([]byte(fmt.Sprintf("Failed to create: %v", err)))
		return
	}
	data, err := json.Marshal(clusterResponse)
	if err != nil {
		res.WriteHeader(500)
		res.Write([]byte(fmt.Sprintf("Failed to create: %v", err)))
		return
	}
	res.WriteHeader(200)
	res.Write(data)
}

func listClusterHandler(res http.ResponseWriter, req *http.Request) {
	params := map[string]string{"project": "", "zone": ""}
	if err := validateClientAndParameters(req, params); err != nil {
		res.WriteHeader(500)
		res.Write([]byte(fmt.Sprintf("Failed to list: %v", err)))
		return
	}
	list, err := service.Projects.Zones.Clusters.List(params["project"], params["zone"]).Do()
	if err != nil {
		res.WriteHeader(500)
		res.Write([]byte(fmt.Sprintf("Failed to list: %v", err)))
		return
	}
	data, err := json.Marshal(list)
	if err != nil {
		res.WriteHeader(500)
		res.Write([]byte(fmt.Sprintf("Failed to list: %v", err)))
		return
	}
	res.WriteHeader(200)
	res.Write(data)
}

func selectDirector(req *http.Request) {
     req.URL.Host = selectedCluster.Endpoint
     req.URL.Scheme = "https"
     req.SetBasicAuth(selectedCluster.MasterAuth.Username, selectedCluster.MasterAuth.Password)
}

func main() {
	flag.Parse()

	http.Handle("/", http.FileServer(http.Dir(*dir)))
	http.HandleFunc("/auth", authHandler)
	http.HandleFunc("/token", tokenHandler)
	http.HandleFunc("/list", listClusterHandler)
	http.HandleFunc("/create", createClusterHandler)
	http.HandleFunc("/select", selectClusterHandler)

	proxy := &httputil.ReverseProxy{
	      Director: selectDirector,
	      Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}
	http.Handle("/api/", proxy)

	http.ListenAndServe(*addr, nil)

}
