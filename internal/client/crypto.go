package client

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"

	"golang.org/x/crypto/chacha20poly1305"
)

const (
	sessionKeySize   = 32
	xchachaNonceSize = chacha20poly1305.NonceSizeX
	graphQLNonceSize = 8
)

func DeriveSessionKey(token string) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, err
	}

	if len(decoded) < sessionKeySize {
		return nil, fmt.Errorf("session token decoded to %d bytes, need at least %d", len(decoded), sessionKeySize)
	}

	key := make([]byte, sessionKeySize)
	copy(key, decoded[:sessionKeySize])

	return key, nil
}

func DeriveLoginKey(password string) []byte {
	sum := sha512.Sum512([]byte(password))
	hexDigest := make([]byte, sha512.Size*2)
	hex.Encode(hexDigest, sum[:])

	return hexDigest[:sessionKeySize]
}

func Encrypt(key, plaintext []byte) ([]byte, error) {
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, xchachaNonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	ciphertext := aead.Seal(nil, nonce, plaintext, nil)
	out := make([]byte, 0, len(nonce)+len(ciphertext))
	out = append(out, nonce...)
	out = append(out, ciphertext...)

	return out, nil
}

func Decrypt(key, ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < xchachaNonceSize {
		return nil, errors.New("ciphertext too short")
	}

	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, err
	}

	nonce := ciphertext[:xchachaNonceSize]
	encrypted := ciphertext[xchachaNonceSize:]

	return aead.Open(nil, nonce, encrypted, nil)
}

func WrapGraphQL(payload []byte, serverTimeMS int64) []byte {
	var nonce [graphQLNonceSize]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		panic(fmt.Errorf("generate graphql nonce: %w", err))
	}

	ts := strconv.FormatInt(serverTimeMS, 10)
	nonceHex := make([]byte, hex.EncodedLen(len(nonce)))
	hex.Encode(nonceHex, nonce[:])

	wrapped := make([]byte, 0, len(ts)+1+len(nonceHex)+1+len(payload))
	wrapped = append(wrapped, ts...)
	wrapped = append(wrapped, '|')
	wrapped = append(wrapped, nonceHex...)
	wrapped = append(wrapped, '|')
	wrapped = append(wrapped, payload...)

	return wrapped
}
