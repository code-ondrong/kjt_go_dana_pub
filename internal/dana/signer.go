package dana

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strings"
	"time"
)

// Signer menangani pembuatan dan verifikasi signature DANA
type Signer struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	clientID   string
}

// NewSigner membuat Signer baru dari PEM key strings
func NewSigner(clientID, privateKeyPEM, publicKeyPEM string) (*Signer, error) {
	s := &Signer{clientID: clientID}

	if privateKeyPEM != "" {
		priv, err := parsePrivateKey(privateKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("parse private key: %w", err)
		}
		s.privateKey = priv
	}

	if publicKeyPEM != "" {
		pub, err := parsePublicKey(publicKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("parse public key: %w", err)
		}
		s.publicKey = pub
	}

	return s, nil
}

// SignRequest membuat signature untuk request ke DANA
// Format signature: BASE64(RSA_SHA256(method + ":" + path + ":" + timestamp + ":" + BASE64(SHA256(body))))
func (s *Signer) SignRequest(method, path string, body interface{}) (string, string, error) {
	if s.privateKey == nil {
		return "", "", fmt.Errorf("private key tidak tersedia")
	}

	timestamp := time.Now().Format("2006-01-02T15:04:05+07:00")

	// Marshal body ke JSON
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return "", "", fmt.Errorf("marshal body: %w", err)
	}

	// Hash body dengan SHA256 lalu encode ke base64
	bodyHash := sha256.Sum256(bodyBytes)
	bodyHashHex := fmt.Sprintf("%x", bodyHash)
	bodyHashLower := strings.ToLower(bodyHashHex)

	// Buat string yang akan di-sign
	// Format: HTTPMethod:RelativeURL:Timestamp:LowercaseHexBodyHash
	stringToSign := fmt.Sprintf("%s:%s:%s:%s",
		strings.ToUpper(method),
		path,
		timestamp,
		bodyHashLower,
	)

	// Sign dengan RSA-SHA256
	hash := sha256.Sum256([]byte(stringToSign))
	signature, err := rsa.SignPKCS1v15(rand.Reader, s.privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", "", fmt.Errorf("sign request: %w", err)
	}

	signatureB64 := base64.StdEncoding.EncodeToString(signature)
	return signatureB64, timestamp, nil
}

// BuildAuthHeader membuat header Authorization untuk request DANA
func (s *Signer) BuildAuthHeader(signature, timestamp string) string {
	return fmt.Sprintf("DANA ClientId=%s,Timestamp=%s,Signature=%s",
		s.clientID, timestamp, signature)
}

// ============================================================
// HELPER FUNCTIONS
// ============================================================

func parsePrivateKey(pemStr string) (*rsa.PrivateKey, error) {
	// Normalize PEM
	pemStr = normalizePEM(pemStr, "RSA PRIVATE KEY")

	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		// Coba PKCS8
		pemStr = normalizePEM(pemStr, "PRIVATE KEY")
		block, _ = pem.Decode([]byte(pemStr))
		if block == nil {
			return nil, fmt.Errorf("gagal decode PEM block")
		}
	}

	// Coba parse sebagai PKCS1
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	// Coba parse sebagai PKCS8
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("bukan RSA private key")
	}
	return rsaKey, nil
}

func parsePublicKey(pemStr string) (*rsa.PublicKey, error) {
	pemStr = normalizePEM(pemStr, "PUBLIC KEY")

	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("gagal decode PEM block")
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}

	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("bukan RSA public key")
	}
	return rsaKey, nil
}

func normalizePEM(key, keyType string) string {
	key = strings.TrimSpace(key)
	if strings.HasPrefix(key, "-----") {
		return key
	}

	header := fmt.Sprintf("-----BEGIN %s-----", keyType)
	footer := fmt.Sprintf("-----END %s-----", keyType)

	var lines []string
	for len(key) > 64 {
		lines = append(lines, key[:64])
		key = key[64:]
	}
	if len(key) > 0 {
		lines = append(lines, key)
	}

	return header + "\n" + strings.Join(lines, "\n") + "\n" + footer
}
