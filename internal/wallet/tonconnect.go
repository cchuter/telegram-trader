package wallet

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/nacl/box"
)

const (
	// Default bridge URL for TON Connect
	DefaultBridgeURL = "https://bridge.tonapi.io/bridge"

	// Default manifest URL (should be hosted by the app)
	DefaultManifestURL = "https://raw.githubusercontent.com/ton-connect/demo-dapp/master/public/tonconnect-manifest.json"

	// Protocol version
	ProtocolVersion = "2"

	// Default TTL for bridge messages (5 minutes)
	DefaultMessageTTL = 300
)

// TonConnectSession represents an active TonConnect session
type TonConnectSession struct {
	ClientID      string    // Hex-encoded client public key
	PrivateKey    []byte    // Client's private key (32 bytes)
	PublicKey     []byte    // Client's public key (32 bytes)
	WalletID      string    // Wallet's public key (hex-encoded)
	WalletAddress string    // Wallet address (UQ... or EQ...)
	BridgeURL     string    // Bridge server URL
	ManifestURL   string    // App manifest URL
	CreatedAt     time.Time
	Connected     bool

	// Internal fields for message handling
	mu            sync.RWMutex
	stopChan      chan struct{}
	eventCallback func(address string, err error)
}

// ConnectRequest represents the initial connection request
type ConnectRequest struct {
	ManifestURL string   `json:"manifestUrl"`
	Items       []Item   `json:"items"`
}

// Item represents a requested permission item
type Item struct {
	Name string `json:"name"`
}

// ConnectEvent represents the wallet's connection response
type ConnectEvent struct {
	Event   string  `json:"event"`
	ID      int64   `json:"id"`
	Payload Payload `json:"payload"`
}

// Payload contains the wallet connection information
type Payload struct {
	Items []PayloadItem `json:"items"`
	Device DeviceInfo   `json:"device"`
}

// PayloadItem contains account information
type PayloadItem struct {
	Name    string  `json:"name"`
	Address string  `json:"address"`
	Network string  `json:"network"`
	PublicKey string `json:"publicKey"`
	WalletStateInit string `json:"walletStateInit"`
}

// DeviceInfo contains device information
type DeviceInfo struct {
	Platform  string `json:"platform"`
	AppName   string `json:"appName"`
	AppVersion string `json:"appVersion"`
}

// BridgeMessage represents a message received from the bridge
type BridgeMessage struct {
	From    string `json:"from"`
	Message string `json:"message"`
}

// TonConnector handles TonConnect protocol operations
type TonConnector struct {
	bridgeURL   string
	manifestURL string
	httpClient  *http.Client
}

