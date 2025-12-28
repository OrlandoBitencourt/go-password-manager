package crypto

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func TestNewService(t *testing.T) {
	service := NewService()
	if service == nil {
		t.Fatal("NewService() returned nil")
	}
}

func TestGenerateSalt(t *testing.T) {
	service := NewService()

	t.Run("generates salt of correct length", func(t *testing.T) {
		salt, err := service.GenerateSalt()
		if err != nil {
			t.Fatalf("GenerateSalt() failed: %v", err)
		}
		if len(salt) != SaltLength {
			t.Errorf("expected salt length %d, got %d", SaltLength, len(salt))
		}
	})

	t.Run("generates unique salts", func(t *testing.T) {
		salt1, err := service.GenerateSalt()
		if err != nil {
			t.Fatalf("GenerateSalt() failed: %v", err)
		}
		salt2, err := service.GenerateSalt()
		if err != nil {
			t.Fatalf("GenerateSalt() failed: %v", err)
		}

		if bytes.Equal(salt1, salt2) {
			t.Error("GenerateSalt() returned identical salts")
		}
	})

	t.Run("generates non-zero salts", func(t *testing.T) {
		salt, err := service.GenerateSalt()
		if err != nil {
			t.Fatalf("GenerateSalt() failed: %v", err)
		}

		allZero := true
		for _, b := range salt {
			if b != 0 {
				allZero = false
				break
			}
		}
		if allZero {
			t.Error("GenerateSalt() returned all-zero salt")
		}
	})
}

func TestDeriveKey(t *testing.T) {
	service := NewService()

	t.Run("derives key successfully", func(t *testing.T) {
		password := "mysecurepassword"
		salt := make([]byte, SaltLength)
		_, _ = rand.Read(salt)

		key, err := service.DeriveKey(password, salt)
		if err != nil {
			t.Fatalf("DeriveKey() failed: %v", err)
		}
		if len(key) != Argon2KeyLen {
			t.Errorf("expected key length %d, got %d", Argon2KeyLen, len(key))
		}
	})

	t.Run("derives consistent keys", func(t *testing.T) {
		password := "mysecurepassword"
		salt := make([]byte, SaltLength)
		_, _ = rand.Read(salt)

		key1, err := service.DeriveKey(password, salt)
		if err != nil {
			t.Fatalf("DeriveKey() failed: %v", err)
		}
		key2, err := service.DeriveKey(password, salt)
		if err != nil {
			t.Fatalf("DeriveKey() failed: %v", err)
		}

		if !bytes.Equal(key1, key2) {
			t.Error("DeriveKey() returned different keys for same password and salt")
		}
	})

	t.Run("derives different keys for different passwords", func(t *testing.T) {
		salt := make([]byte, SaltLength)
		_, _ = rand.Read(salt)

		key1, err := service.DeriveKey("password1", salt)
		if err != nil {
			t.Fatalf("DeriveKey() failed: %v", err)
		}
		key2, err := service.DeriveKey("password2", salt)
		if err != nil {
			t.Fatalf("DeriveKey() failed: %v", err)
		}

		if bytes.Equal(key1, key2) {
			t.Error("DeriveKey() returned same keys for different passwords")
		}
	})

	t.Run("derives different keys for different salts", func(t *testing.T) {
		password := "mysecurepassword"
		salt1 := make([]byte, SaltLength)
		salt2 := make([]byte, SaltLength)
		_, _ = rand.Read(salt1)
		_, _ = rand.Read(salt2)

		key1, err := service.DeriveKey(password, salt1)
		if err != nil {
			t.Fatalf("DeriveKey() failed: %v", err)
		}
		key2, err := service.DeriveKey(password, salt2)
		if err != nil {
			t.Fatalf("DeriveKey() failed: %v", err)
		}

		if bytes.Equal(key1, key2) {
			t.Error("DeriveKey() returned same keys for different salts")
		}
	})

	t.Run("returns error for empty password", func(t *testing.T) {
		salt := make([]byte, SaltLength)
		_, _ = rand.Read(salt)

		_, err := service.DeriveKey("", salt)
		if err == nil {
			t.Error("DeriveKey() should return error for empty password")
		}
		expectedMsg := "password cannot be empty"
		if err.Error() != expectedMsg {
			t.Errorf("expected error %q, got %q", expectedMsg, err.Error())
		}
	})

	t.Run("returns error for empty salt", func(t *testing.T) {
		_, err := service.DeriveKey("password", []byte{})
		if err == nil {
			t.Error("DeriveKey() should return error for empty salt")
		}
		expectedMsg := "salt cannot be empty"
		if err.Error() != expectedMsg {
			t.Errorf("expected error %q, got %q", expectedMsg, err.Error())
		}
	})

	t.Run("returns error for nil salt", func(t *testing.T) {
		_, err := service.DeriveKey("password", nil)
		if err == nil {
			t.Error("DeriveKey() should return error for nil salt")
		}
	})
}

