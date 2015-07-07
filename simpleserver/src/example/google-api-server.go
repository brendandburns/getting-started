package example

// Heavily copied from https://github.com/google/google-api-go-client/blob/master/examples/main.go
import (
	"fmt"
	"net/http"
	"time"

        "github.com/golang/glog"
	container "github.com/google/google-api-go-client/container/v1"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func NewClientConfigAndContext(clientID, clientSecret string) (*oauth2.Config, context.Context) {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  "urn:ietf:wg:oauth:2.0:oob",
		Scopes:       []string{container.CloudPlatformScope},
	}

	ctx := context.Background()

	return config, ctx
}

func NewOAuthClient(ctx context.Context, config *oauth2.Config, code string) *http.Client {
	token := exchangeToken(ctx, config, code)

	return config.Client(ctx, token)
}

func SendTokenRequest(config *oauth2.Config) string {
	randState := fmt.Sprintf("st%d", time.Now().UnixNano())
	authURL := config.AuthCodeURL(randState)
	return authURL
}

func exchangeToken(ctx context.Context, config *oauth2.Config, code string) *oauth2.Token {
	token, err := config.Exchange(ctx, code)
	if err != nil {
		glog.Fatalf("Token exchange error: %v", err)
	}
	return token
}

