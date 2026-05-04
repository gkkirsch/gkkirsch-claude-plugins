package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"time"
)

const (
	keychainService = "director-slack-user"
	keychainAccount = "oauth_token"
	oauthPort       = 9876
	oauthRedirect   = "https://localhost:9876/oauth/callback"
	oauthTimeout    = 5 * time.Minute
)

// User scopes — channel/group listing, history, sending, reactions, search, file uploads.
var oauthScopes = []string{
	"channels:read", "groups:read",
	"channels:history", "groups:history",
	"im:history", "im:read",
	"mpim:history", "mpim:read",
	"chat:write",
	"reactions:write",
	"search:read",
	"users:read", "users:read.email",
	"files:write",
}

// keychainGet reads the OAuth token from the macOS Keychain (or returns "" if absent).
func keychainGet() (string, error) {
	out, err := exec.Command("/usr/bin/security",
		"find-generic-password",
		"-s", keychainService,
		"-a", keychainAccount,
		"-w",
	).Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return "", nil // not found
		}
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// keychainSet writes the token (replacing any existing entry).
func keychainSet(token string) error {
	cmd := exec.Command("/usr/bin/security",
		"add-generic-password",
		"-U", // update if exists
		"-s", keychainService,
		"-a", keychainAccount,
		"-w", token,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("keychain write failed: %s", strings.TrimSpace(string(out)))
	}
	return nil
}

// keychainDelete removes the token entry.
func keychainDelete() error {
	cmd := exec.Command("/usr/bin/security",
		"delete-generic-password",
		"-s", keychainService,
		"-a", keychainAccount,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		s := strings.TrimSpace(string(out))
		if strings.Contains(s, "could not be found") {
			return nil
		}
		return fmt.Errorf("keychain delete failed: %s", s)
	}
	return nil
}

// ensureToken returns a usable OAuth token, kicking off OAuth if absent.
func ensureToken() (string, error) {
	tok, err := keychainGet()
	if err != nil {
		return "", err
	}
	if tok != "" {
		return tok, nil
	}
	return runOAuth(false)
}

// runOAuth performs the full OAuth user-token flow.
// If reauth is true, the existing token is cleared first.
func runOAuth(reauth bool) (string, error) {
	if reauth {
		_ = keychainDelete()
	}
	clientID := getEnv("SLACK_CLIENT_ID", "")
	clientSecret := getEnv("SLACK_CLIENT_SECRET", "")
	if clientID == "" || clientSecret == "" {
		return "", fmt.Errorf("SLACK_CLIENT_ID and SLACK_CLIENT_SECRET must be set in the environment")
	}

	state := randHex(16)
	verifier := randHex(32)
	challenge := pkceChallenge(verifier)

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	srv, err := startOAuthCallback(state, codeCh, errCh)
	if err != nil {
		return "", err
	}
	defer srv.Shutdown(context.Background())

	authURL := buildAuthURL(clientID, state, challenge)
	fmt.Fprintln(stderr(), "Opening browser to authorize Slack...")
	fmt.Fprintln(stderr(), "(if it doesn't open, visit:", authURL+")")
	openBrowser(authURL)

	select {
	case code := <-codeCh:
		return exchangeCode(clientID, clientSecret, code, verifier)
	case err := <-errCh:
		return "", err
	case <-time.After(oauthTimeout):
		return "", fmt.Errorf("OAuth timed out after %v", oauthTimeout)
	}
}

func buildAuthURL(clientID, state, challenge string) string {
	q := url.Values{}
	q.Set("client_id", clientID)
	q.Set("user_scope", strings.Join(oauthScopes, ","))
	q.Set("redirect_uri", oauthRedirect)
	q.Set("state", state)
	q.Set("code_challenge", challenge)
	q.Set("code_challenge_method", "S256")
	return "https://slack.com/oauth/v2/authorize?" + q.Encode()
}

// startOAuthCallback brings up an HTTPS server on oauthPort with a self-signed cert.
func startOAuthCallback(expectedState string, codeCh chan<- string, errCh chan<- error) (*http.Server, error) {
	cert, err := selfSignedCert()
	if err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != expectedState {
			http.Error(w, "state mismatch", http.StatusBadRequest)
			errCh <- errors.New("OAuth state mismatch")
			return
		}
		if errParam := r.URL.Query().Get("error"); errParam != "" {
			http.Error(w, errParam, http.StatusBadRequest)
			errCh <- fmt.Errorf("OAuth error: %s", errParam)
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "missing code", http.StatusBadRequest)
			errCh <- errors.New("OAuth callback missing code")
			return
		}
		fmt.Fprintln(w, "<html><body style='font-family:system-ui;padding:2em'><h1>Authorized.</h1><p>You can close this tab.</p></body></html>")
		codeCh <- code
	})

	srv := &http.Server{
		Addr:      fmt.Sprintf(":%d", oauthPort),
		Handler:   mux,
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{cert}},
	}
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return nil, fmt.Errorf("OAuth callback listen failed: %w", err)
	}
	tlsLn := tls.NewListener(ln, srv.TLSConfig)
	go func() {
		if err := srv.Serve(tlsLn); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()
	return srv, nil
}

func exchangeCode(clientID, clientSecret, code, verifier string) (string, error) {
	form := url.Values{}
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)
	form.Set("code", code)
	form.Set("code_verifier", verifier)
	form.Set("redirect_uri", oauthRedirect)

	resp, err := http.PostForm("https://slack.com/api/oauth.v2.access", form)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var out struct {
		OK         bool   `json:"ok"`
		Error      string `json:"error"`
		AuthedUser struct {
			AccessToken string `json:"access_token"`
		} `json:"authed_user"`
		Team struct {
			Name string `json:"name"`
			ID   string `json:"id"`
		} `json:"team"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", fmt.Errorf("oauth response parse: %w", err)
	}
	if !out.OK {
		return "", fmt.Errorf("oauth error: %s", out.Error)
	}
	if out.AuthedUser.AccessToken == "" {
		return "", errors.New("oauth response missing user token")
	}
	if err := keychainSet(out.AuthedUser.AccessToken); err != nil {
		return "", err
	}
	fmt.Fprintf(stderr(), "Authorized to %s (%s).\n", out.Team.Name, out.Team.ID)
	return out.AuthedUser.AccessToken, nil
}

func selfSignedCert() (tls.Certificate, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}
	tmpl := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, err
	}
	keyDER, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return tls.Certificate{}, err
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	return tls.X509KeyPair(certPEM, keyPEM)
}

func openBrowser(rawurl string) {
	_ = exec.Command("/usr/bin/open", rawurl).Start()
}

func randHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func pkceChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}
