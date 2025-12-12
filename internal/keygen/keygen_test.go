package keygen

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEd25519Generator_Generate(t *testing.T) {
	gen := NewEd25519Generator()

	t.Run("generates valid key pair", func(t *testing.T) {
		keyPair, err := gen.Generate("test@example.com")
		require.NoError(t, err)
		require.NotNil(t, keyPair)

		// Verify private key is not empty
		assert.NotEmpty(t, keyPair.PrivateKey)
		// Verify public key is not empty
		assert.NotEmpty(t, keyPair.PublicKey)
		assert.NotEmpty(t, keyPair.PublicKeyStr)
	})

	t.Run("private key is OpenSSH format", func(t *testing.T) {
		keyPair, err := gen.Generate("test@example.com")
		require.NoError(t, err)

		privateKeyStr := string(keyPair.PrivateKey)
		assert.Contains(t, privateKeyStr, "-----BEGIN OPENSSH PRIVATE KEY-----")
		assert.Contains(t, privateKeyStr, "-----END OPENSSH PRIVATE KEY-----")
	})

	t.Run("public key is authorized_keys format", func(t *testing.T) {
		keyPair, err := gen.Generate("test@example.com")
		require.NoError(t, err)

		// Ed25519 public keys start with ssh-ed25519
		assert.True(t, strings.HasPrefix(keyPair.PublicKeyStr, "ssh-ed25519 "))
		// Public key should be base64 encoded (contains AAAA for ed25519)
		assert.Contains(t, keyPair.PublicKeyStr, "AAAA")
	})

	t.Run("comment is included in public key", func(t *testing.T) {
		comment := "gdk:myuser/myrepo"
		keyPair, err := gen.Generate(comment)
		require.NoError(t, err)

		// Comment should be at the end of the public key
		assert.True(t, strings.HasSuffix(strings.TrimSpace(keyPair.PublicKeyStr), comment))
	})

	t.Run("generates unique keys each time", func(t *testing.T) {
		kp1, err := gen.Generate("test1")
		require.NoError(t, err)

		kp2, err := gen.Generate("test2")
		require.NoError(t, err)

		// Keys should be different
		assert.NotEqual(t, kp1.PublicKeyStr, kp2.PublicKeyStr)
		assert.NotEqual(t, string(kp1.PrivateKey), string(kp2.PrivateKey))
	})

	t.Run("handles empty comment", func(t *testing.T) {
		keyPair, err := gen.Generate("")
		require.NoError(t, err)
		assert.NotEmpty(t, keyPair.PublicKeyStr)
	})
}

func TestKeyPair_Validate(t *testing.T) {
	gen := NewEd25519Generator()

	t.Run("valid key pair passes validation", func(t *testing.T) {
		keyPair, err := gen.Generate("test")
		require.NoError(t, err)

		err = keyPair.Validate()
		assert.NoError(t, err)
	})

	t.Run("empty private key fails validation", func(t *testing.T) {
		keyPair := &KeyPair{
			PrivateKey:   nil,
			PublicKey:    []byte("ssh-ed25519 AAAA test"),
			PublicKeyStr: "ssh-ed25519 AAAA test",
		}

		err := keyPair.Validate()
		assert.Error(t, err)
	})

	t.Run("empty public key fails validation", func(t *testing.T) {
		gen := NewEd25519Generator()
		keyPair, _ := gen.Generate("test")
		keyPair.PublicKey = nil

		err := keyPair.Validate()
		assert.Error(t, err)
	})
}
