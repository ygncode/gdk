package keygen

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"errors"

	"golang.org/x/crypto/ssh"
)

var (
	// ErrKeyGenFailed indicates key generation failed
	ErrKeyGenFailed = errors.New("failed to generate key")

	// ErrInvalidKeyPair indicates the key pair is invalid
	ErrInvalidKeyPair = errors.New("invalid key pair")
)

// KeyPair holds the generated key pair
type KeyPair struct {
	PrivateKey   []byte // OpenSSH format PEM
	PublicKey    []byte // OpenSSH authorized_keys format
	PublicKeyStr string // Human-readable public key string
}

// Validate checks if the key pair is valid
func (kp *KeyPair) Validate() error {
	if len(kp.PrivateKey) == 0 {
		return errors.New("private key is empty")
	}
	if len(kp.PublicKey) == 0 {
		return errors.New("public key is empty")
	}
	return nil
}

// KeyGenerator defines the interface for SSH key generation
type KeyGenerator interface {
	Generate(comment string) (*KeyPair, error)
}

// Ed25519Generator generates Ed25519 SSH keys
type Ed25519Generator struct{}

// NewEd25519Generator creates a new Ed25519 key generator
func NewEd25519Generator() *Ed25519Generator {
	return &Ed25519Generator{}
}

// Generate creates a new Ed25519 key pair
func (g *Ed25519Generator) Generate(comment string) (*KeyPair, error) {
	// Generate Ed25519 key pair
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, errors.Join(ErrKeyGenFailed, err)
	}

	// Convert to SSH format
	sshPubKey, err := ssh.NewPublicKey(pubKey)
	if err != nil {
		return nil, errors.Join(ErrKeyGenFailed, err)
	}

	// Marshal private key to OpenSSH format
	privateKeyPEM, err := ssh.MarshalPrivateKey(privKey, comment)
	if err != nil {
		return nil, errors.Join(ErrKeyGenFailed, err)
	}

	// Encode to PEM format with proper headers
	privateKeyBytes := pem.EncodeToMemory(privateKeyPEM)

	// Marshal public key to authorized_keys format
	publicKeyBytes := ssh.MarshalAuthorizedKey(sshPubKey)

	// Create public key string with comment
	publicKeyStr := string(publicKeyBytes)
	// ssh.MarshalAuthorizedKey adds a newline, trim it
	publicKeyStr = publicKeyStr[:len(publicKeyStr)-1]
	if comment != "" {
		publicKeyStr = publicKeyStr + " " + comment
	}

	return &KeyPair{
		PrivateKey:   privateKeyBytes,
		PublicKey:    publicKeyBytes,
		PublicKeyStr: publicKeyStr,
	}, nil
}
