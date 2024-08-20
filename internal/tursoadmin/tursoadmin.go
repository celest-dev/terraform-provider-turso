package tursoadmin

import (
	"errors"
	"net/http"

	"github.com/celest-dev/terraform-provider-turso/internal/tursoadmin/internal/httpmodel"
)

// Config is the configuration for a Turso admin client.
type Config struct {
	// Required. The name of the organization.
	OrgName string

	// Required. The API token for the organization.
	APIToken string

	// Optional. The HTTP client to use for API requests.
	//
	// Default: `http.DefaultClient`
	HTTPClient *http.Client

	// Optional. The base URL for API requests.
	//
	// Default: `https://api.turso.tech/v1`
	BaseURL string
}

// Client is a Turso admin client.
type Client struct {
	http    *httpmodel.JSONClient
	baseURL string
	orgName string
}

// NewClient creates a new Turso admin client.
func NewClient(config Config) (*Client, error) {
	if config.OrgName == "" {
		return nil, errors.New("missing organization name")
	}
	if config.APIToken == "" {
		return nil, errors.New("missing API token")
	}
	baseURL := coalesce(config.BaseURL, "https://api.turso.tech/v1")
	return &Client{
		http: httpmodel.NewJSONClient(
			coalesce(config.HTTPClient, http.DefaultClient),
			baseURL,
			map[string]string{
				"Authorization": "Bearer " + config.APIToken,
			},
		),
		baseURL: baseURL,
		orgName: config.OrgName,
	}, nil
}

func coalesce[T comparable](values ...T) T {
	var zero T
	for _, value := range values {
		if value != zero {
			return value
		}
	}
	return zero
}