func TestEncrypt(t *testing.T) {
	service := NewService()

	t.Run("encrypts successfully", func(t *testing.T) {
		key := make([]byte, Argon2KeyLen)
		_, _ = rand.Read(key)
		plaintext := []byte("secret data")

		nonce, ciphertext, err := service.Encrypt(plaintext, key)
		if err != nil {
			t.Fatalf("Encrypt() failed: %v", err)
		}
		if len(nonce) != NonceSize {
			t.Errorf("expected nonce size %d, got %d", NonceSize, len(nonce))
		}
		if len(ciphertext) == 0 {
			t.Error("Encrypt() returned empty ciphertext")
		}
		if bytes.Equal(plaintext, ciphertext) {
			t.Error("ciphertext should not equal plaintext")
		}
	})

	t.Run("generates unique nonces", func(t *testing.T) {
		key := make([]byte, Argon2KeyLen)
		_, _ = rand.Read(key)
		plaintext := []byte("secret data")

		nonce1, _, err := service.Encrypt(plaintext, key)
		if err != nil {
			t.Fatalf("Encrypt() failed: %v", err)
		}
		nonce2, _, err := service.Encrypt(plaintext, key)
		if err != nil {
			t.Fatalf("Encrypt() failed: %v", err)
		}

		if bytes.Equal(nonce1, nonce2) {
			t.Error("Encrypt() generated identical nonces")
		}
	})

	t.Run("encrypts empty plaintext", func(t *testing.T) {
		key := make([]byte, Argon2KeyLen)
		_, _ = rand.Read(key)
		plaintext := []byte("")

		nonce, ciphertext, err := service.Encrypt(plaintext, key)
		if err != nil {
			t.Fatalf("Encrypt() failed: %v", err)
		}
		if len(nonce) != NonceSize {
			t.Errorf("expected nonce size %d, got %d", NonceSize, len(nonce))
		}
		if len(ciphertext) == 0 {
			t.Error("Encrypt() should produce ciphertext even for empty plaintext (due to auth tag)")
		}
	})

	t.Run("returns error for invalid key length", func(t *testing.T) {
		key := make([]byte, 16) // Wrong size
		plaintext := []byte("secret data")

		_, _, err := service.Encrypt(plaintext, key)
		if err == nil {
			t.Error("Encrypt() should return error for invalid key length")
		}
	})

	t.Run("returns error for nil key", func(t *testing.T) {
		plaintext := []byte("secret data")

		_, _, err := service.Encrypt(plaintext, nil)
		if err == nil {
			t.Error("Encrypt() should return error for nil key")
		}
	})

	t.Run("encrypts large plaintext", func(t *testing.T) {
		key := make([]byte, Argon2KeyLen)
		_, _ = rand.Read(key)
		plaintext := make([]byte, 1024*1024) // 1MB
		_, _ = rand.Read(plaintext)

		nonce, ciphertext, err := service.Encrypt(plaintext, key)
		if err != nil {
			t.Fatalf("Encrypt() failed: %v", err)
		}
		if len(nonce) != NonceSize {
			t.Errorf("expected nonce size %d, got %d", NonceSize, len(nonce))
		}
		if len(ciphertext) < len(plaintext) {
			t.Error("ciphertext should be at least as long as plaintext (with auth tag)")
		}
	})
}

