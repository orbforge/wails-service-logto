package wailslogto

import (
	"time"

	"github.com/logto-io/go/v2/client"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// Config is the configuration for the LogTo authentication Service
type Config struct {
	// RedirectAddresses provides and ordered list of addresses to use for short-lived callback server
	// started to handle callbacks from LogTo during authentication operations.
	// The first URI that can be successfully listened on on the current machine will be used.
	// The URIs should be in the format "http://<address>:<port>/<path>" and must be HTTP
	// addresses that resolves to the current machine (i.e. http://127.0.0.1:1234/auth/callback)
	// Ensure all URIs listed are configured as "Redirect URIs" and "Post sign-out redirect URIs"
	// in your LogTo application settings.
	RedirectAddresses []RedirectURIs

	// LogToConfig is the authentication configuration for LogTo, including Endpoint and App ID
	LogToConfig *client.LogtoConfig

	// WindowConfig is the configuration for the popup window that will be opened for the auth flow
	WindowOptions application.WebviewWindowOptions

	// AuthTimeout is the duration to wait for the user to complete the auth flow before timing out
	// and closing the window (optional)
	AuthTimeout time.Duration
}

// RedirectURIs is a pair of URIs to use for sign in and sign out redirects
type RedirectURIs struct {
	// SignIn is the URI to redirect to after sign in
	SignIn string

	// SignOut is the URI to redirect to after sign out
	SignOut string
}

// SignInURIs returns the ordered list of URIs to use for sign in redirects
func (c *Config) SignInURIs() []string {
	uris := []string{}
	for _, redirect := range c.RedirectAddresses {
		uris = append(uris, redirect.SignIn)
	}
	return uris
}

// SignOutURIs returns the ordered list of URIs to use for sign out redirects
func (c *Config) SignOutURIs() []string {
	uris := []string{}
	for _, redirect := range c.RedirectAddresses {
		uris = append(uris, redirect.SignOut)
	}
	return uris
}
