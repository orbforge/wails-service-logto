package wailslogto

import (
	"context"
	"errors"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/logto-io/go/v2/client"
	"github.com/logto-io/go/v2/core"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// Service is a LogTo authentication service
type Service struct {
	config *Config
	client *client.LogtoClient
	ctx    context.Context
	cancel context.CancelFunc
}

// New creates a new LogTo authentication service with the provided configuration
func New(config *Config) *Service {
	serviceCtx, cancel := context.WithCancel(context.Background())
	return &Service{
		config: config,
		client: client.NewLogtoClient(config.LogToConfig, NewStore()),
		ctx:    serviceCtx,
		cancel: cancel,
	}
}

// ServiceName returns the name of the plugin.
func (s *Service) ServiceName() string {
	return "github.com/orbforge/wails-service-logto"
}

// ServiceStartup is called when the Service is loaded. No-op for this service.
func (s *Service) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	return nil
}

// ServiceShutdown is called when the Service is unloaded. Cleanup any active listeners.
func (s *Service) ServiceShutdown() error {
	s.cancel()
	return nil
}

// SignIn initiates a sign in flow with LogTo.
// This will open a new window with the LogTo sign in URL and listen for the callback on completion.
// This function blocks until the user completes the sign in flow or an error occurs, including a configured timeout.
// If the user completes the sign in flow successfully, this function will return true and user sign in status and
// details will be available through this service.
// If the user alredy has an active auth session cookie, this function should complete without ever showing a new window.
func (s *Service) SignIn(opts ...*client.SignInOptions) (bool, error) {
	// start a callback listener
	listener, callbackUrl, err := StartListener(s.client.HandleSignInCallback, s.config.SignInURIs()...)
	if err != nil {
		return false, err
	}

	clientOpts := client.SignInOptions{
		RedirectUri: callbackUrl,
	}
	var options *client.SignInOptions
	if len(opts) > 0 {
		options = opts[0]
	}

	if options != nil {
		if options.RedirectUri != "" {
			clientOpts.RedirectUri = options.RedirectUri
		} 
		if options.Prompt != "" {
			clientOpts.Prompt = options.Prompt
		}
		if options.FirstScreen != "" {
			clientOpts.FirstScreen = options.FirstScreen
		}
		if len(options.Identifiers) > 0 {
			clientOpts.Identifiers = options.Identifiers
		}
		if options.DirectSignIn != nil {
			clientOpts.DirectSignIn = options.DirectSignIn
		}
		if options.LoginHint != "" {
			clientOpts.LoginHint = options.LoginHint
		}
		if len(options.ExtraParams) > 0 {
			clientOpts.ExtraParams = options.ExtraParams
		}
	}

	signinUrl, err := s.client.SignIn(&clientOpts)
	if err != nil {
		return false, err
	}
	closed := atomic.Bool{}
	// open a window with the LogTo sign in URL
	windowOpts := s.config.WindowOptions
	windowOpts.URL = signinUrl
	authWindow := s.newWindow(windowOpts, func() {
		closed.Store(true)
		listener.Close()
	})

	// wait for the callback (or timeout, if configured)
	result, err := listener.Await(s.newAuthCtx())
	if !closed.Load() {
		authWindow.Close()
	}
	return result, err
}

// TryAutoSignIn attempts to sign in the user without showing a window.
// If the user has an active auth session, this function will complete successfully and return true.
// if the user does not have an active session, this function will return false
// and the user will need to sign in normally.
func (s *Service) TryAutoSignIn(timeAllowed string) (bool, error) {
	timeoutDuration, err := time.ParseDuration(timeAllowed)
	if err != nil {
		return false, err
	}
	// start a callback listener
	listener, callbackUrl, err := StartListener(s.client.HandleSignInCallback, s.config.SignInURIs()...)
	if err != nil {
		return false, err
	}

	signinUrl, err := s.client.SignIn(
		&client.SignInOptions{
			RedirectUri: callbackUrl,
		})
	if err != nil {
		return false, err
	}
	closed := atomic.Bool{}
	// open a window with the LogTo sign in URL
	windowOpts := s.config.WindowOptions
	windowOpts.URL = signinUrl
	windowOpts.Hidden = true
	authWindow := s.newWindow(windowOpts, func() {
		closed.Store(true)
		listener.Close()
	})

	// wait for the callback (or timeout)
	ctx, _ := context.WithTimeout(s.newAuthCtx(), timeoutDuration)
	result, err := listener.Await(ctx)
	if !closed.Load() {
		authWindow.Close()
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return false, nil
	}
	return result, err
}

// SignOut initiates a sign out flow with LogTo.
// This will open a new window with the LogTo sign out URL and listen for the callback on completion.
// This function blocks until the user completes the sign out flow or an error occurs, including a configured timeout.
func (s *Service) SignOut() (bool, error) {
	handleSignOutCallback := func(r *http.Request) error {
		// handle the sign out callback, no-op
		return nil
	}

	// start a callback listener
	listener, callbackUrl, err := StartListener(handleSignOutCallback, s.config.SignOutURIs()...)
	if err != nil {
		return false, err
	}

	signoutUrl, err := s.client.SignOut(callbackUrl)
	if err != nil {
		return false, err
	}
	closed := atomic.Bool{}
	// open a window with the LogTo sign in URL
	windowOpts := s.config.WindowOptions
	windowOpts.Hidden = true
	windowOpts.URL = signoutUrl

	authWindow := s.newWindow(windowOpts, func() {
		closed.Store(true)
		listener.Close()
	})

	// wait for the callback (or timeout, if configured)
	success, err := listener.Await(s.newAuthCtx())
	if !closed.Load() {
		authWindow.Close()
	}
	return success, err
}

// IsAuthenticated returns true if the user is signed in
func (s *Service) IsAuthenticated() bool {
	return s.client.IsAuthenticated()
}

// GetIdToken returns the current ID token
func (s *Service) GetIdToken() string {
	return s.client.GetIdToken()
}

// GetAccessToken returns the current access token for the given resource ("" for default)
func (s *Service) GetAccessToken(resource string) (client.AccessToken, error) {
	return s.client.GetAccessToken(resource)
}

// FetchUserInfo returns the user info for the current user
func (s *Service) FetchUserInfo() (core.UserInfoResponse, error) {
	return s.client.FetchUserInfo()
}

// newWindow creates a new webview window with the given URL and visibility, and an onClose callback
func (s *Service) newWindow(conf application.WebviewWindowOptions, onClose func()) *application.WebviewWindow {
	window := application.Get().NewWebviewWindowWithOptions(conf)
	window.OnWindowEvent(events.Common.WindowClosing, func(event *application.WindowEvent) {
		onClose()
	})
	return window
}

// newAuthCtx returns a new context for authentication operations, with a timeout if configured.
// This context is also cancelled if the service is shutdown.
func (s *Service) newAuthCtx() context.Context {
	if s.config.AuthTimeout == 0 {
		return s.ctx
	}
	ctx, _ := context.WithTimeout(s.ctx, s.config.AuthTimeout)
	return ctx
}