func TestDecrypt(t *testing.T) {
	service := NewService()

	t.Run("decrypts successfully", func(t *testing.T) {
		key := make([]byte, Argon2KeyLen)
		_, _ = rand.Read(key)
		plaintext := []byte("secret data")

		nonce, ciphertext, err := service.Encrypt(plaintext, key)
		if err != nil {
			t.Fatalf("Encrypt() failed: %v", err)
		}

		decrypted, err := service.Decrypt(nonce, ciphertext, key)
		if err != nil {
			t.Fatalf("Decrypt() failed: %v", err)
		}
		if !bytes.Equal(plaintext, decrypted) {
			t.Errorf("expected %q, got %q", plaintext, decrypted)
		}
	})

	t.Run("decrypts empty plaintext", func(t *testing.T) {
		key := make([]byte, Argon2KeyLen)
		_, _ = rand.Read(key)
		plaintext := []byte("")

		nonce, ciphertext, err := service.Encrypt(plaintext, key)
		if err != nil {
			t.Fatalf("Encrypt() failed: %v", err)
		}

		decrypted, err := service.Decrypt(nonce, ciphertext, key)
		if err != nil {
			t.Fatalf("Decrypt() failed: %v", err)
		}
		if !bytes.Equal(plaintext, decrypted) {
			t.Errorf("expected empty plaintext, got %q", decrypted)
		}
	})

	t.Run("decrypts large plaintext", func(t *testing.T) {
		key := make([]byte, Argon2KeyLen)
		_, _ = rand.Read(key)
		plaintext := make([]byte, 1024*1024) // 1MB
		_, _ = rand.Read(plaintext)

		nonce, ciphertext, err := service.Encrypt(plaintext, key)
		if err != nil {
			t.Fatalf("Encrypt() failed: %v", err)
		}

		decrypted, err := service.Decrypt(nonce, ciphertext, key)
		if err != nil {
			t.Fatalf("Decrypt() failed: %v", err)
		}
		if !bytes.Equal(plaintext, decrypted) {
			t.Error("decrypted data does not match original plaintext")
		}
	})

	t.Run("returns error for wrong key", func(t *testing.T) {
		key1 := make([]byte, Argon2KeyLen)
		key2 := make([]byte, Argon2KeyLen)
		_, _ = rand.Read(key1)
		_, _ = rand.Read(key2)
		plaintext := []byte("secret data")

		nonce, ciphertext, err := service.Encrypt(plaintext, key1)
		if err != nil {
			t.Fatalf("Encrypt() failed: %v", err)
		}

		_, err = service.Decrypt(nonce, ciphertext, key2)
		if err == nil {
			t.Error("Decrypt() should fail with wrong key")
		}
	})

	t.Run("returns error for invalid nonce size", func(t *testing.T) {
		key := make([]byte, Argon2KeyLen)
		_, _ = rand.Read(key)
		plaintext := []byte("secret data")

		_, ciphertext, err := service.Encrypt(plaintext, key)
		if err != nil {
			t.Fatalf("Encrypt() failed: %v", err)
		}

		wrongNonce := make([]byte, 8) // Wrong size
		_, err = service.Decrypt(wrongNonce, ciphertext, key)
		if err == nil {
			t.Error("Decrypt() should return error for invalid nonce size")
		}
	})

	t.Run("returns error for invalid key length", func(t *testing.T) {
		key := make([]byte, Argon2KeyLen)
		_, _ = rand.Read(key)
		plaintext := []byte("secret data")

		nonce, ciphertext, err := service.Encrypt(plaintext, key)
		if err != nil {
			t.Fatalf("Encrypt() failed: %v", err)
		}

		wrongKey := make([]byte, 16) // Wrong size
		_, err = service.Decrypt(nonce, ciphertext, wrongKey)
		if err == nil {
			t.Error("Decrypt() should return error for invalid key length")
		}
	})

	t.Run("returns error for tampered ciphertext", func(t *testing.T) {
		key := make([]byte, Argon2KeyLen)
		_, _ = rand.Read(key)
		plaintext := []byte("secret data")

		nonce, ciphertext, err := service.Encrypt(plaintext, key)
		if err != nil {
			t.Fatalf("Encrypt() failed: %v", err)
		}

		// Tamper with ciphertext
		if len(ciphertext) > 0 {
			ciphertext[0] ^= 0xFF
		}

		_, err = service.Decrypt(nonce, ciphertext, key)
		if err == nil {
			t.Error("Decrypt() should fail for tampered ciphertext")
		}
	})

	t.Run("returns error for nil key", func(t *testing.T) {
		nonce := make([]byte, NonceSize)
		ciphertext := []byte("fake ciphertext")

		_, err := service.Decrypt(nonce, ciphertext, nil)
		if err == nil {
			t.Error("Decrypt() should return error for nil key")
		}
	})
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	service := NewService()

	testCases := []struct {
		name      string
		plaintext []byte
	}{
		{"empty", []byte("")},
		{"short", []byte("a")},
		{"typical", []byte("This is a secret password: P@ssw0rd123!")},
		{"unicode", []byte("こんにちは世界 🔐 🔑")},
		{"binary", []byte{0x00, 0xFF, 0x01, 0xFE, 0x02, 0xFD}},
		{"large", make([]byte, 10000)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			key := make([]byte, Argon2KeyLen)
			_, _ = rand.Read(key)
			if tc.name == "large" {
				_, _ = rand.Read(tc.plaintext)
			}

			nonce, ciphertext, err := service.Encrypt(tc.plaintext, key)
			if err != nil {
				t.Fatalf("Encrypt() failed: %v", err)
			}

			decrypted, err := service.Decrypt(nonce, ciphertext, key)
			if err != nil {
				t.Fatalf("Decrypt() failed: %v", err)
			}

			if !bytes.Equal(tc.plaintext, decrypted) {
				t.Errorf("round trip failed: expected %v, got %v", tc.plaintext, decrypted)
			}
		})
	}
}

