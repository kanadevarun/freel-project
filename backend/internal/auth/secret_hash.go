package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

// ComputeSecretHash calculates the secret hash required by Cognito User Pools
// when an App Client has a Client Secret configured.
// The formula is: Base64(HMAC_SHA256(ClientSecret, Username + ClientID))
func ComputeSecretHash(clientSecret, username, clientID string) string {
	mac := hmac.New(sha256.New, []byte(clientSecret))
	mac.Write([]byte(username + clientID))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
