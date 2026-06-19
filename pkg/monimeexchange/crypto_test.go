package monimeexchange

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateTestKey(t *testing.T) ([]byte, *rsa.PrivateKey) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	return pemBytes, key
}

func TestNewCrypto_PKCS1(t *testing.T) {
	pemBytes, _ := generateTestKey(t)
	c, err := NewCrypto(pemBytes)
	require.NoError(t, err)
	require.NotNil(t, c)
}

func TestNewCrypto_InvalidPEM(t *testing.T) {
	_, err := NewCrypto([]byte("not a pem"))
	assert.Error(t, err)
}

func TestNewCrypto_Empty(t *testing.T) {
	_, err := NewCrypto([]byte{})
	assert.Error(t, err)
}

func TestDecryptRequest_RoundTrip(t *testing.T) {
	_, privKey := generateTestKey(t)
	c := &Crypto{privateKey: privKey}

	// Simulate Monime encrypting a payload
	aesKey := make([]byte, 16)
	_, err := rand.Read(aesKey)
	require.NoError(t, err)

	payload := ExchangePayload{
		CurrentPage: "mp_test",
	}
	payload.Global.SessionID = "sess-123"
	payload.Global.SubscriberID = "sub-456"

	payloadJSON, err := json.Marshal(payload)
	require.NoError(t, err)

	encryptedData, err := encryptAESGCM(payloadJSON, aesKey)
	require.NoError(t, err)

	encryptedKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, &privKey.PublicKey, aesKey, nil)
	require.NoError(t, err)

	req := EncryptedRequest{
		EncryptedAesKey:       base64.StdEncoding.EncodeToString(encryptedKey),
		EncryptedExchangeData: base64.StdEncoding.EncodeToString(encryptedData),
	}

	decrypted, returnedKey, err := c.DecryptRequest(req)
	require.NoError(t, err)
	assert.Equal(t, "sess-123", decrypted.Global.SessionID)
	assert.Equal(t, "sub-456", decrypted.Global.SubscriberID)
	assert.Equal(t, "mp_test", decrypted.CurrentPage)
	assert.Equal(t, aesKey, returnedKey)
}

func TestEncryptResponse_RoundTrip(t *testing.T) {
	_, privKey := generateTestKey(t)
	c := &Crypto{privateKey: privKey}

	aesKey := make([]byte, 16)
	_, err := rand.Read(aesKey)
	require.NoError(t, err)

	resp := NavigateResponse{
		Action: "navigate",
		PageID: "mp_result",
		PageData: map[string]interface{}{
			"message": "Success!",
		},
	}

	encrypted, err := c.EncryptResponse(resp, aesKey)
	require.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	// Decrypt to verify
	encryptedBytes, err := base64.StdEncoding.DecodeString(encrypted)
	require.NoError(t, err)

	plaintext, err := decryptAESGCM(encryptedBytes, aesKey)
	require.NoError(t, err)

	var decoded NavigateResponse
	err = json.Unmarshal(plaintext, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "navigate", decoded.Action)
	assert.Equal(t, "mp_result", decoded.PageID)
	assert.Equal(t, "Success!", decoded.PageData["message"])
}

func TestStopResponse(t *testing.T) {
	_, privKey := generateTestKey(t)
	c := &Crypto{privateKey: privKey}

	aesKey := make([]byte, 16)
	_, err := rand.Read(aesKey)
	require.NoError(t, err)

	resp := StopResponse{
		Action:  "stop",
		Message: "Something went wrong.",
	}

	encrypted, err := c.EncryptResponse(resp, aesKey)
	require.NoError(t, err)

	encryptedBytes, err := base64.StdEncoding.DecodeString(encrypted)
	require.NoError(t, err)

	plaintext, err := decryptAESGCM(encryptedBytes, aesKey)
	require.NoError(t, err)

	var decoded StopResponse
	err = json.Unmarshal(plaintext, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "stop", decoded.Action)
	assert.Equal(t, "Something went wrong.", decoded.Message)
}