func TestFullCryptoWorkflow(t *testing.T) {
	service := NewService()

	// Generate salt
	salt, err := service.GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt() failed: %v", err)
	}

	// Derive key from password
	password := "my-master-password-123"
	key, err := service.DeriveKey(password, salt)
	if err != nil {
		t.Fatalf("DeriveKey() failed: %v", err)
	}

	// Encrypt data
	plaintext := []byte(`{"site":"example.com","username":"user@example.com","password":"P@ssw0rd"}`)
	nonce, ciphertext, err := service.Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt() failed: %v", err)
	}

	// Decrypt data with correct password
	decryptedKey, err := service.DeriveKey(password, salt)
	if err != nil {
		t.Fatalf("DeriveKey() failed: %v", err)
	}
	decrypted, err := service.Decrypt(nonce, ciphertext, decryptedKey)
	if err != nil {
		t.Fatalf("Decrypt() failed: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("full workflow failed: expected %q, got %q", plaintext, decrypted)
	}

	// Try to decrypt with wrong password
	wrongKey, err := service.DeriveKey("wrong-password", salt)
	if err != nil {
		t.Fatalf("DeriveKey() failed: %v", err)
	}
	_, err = service.Decrypt(nonce, ciphertext, wrongKey)
	if err == nil {
		t.Error("Decrypt() should fail with wrong password")
	}
}

// Benchmark tests
func BenchmarkGenerateSalt(b *testing.B) {
	service := NewService()
	for i := 0; i < b.N; i++ {
		_, _ = service.GenerateSalt()
	}
}

func BenchmarkDeriveKey(b *testing.B) {
	service := NewService()
	password := "mysecurepassword"
	salt := make([]byte, SaltLength)
	_, _ = rand.Read(salt)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.DeriveKey(password, salt)
	}
}

func BenchmarkEncrypt(b *testing.B) {
	service := NewService()
	key := make([]byte, Argon2KeyLen)
	_, _ = rand.Read(key)
	plaintext := []byte("This is a secret password that needs to be encrypted")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = service.Encrypt(plaintext, key)
	}
}

func BenchmarkDecrypt(b *testing.B) {
	service := NewService()
	key := make([]byte, Argon2KeyLen)
	_, _ = rand.Read(key)
	plaintext := []byte("This is a secret password that needs to be encrypted")
	nonce, ciphertext, _ := service.Encrypt(plaintext, key)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.Decrypt(nonce, ciphertext, key)
	}
}
