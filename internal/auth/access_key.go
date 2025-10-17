package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

func GenerateAccessKey(userID, secret string) (string, error) {
	if userID == "" {
		return "", errors.New("userID is empty")
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(userID))
	signature := mac.Sum(nil)
	payload := fmt.Sprintf("%s:%s", userID, base64.StdEncoding.EncodeToString(signature))
	return base64.StdEncoding.EncodeToString([]byte(payload)), nil
}

func ParseAccessKey(accessKey, secret string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(accessKey)
	if err != nil {
		return "", fmt.Errorf("decode access key: %w", err)
	}

	parts := strings.Split(string(raw), ":")
	if len(parts) != 2 {
		return "", errors.New("invalid access key format")
	}

	userID := parts[0]
	providedSig, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("decode signature: %w", err)
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(userID))
	expectedSig := mac.Sum(nil)

	if !hmac.Equal(expectedSig, providedSig) {
		return "", errors.New("signature mismatch")
	}

	return userID, nil
}
