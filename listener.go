package wailslogto

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
)

// Listener is an HTTP server that listens for a single request on a provided URL (host and path).
// Used to listen for a single callback from an external service for a short-lived operation.
type Listener struct {
	server      *http.Server
	tcpListener net.Listener
	result      chan Result
	ctx         context.Context
	cancel      context.CancelFunc
}

// CallbackHandler is a function that handles an incoming HTTP request and returns an error if one occurred
type CallbackHandler func(r *http.Request) error

// Result is the result of a callback listen operation
type Result struct {
	// Success is true if the callback was received and handled successfully
	Success bool
	// Error is the sign in or listener error that occurred, if not successful
	Error error
}

// StartListener starts an HTTP server that listens for a single request on the provided URL (host and path).
// and returns a Listener instance that can be used to Await the result, as well as the URL that the server is listening on.
func StartListener(handler CallbackHandler, urlOptions ...string) (*Listener, string, error) {
	result := make(chan Result)

	var server *http.Server
	var tcpListener net.Listener
	var lastErr error
	for _, serveUrl := range urlOptions {
		route, err := url.Parse(serveUrl)
		if err != nil {
			return nil, "", err
		}

		listening, err := net.Listen("tcp", route.Host)
		if err != nil {
			// if we can't listen on this address, try the next one
			lastErr = err
			continue
		}

		// start the server on established TCP listener
		router := http.NewServeMux()
		router.HandleFunc(
			fmt.Sprintf("GET %s", route.Path),
			func(w http.ResponseWriter, r *http.Request) {
				err := handler(r)
				result <- Result{
					Success: err == nil,
					Error:   err,
				}
			},
		)

		tcpListener = listening
		server = &http.Server{
			Addr:    route.Host,
			Handler: router,
		}
		go server.Serve(listening)
		// we're done after the first successful server start
		serverCtx, serverCancel := context.WithCancel(context.Background())
		return &Listener{
			server:      server,
			tcpListener: tcpListener,
			result:      result,
			ctx:         serverCtx,
			cancel:      serverCancel,
		}, serveUrl, nil
	}
	return nil, "", lastErr
}

// Await waits for the callback to be received by the listener and returns the success
// result as indicated by the handler, or any error that occurred
func (c *Listener) Await(ctx context.Context) (bool, error) {
	defer c.Close()

	select {
	case result := <-c.result:
		return result.Success, result.Error
	case <-ctx.Done():
		return false, ctx.Err()
	case <-c.ctx.Done():
		return false, c.ctx.Err()
	}
}

// Close stops the listener and cleans up resources
func (c *Listener) Close() {
	c.cancel()
	c.server.Close()
	c.tcpListener.Close()
}
