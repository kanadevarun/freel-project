package crypto

import (
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	key := "my-secret-encryption-key-for-test"
	plain := `{"client_id":"123","client_secret":"supersecret"}`

	encrypted, err := Encrypt(plain, key)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}
	if encrypted == plain {
		t.Fatalf("Encrypted ciphertext matches plaintext")
	}

	decrypted, err := Decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}
	if decrypted != plain {
		t.Fatalf("Decrypted text %q != original %q", decrypted, plain)
	}
}

func TestDecryptRow90(t *testing.T) {
	cipherText := "CkMMJjooxkm8sI9UgG89yXZNk5OFCygXflbnA5EbGusWK+x8wFxLvZuz2Of6cCcZKYGjLz0AAmTvRuxosLTSXzVes89tYcx+ExBtynZQh4mVtVvZxg=="
	
	fallbackKey := "LogisticsHQ_Carrier_Secret_Encryption_Key_32B!"
	decFallback, errFallback := Decrypt(cipherText, fallbackKey)
	t.Logf("Fallback key decryption result: %q, err: %v", decFallback, errFallback)

	envKey := "YRYWADPpt3d1rNQoenZPtXYpTFkiG6WqO9cHV9BoOaM="
	decEnv, errEnv := Decrypt(cipherText, envKey)
	t.Logf("Env key decryption result: %q, err: %v", decEnv, errEnv)

	envKeyWithComment := "YRYWADPpt3d1rNQoenZPtXYpTFkiG6WqO9cHV9BoOaM= # openssl rand -base64 32"
	decComment, errComment := Decrypt(cipherText, envKeyWithComment)
	t.Logf("Env key with comment decryption result: %q, err: %v", decComment, errComment)
}

func TestCryptoErrorCases(t *testing.T) {
	// Missing key
	_, err := Encrypt("test", "")
	if err == nil {
		t.Fatal("Expected error for empty encryption key")
	}
	_, err = Decrypt("test", "")
	if err == nil {
		t.Fatal("Expected error for empty decryption key")
	}

	// Ciphertext too short
	_, err = Decrypt("YQ==", "some-key")
	if err == nil {
		t.Fatal("Expected error for short ciphertext")
	}

	// Malformed base64
	_, err = Decrypt("!!!not-base64!!!", "some-key")
	if err == nil {
		t.Fatal("Expected error for invalid base64")
	}

	// Tampered ciphertext
	validCipher, _ := Encrypt("hello world", "test-key")
	tampered := validCipher[:len(validCipher)-4] + "AAAA"
	_, err = Decrypt(tampered, "test-key")
	if err == nil {
		t.Fatal("Expected error for tampered ciphertext")
	}
}