// NewTonConnector creates a new TonConnect handler
func NewTonConnector(bridgeURL, manifestURL string) *TonConnector {
	if bridgeURL == "" {
		bridgeURL = DefaultBridgeURL
	}
	if manifestURL == "" {
		manifestURL = DefaultManifestURL
	}

	return &TonConnector{
		bridgeURL:   bridgeURL,
		manifestURL: manifestURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateSession creates a new TonConnect session with generated keypair
func (tc *TonConnector) CreateSession() (*TonConnectSession, error) {
	// Generate X25519 keypair for the client
	publicKey, privateKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate keypair: %w", err)
	}

	// Client ID is the hex-encoded public key
	clientID := hex.EncodeToString(publicKey[:])

	session := &TonConnectSession{
		ClientID:    clientID,
		PrivateKey:  privateKey[:],
		PublicKey:   publicKey[:],
		BridgeURL:   tc.bridgeURL,
		ManifestURL: tc.manifestURL,
		CreatedAt:   time.Now(),
		Connected:   false,
		stopChan:    make(chan struct{}),
	}

	return session, nil
}

// GenerateConnectionURL generates the TonConnect universal link
// Format: https://<wallet-url>?v=2&id=<hex(A)>&r=<urlsafe(json(ConnectRequest))>&ret=back
func (tc *TonConnector) GenerateConnectionURL(session *TonConnectSession, walletUniversalURL string) (string, error) {
	// Create connect request
	request := ConnectRequest{
		ManifestURL: tc.manifestURL,
		Items: []Item{
			{Name: "ton_addr"},
			{Name: "ton_proof"},
		},
	}

	// Marshal to JSON
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal connect request: %w", err)
	}

	// URL-safe base64 encode
	requestEncoded := base64.RawURLEncoding.EncodeToString(requestJSON)

	// Build connection URL
	u, err := url.Parse(walletUniversalURL)
	if err != nil {
		return "", fmt.Errorf("invalid wallet URL: %w", err)
	}

	q := u.Query()
	q.Set("v", ProtocolVersion)
	q.Set("id", session.ClientID)
	q.Set("r", requestEncoded)
	q.Set("ret", "back")
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// GenerateQRCodeURL generates a QR code URL for the connection
// This returns a tc:// deep link that works with any TonConnect-compatible wallet
func (tc *TonConnector) GenerateQRCodeURL(session *TonConnectSession) (string, error) {
	// Create connect request
	request := ConnectRequest{
		ManifestURL: tc.manifestURL,
		Items: []Item{
			{Name: "ton_addr"},
			{Name: "ton_proof"},
		},
	}

	// Marshal to JSON
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal connect request: %w", err)
	}

	// URL-safe base64 encode
	requestEncoded := base64.RawURLEncoding.EncodeToString(requestJSON)

	// Build tc:// deep link
	deepLink := fmt.Sprintf("tc://connect?v=%s&id=%s&r=%s&ret=back",
		ProtocolVersion,
		session.ClientID,
		url.QueryEscape(requestEncoded),
	)

	return deepLink, nil
}

// GenerateTonKeeperURL generates a TonKeeper-specific connection URL
func (tc *TonConnector) GenerateTonKeeperURL(session *TonConnectSession) (string, error) {
	return tc.GenerateConnectionURL(session, "https://app.tonkeeper.com/ton-connect")
}

// ListenForConnection starts listening for wallet connection via bridge
func (tc *TonConnector) ListenForConnection(ctx context.Context, session *TonConnectSession, callback func(address string, err error)) error {
	session.mu.Lock()
	session.eventCallback = callback
	session.mu.Unlock()

	// Start SSE connection to bridge
	go tc.bridgeSSEListener(ctx, session)

	return nil
}

