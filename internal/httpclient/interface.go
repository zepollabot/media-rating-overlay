package common

import (
	"net/http"

	"github.com/gojek/heimdall/v7"
)

// HTTPClient defines the interface for the underlying HTTP client
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
	AddPlugin(p heimdall.Plugin)
}

type ServiceHTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}
