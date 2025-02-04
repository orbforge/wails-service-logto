<h1 align="center">Wails Service Logto</h1>

<p align="center">
  <img src="./wails-service-logto-logos.png" height="280" />
</p>

<p align="center">
  Wails V3 Service for Logto authentication
</p>

<p align="center">
  <a href="https://github.com/orbforge/wails-service-logto/blob/main/LICENSE">
    <img alt="License-MIT" src="https://img.shields.io/badge/License-MIT-yellow.svg"/>
  </a>
  <a href="https://github.com/orbforge/wails-service-logto/releases">
    <img alt="GitHub release (latest by date including pre-releases)" src="https://img.shields.io/github/v/release/orbforge/wails-service-logto?include_prereleases&sort=semver">
  </a>
  <a href="https://pkg.go.dev/github.com/orbforge/wails-service-logto">
    <img alt="Go Reference" src="https://pkg.go.dev/badge/github.com/orbforge/wails-logto-service.svg">
  </a>
</p>

## Overview

This package provides a [Wails V3 Service](https://v3alpha.wails.io/learn/services) to easily add [Logto](https://logto.io) authentication to your application.

## Installation

```bash
go get github.com/orbforge/wails-service-logto
```

## Usage

### Import
```go
import wailslogto "github.com/orbforge/wails-service-logto"
```

### Configure
```go
application.Options{
    Services: []application.Service{
        application.NewService(
            wailslogto.New(
                &wailslogto.Config{
                    RedirectAddresses: []logto.RedirectURIs{
                        // ensure these addresses are added to your configuration in Logto
                        {
                            SignIn:  "http://127.0.0.1:5447/auth/callback",
                            SignOut: "http://127.0.0.1:5447/auth/callback",
                        },
                        // optionally, you can add fallback listener addresses
                        {
                            SignIn:  "http://127.0.0.1:5448/auth/callback",
                            SignOut: "http://127.0.0.1:5448/auth/callback",
                        },
                    },
                    LogToConfig: &client.LogtoConfig{
                        Endpoint: "https://your-logto.endpoint.net/",
                        AppId:    "your-app-id",
                    },
                    WindowOptions: application.WebviewWindowOptions{
                        Title:     "Example Sign In",
                    },
                    // optional maximum time to wait for the user to sign in before
                    // closing the sign-in window and callback listener
                    AuthTimeout: 5 * time.Minute,
                },
            ),
        ),
    },
}
``` 

### Use from JS via Wails Bindings

```ts
import { SignIn } from '../../bindings/github.com/orbforge/wails-service-logto';

// Sign in
SignIn().then(() => {
    console.log('Signed in');
}).catch((err) => {
    console.error('Failed to sign in', err);
});
```

### API Reference

See the [Go Reference](https://pkg.go.dev/github.com/orbforge/wails-service-logto) for detailed API documentation.

## License

[License MIT](./LICENSE)