// bridgeSSEListener listens for Server-Sent Events from the bridge
func (tc *TonConnector) bridgeSSEListener(ctx context.Context, session *TonConnectSession) {
	bridgeURL := fmt.Sprintf("%s/events?client_id=%s", session.BridgeURL, session.ClientID)

	// Create request with SSE headers
	req, err := http.NewRequestWithContext(ctx, "GET", bridgeURL, nil)
	if err != nil {
		tc.notifyCallback(session, "", fmt.Errorf("failed to create bridge request: %w", err))
		return
	}

	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	// Make request
	resp, err := tc.httpClient.Do(req)
	if err != nil {
		tc.notifyCallback(session, "", fmt.Errorf("failed to connect to bridge: %w", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		tc.notifyCallback(session, "", fmt.Errorf("bridge returned status %d", resp.StatusCode))
		return
	}

	// Read SSE stream
	reader := &sseReader{reader: resp.Body}

	for {
		select {
		case <-ctx.Done():
			return
		case <-session.stopChan:
			return
		default:
			event, data, err := reader.ReadEvent()
			if err != nil {
				if err != io.EOF {
					tc.notifyCallback(session, "", fmt.Errorf("error reading SSE: %w", err))
				}
				return
			}

			// Handle message events
			if event == "message" && data != "heartbeat" {
				tc.handleBridgeMessage(session, data)
			}
		}
	}
}

// handleBridgeMessage processes a message from the bridge
func (tc *TonConnector) handleBridgeMessage(session *TonConnectSession, data string) {
	var bridgeMsg BridgeMessage
	if err := json.Unmarshal([]byte(data), &bridgeMsg); err != nil {
		tc.notifyCallback(session, "", fmt.Errorf("failed to parse bridge message: %w", err))
		return
	}

	// Decode base64 message
	encryptedData, err := base64.StdEncoding.DecodeString(bridgeMsg.Message)
	if err != nil {
		tc.notifyCallback(session, "", fmt.Errorf("failed to decode message: %w", err))
		return
	}

	// Decrypt message using NaCl box
	// The encrypted message format is: nonce (24 bytes) + ciphertext
	if len(encryptedData) < 24 {
		tc.notifyCallback(session, "", fmt.Errorf("encrypted message too short"))
		return
	}

	var nonce [24]byte
	copy(nonce[:], encryptedData[:24])
	ciphertext := encryptedData[24:]

	// Decode wallet's public key (sender)
	walletPubKey, err := hex.DecodeString(bridgeMsg.From)
	if err != nil {
		tc.notifyCallback(session, "", fmt.Errorf("failed to decode wallet public key: %w", err))
		return
	}

	var walletPubKey32 [32]byte
	copy(walletPubKey32[:], walletPubKey)

	var clientPrivKey32 [32]byte
	copy(clientPrivKey32[:], session.PrivateKey)

	// Decrypt using box.Open
	decrypted, ok := box.Open(nil, ciphertext, &nonce, &walletPubKey32, &clientPrivKey32)
	if !ok {
		tc.notifyCallback(session, "", fmt.Errorf("failed to decrypt message"))
		return
	}

	// Parse connect event
	var connectEvent ConnectEvent
	if err := json.Unmarshal(decrypted, &connectEvent); err != nil {
		tc.notifyCallback(session, "", fmt.Errorf("failed to parse connect event: %w", err))
		return
	}

	// Extract wallet address
	if connectEvent.Event == "connect" && len(connectEvent.Payload.Items) > 0 {
		session.mu.Lock()
		session.WalletAddress = connectEvent.Payload.Items[0].Address
		session.WalletID = bridgeMsg.From
		session.Connected = true
		session.mu.Unlock()

		tc.notifyCallback(session, connectEvent.Payload.Items[0].Address, nil)
	}
}

// notifyCallback safely calls the event callback
func (tc *TonConnector) notifyCallback(session *TonConnectSession, address string, err error) {
	session.mu.RLock()
	callback := session.eventCallback
	session.mu.RUnlock()

	if callback != nil {
		callback(address, err)
	}
}

// StopListening stops listening for connection events
func (tc *TonConnector) StopListening(session *TonConnectSession) {
	close(session.stopChan)
}

// sseReader reads Server-Sent Events from an io.Reader
type sseReader struct {
	reader io.Reader
	buffer []byte
}

// ReadEvent reads the next SSE event
func (r *sseReader) ReadEvent() (event string, data string, err error) {
	event = "message" // default event type

	for {
		line, err := r.readLine()
		if err != nil {
			return "", "", err
		}

		// Empty line signals end of event
		if len(line) == 0 {
			return event, data, nil
		}

		// Parse field
		colonIdx := strings.Index(line, ":")
		if colonIdx == -1 {
			continue
		}

		field := line[:colonIdx]
		value := strings.TrimSpace(line[colonIdx+1:])

		switch field {
		case "event":
			event = value
		case "data":
			if data != "" {
				data += "\n"
			}
			data += value
		}
	}
}

// readLine reads a single line from the SSE stream
func (r *sseReader) readLine() (string, error) {
	var line []byte

	for {
		b := make([]byte, 1)
		n, err := r.reader.Read(b)
		if err != nil {
			return "", err
		}
		if n == 0 {
			continue
		}

		if b[0] == '\n' {
			return string(line), nil
		}
		if b[0] != '\r' {
			line = append(line, b[0])
		}
	}
}

// IsConnected returns whether the session is connected
func (session *TonConnectSession) IsConnected() bool {
	session.mu.RLock()
	defer session.mu.RUnlock()
	return session.Connected
}

// GetWalletAddress returns the connected wallet address
func (session *TonConnectSession) GetWalletAddress() string {
	session.mu.RLock()
	defer session.mu.RUnlock()
	return session.WalletAddress
}
