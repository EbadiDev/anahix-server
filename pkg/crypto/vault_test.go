package crypto

import (
	"testing"
)

func TestVaultEncryptionDecryption(t *testing.T) {
	vault, err := NewVault("test-secret-key-for-anahix-vault-12345")
	if err != nil {
		t.Fatalf("Failed to initialize vault: %v", err)
	}

	testCases := []string{
		"https://gemini.google.com/redeem?code=PROMO_18M_XYZ987654",
		"user:password123:backup_code_abc",
		"sk-proj-super-secret-api-key-test",
		"فارسی تست رمزگذاری اطلاعات کاربر",
	}

	for _, original := range testCases {
		encrypted, err := vault.Encrypt(original)
		if err != nil {
			t.Fatalf("Encryption failed for %s: %v", original, err)
		}

		if encrypted == original {
			t.Fatalf("Ciphertext equals plaintext for %s", original)
		}

		decrypted, err := vault.Decrypt(encrypted)
		if err != nil {
			t.Fatalf("Decryption failed for %s: %v", original, err)
		}

		if decrypted != original {
			t.Fatalf("Decrypted string %s does not match original %s", decrypted, original)
		}
	}
}

func TestVaultTamperDetection(t *testing.T) {
	vault, err := NewVault("secret-key")
	if err != nil {
		t.Fatalf("Failed to initialize vault: %v", err)
	}

	encrypted, err := vault.Encrypt("sensitive-data")
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	// Tamper with base64 ciphertext
	tampered := "A" + encrypted[1:]
	_, err = vault.Decrypt(tampered)
	if err == nil {
		t.Fatal("Expected error on tampered ciphertext, but got nil")
	}
}
