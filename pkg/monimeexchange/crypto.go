package monimeexchange

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
)

// EncryptedRequest is the inbound body from Monime.
type EncryptedRequest struct {
	EncryptedAesKey         string `json:"encryptedAesKey"`
	EncryptedExchangeData   string `json:"encryptedExchangeData"`
}

// ExchangePayload is the decrypted request payload.
type ExchangePayload struct {
	Global struct {
		SessionID            string `json:"sessionId"`
		NetworkName          string `json:"networkName"`
		SubscriberID         string `json:"subscriberId"`
		SubscriberMsisdn     string `json:"subscriberMsisdn"`
		NetworkProviderName  string `json:"networkProviderName"`
	} `json:"global"`
	CurrentPage   string                 `json:"currentPage"`
	FlowData      map[string]interface{} `json:"flowData"`
	ExportedData  map[string]interface{} `json:"exportedData"`
	SessionContext map[string]interface{} `json:"sessionContext"`
}

// NavigateResponse directs Monime to another page.
type NavigateResponse struct {
	Action   string                 `json:"action"`
	PageID   string                 `json:"pageId"`
	PageData map[string]interface{} `json:"pageData,omitempty"`
}

// StopResponse ends the USSD session.
type StopResponse struct {
	Action  string `json:"action"`
	Message string `json:"message"`
}

// Crypto handles Monime AES-GCM + RSA-OAEP exchange encryption.
type Crypto struct {
	privateKey *rsa.PrivateKey
}

// NewCrypto loads an RSA private key from PEM bytes.
func NewCrypto(pemBytes []byte) (*Crypto, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("invalid PEM block")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parse private key: %w", err)
		}
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("key is not RSA private key")
	}
	return &Crypto{privateKey: rsaKey}, nil
}

// DecryptRequest unwraps and decrypts an inbound exchange request.
func (c *Crypto) DecryptRequest(req EncryptedRequest) (*ExchangePayload, []byte, error) {
	encryptedKey, err := base64.StdEncoding.DecodeString(req.EncryptedAesKey)
	if err != nil {
		return nil, nil, fmt.Errorf("decode aes key: %w", err)
	}
	encryptedData, err := base64.StdEncoding.DecodeString(req.EncryptedExchangeData)
	if err != nil {
		return nil, nil, fmt.Errorf("decode exchange data: %w", err)
	}

	aesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, c.privateKey, encryptedKey, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("rsa decrypt: %w", err)
	}
	if len(aesKey) != 16 {
		return nil, nil, fmt.Errorf("unexpected aes key length: %d", len(aesKey))
	}

	plaintext, err := decryptAESGCM(encryptedData, aesKey)
	if err != nil {
		return nil, nil, err
	}

	var payload ExchangePayload
	if err := json.Unmarshal(plaintext, &payload); err != nil {
		return nil, nil, fmt.Errorf("unmarshal payload: %w", err)
	}
	return &payload, aesKey, nil
}

// EncryptResponse encrypts a navigate/stop response and returns base64 text/plain blob.
func (c *Crypto) EncryptResponse(response interface{}, aesKey []byte) (string, error) {
	responseJSON, err := json.Marshal(response)
	if err != nil {
		return "", err
	}
	blob, err := encryptAESGCM(responseJSON, aesKey)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(blob), nil
}

func decryptAESGCM(encryptedData, aesKey []byte) ([]byte, error) {
	if len(encryptedData) < 1+12+16 {
		return nil, errors.New("encrypted data too short")
	}
	ivLen := int(encryptedData[0])
	if ivLen != 12 {
		return nil, fmt.Errorf("unexpected iv length: %d", ivLen)
	}
	iv := encryptedData[1 : 1+ivLen]
	ciphertext := encryptedData[1+ivLen:]

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return gcm.Open(nil, iv, ciphertext, nil)
}

func encryptAESGCM(plaintext, aesKey []byte) ([]byte, error) {
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	iv := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	encrypted := gcm.Seal(nil, iv, plaintext, nil)
	blob := make([]byte, 1+len(iv)+len(encrypted))
	blob[0] = byte(len(iv))
	copy(blob[1:], iv)
	copy(blob[1+len(iv):], encrypted)
	return blob, nil
}
