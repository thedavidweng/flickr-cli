package flickr

import (
	"context"
	"fmt"
	"net"
	"net/http"
)

// Package-level seams for testing.
var (
	oauthListen = net.Listen
	oauthDial   = net.Dial
)

// OAuthFlow orchestrates the OAuth 1.0a three-step authorization dance.
// It encapsulates the protocol logic — binding a callback port, running
// a callback HTTP server, detecting the outbound IP — that previously
// lived in the CLI adapter.
type OAuthFlow struct {
	Client *Client
}

// LocalhostFlow runs the localhost-based OAuth authorization flow.
// It starts a callback server first to determine the actual port, then
// registers that port as the OAuth callback with Flickr. The server binds
// to 0.0.0.0 so browsers on other machines can also complete the authorization.
//
// onAuth is called with the authorization URL so the caller can display it
// to the user (e.g. print to stderr, open a browser).
func (f *OAuthFlow) LocalhostFlow(ctx context.Context, perms string, port int, onAuth func(authURL, callbackURL string)) (*RequestTokenResponse, string, error) {
	serverIP := detectOutboundIP()
	ln, err := oauthListen("tcp", fmt.Sprintf("%s:%d", serverIP, port))
	if err != nil {
		// Fallback to localhost if specific IP fails
		serverIP = "127.0.0.1"
		ln, err = oauthListen("tcp", fmt.Sprintf("%s:%d", serverIP, port))
		if err != nil {
			return nil, "", fmt.Errorf("listening on port %d: %w", port, err)
		}
	}
	defer func() { _ = ln.Close() }()

	addr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		return nil, "", fmt.Errorf("unexpected listener address type")
	}
	callbackURL := fmt.Sprintf("http://%s:%d", serverIP, addr.Port)

	reqToken, err := f.Client.RequestToken(ctx, callbackURL)
	if err != nil {
		return nil, "", err
	}

	authURL := f.Client.AuthorizationURL(reqToken.Token, perms)
	if onAuth != nil {
		onAuth(authURL, callbackURL)
	}

	verifier, err := waitForCallback(ctx, ln)
	if err != nil {
		return nil, "", err
	}

	return reqToken, verifier, nil
}

// OOBRequestToken obtains a request token with "oob" callback and returns
// the authorization URL. The caller must obtain the verifier from the user
// and exchange it via Client.AccessToken.
func (f *OAuthFlow) OOBRequestToken(ctx context.Context, perms string) (string, *RequestTokenResponse, error) {
	reqToken, err := f.Client.RequestToken(ctx, "oob")
	if err != nil {
		return "", nil, err
	}
	authURL := f.Client.AuthorizationURL(reqToken.Token, perms)
	return authURL, reqToken, nil
}

// waitForCallback starts an HTTP server on the given listener and waits
// for an OAuth callback containing the verifier.
func waitForCallback(ctx context.Context, ln net.Listener) (string, error) {
	verifierCh := make(chan string, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		v := req.URL.Query().Get("oauth_verifier")
		if v == "" {
			// Browsers may request favicon.ico or other resources;
			// respond with 200 instead of error to avoid deadlock.
			_, _ = fmt.Fprintf(w, "Waiting for authorization...")
			return
		}
		_, _ = fmt.Fprintf(w, "Authorization successful! You can close this window.")
		verifierCh <- v
	})

	srv := &http.Server{Handler: mux}
	go func() {
		_ = srv.Serve(ln) // http.ErrServerClosed is expected
	}()

	select {
	case verifier := <-verifierCh:
		_ = srv.Shutdown(ctx)
		return verifier, nil
	case <-ctx.Done():
		_ = srv.Shutdown(ctx)
		return "", ctx.Err()
	}
}

// detectOutboundIP returns the preferred outbound IP of this machine.
func detectOutboundIP() string {
	conn, err := oauthDial("udp", "8.8.8.8:80")
	if err != nil {
		return "localhost"
	}
	defer func() { _ = conn.Close() }()
	udpAddr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok {
		return "localhost"
	}
	return udpAddr.IP.String()
}
