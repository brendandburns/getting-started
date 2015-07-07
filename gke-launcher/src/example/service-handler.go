package example

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/golang/glog"
	container "github.com/google/google-api-go-client/container/v1"
)

var (
	clientID     = flag.String("client-id", "", "OAuth 2.0 Client ID.  If non-empty, overrides --clientid_file")
	clientIDFile = flag.String("client-id-file", "",
		"Name of a file containing just the project's OAuth 2.0 Client ID from https://developers.google.com/console.")
	clientSecret     = flag.String("secret", "", "OAuth 2.0 Client Secret.  If non-empty, overrides --secret_file")
	clientSecretFile = flag.String("secret-file", "",
		"Name of a file containing just the project's OAuth 2.0 Client Secret from https://developers.google.com/console")
)

type ServiceHandler struct {
	Service         *container.Service
	Delegate        *http.ServeMux
	selectedCluster *container.Cluster
}

// Validates that there is a service impl. and that all keys in the params map are filled in from req
func (s *ServiceHandler) validateClientAndParameters(req *http.Request, params map[string]string) error {
	if s.Service == nil {
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
		data[ix] = charset[val%l]
	}
	return string(data)
}

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

func (s *ServiceHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if s.Service == nil {
		code, found := req.URL.Query()["code"]
		if !found {
			s.authHandler(res, req)
			return
		}
		if err := s.tokenHandler(res, req, code[0]); err != nil {
			res.WriteHeader(500)
			response := fmt.Sprintf("Error: %v", err)
			res.Write([]byte(response))
			return
		}
		req.URL.Query().Del("code")
		req.URL.Query().Del("state")
		http.Redirect(res, req, req.URL.String(), 302)
		return
	}
	s.Delegate.ServeHTTP(res, req)
}

func loadSecretAndID() (secret string, id string) {
	if len(*clientID) != 0 || len(*clientIDFile) != 0 {
		id = loadFileOrString(*clientID, *clientIDFile)
	} else {
		id = "255964991331-b0l3n9c5pqc0u0ijtniv8vls226d3d5j.apps.googleusercontent.com"
	}
	if len(*clientSecret) != 0 || len(*clientSecretFile) != 0 {
		secret = loadFileOrString(*clientSecret, *clientSecretFile)
	} else {
		secret = "BWm6fPAY2gS1jaRT-Xn2y-uT"
	}
	return
}

func (s *ServiceHandler) authHandler(res http.ResponseWriter, req *http.Request) {
	secret, id := loadSecretAndID()
	config, _ := NewClientConfigAndContext(id, secret)
	
	_, port, err := net.SplitHostPort(req.Host)
	if err != nil {
		res.WriteHeader(500)
		response := fmt.Sprintf("Error: %v", err)
		res.Write([]byte(response))
		return
	}

	// This is the URL to return to
	config.RedirectURL = fmt.Sprintf("http://localhost:%s%s", port, req.URL.Path)

	// This is the URL to send people to
	redirectURL := SendTokenRequest(config)
	http.Redirect(res, req, redirectURL, 302)
}

func (s *ServiceHandler) tokenHandler(res http.ResponseWriter, req *http.Request, code string) error {
	if len(code) == 0 {
		return errors.New("Invalid code")
	}

	secret, id := loadSecretAndID()
	config, ctx := NewClientConfigAndContext(id, secret)

	_, port, err := net.SplitHostPort(req.Host)
	if err != nil {
		return err
	}

	// This is the URL to return to
	config.RedirectURL = fmt.Sprintf("http://localhost:%s%s", port, req.URL.Path)

	client := NewOAuthClient(ctx, config, code)
	containerService, err := container.New(client)
	if err != nil {
		return err
	}
	s.Service = containerService
	return nil
}

func (s *ServiceHandler) ListClusterHandler(res http.ResponseWriter, req *http.Request) {
	params := map[string]string{"project": "", "zone": ""}
	if err := s.validateClientAndParameters(req, params); err != nil {
		res.WriteHeader(500)
		res.Write([]byte(fmt.Sprintf("Failed to list: %v", err)))
		return
	}
	list, err := s.Service.Projects.Zones.Clusters.List(params["project"], params["zone"]).Do()
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

func (s *ServiceHandler) CreateClusterHandler(res http.ResponseWriter, req *http.Request) {
	params := map[string]string{"project": "", "zone": "", "cluster": ""}
	if err := s.validateClientAndParameters(req, params); err != nil {
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
	clusterResponse, err := s.Service.Projects.Zones.Clusters.Create(params["project"], params["zone"], createRequest).Do()
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

func (s *ServiceHandler) SelectClusterHandler(res http.ResponseWriter, req *http.Request) {
	params := map[string]string{"project": "", "zone": "", "cluster": ""}
	if err := s.validateClientAndParameters(req, params); err != nil {
		res.WriteHeader(500)
		res.Write([]byte(fmt.Sprintf("Failed to list: %v", err)))
		return
	}
	cluster, err := s.Service.Projects.Zones.Clusters.Get(params["project"], params["zone"], params["cluster"]).Do()
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
	s.selectedCluster = cluster
	res.WriteHeader(200)
	res.Write(data)
}

func (s *ServiceHandler) SelectDirector(req *http.Request) {
     	if s.selectedCluster != nil {
		req.URL.Host = s.selectedCluster.Endpoint
		req.URL.Scheme = "https"
		req.SetBasicAuth(s.selectedCluster.MasterAuth.Username, s.selectedCluster.MasterAuth.Password)
	}
}

func StaticFileHandler(res http.ResponseWriter, req *http.Request) {
	glog.Infof("Looking for: %s", req.URL.Path)
	file := req.URL.Path[1:]
	if len(file) == 0 {
		file = "index.html"
	}
	asset, err := Asset(file)
	if err != nil {
		res.WriteHeader(500)
		res.Write([]byte(err.Error()))
		return
	}
	res.WriteHeader(200)
	res.Write(asset)
}
