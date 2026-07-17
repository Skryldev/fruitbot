package crypto

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
)

type Version uint8

const (
	Version1 Version = 1
	Version2 Version = 2
)

var (
	defaultKey  = []byte("ali1343faraz1055antler288based")
	socketKey   = []byte("floatint201412bool23string")
	version2Key = []byte("mwBSDp1nMhcdCravltVGADXTFx7bN9mr0XMgyDezIJghf65lvXhRdLWrScCk")
)

type Encryption struct {
	key               []byte
	requiresDoubleEnc bool
	version           Version
	encryptPool       sync.Pool
	decryptPool       sync.Pool
	encryptCount      uint64
	decryptCount      uint64
	decryptErrors     uint64
}

type Option func(*Encryption)

func WithKey(key []byte) Option {
	return func(e *Encryption) {
		e.key = make([]byte, len(key))
		copy(e.key, key)
	}
}

func WithSocketMode() Option {
	return func(e *Encryption) {
		e.requiresDoubleEnc = true
		e.key = make([]byte, len(socketKey))
		copy(e.key, socketKey)
	}
}

func WithVersion(v Version) Option {
	return func(e *Encryption) {
		e.version = v
	}
}

func NewEncryption(opts ...Option) *Encryption {
	e := &Encryption{
		version: Version1,
	}

	for _, opt := range opts {
		opt(e)
	}

	if len(e.key) == 0 {
		if e.requiresDoubleEnc {
			e.key = make([]byte, len(socketKey))
			copy(e.key, socketKey)
		} else if e.version == Version2 {
			e.key = make([]byte, len(version2Key))
			copy(e.key, version2Key)
		} else {
			e.key = make([]byte, len(defaultKey))
			copy(e.key, defaultKey)
		}
	}

	e.encryptPool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
	e.decryptPool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}

	return e
}

func (e *Encryption) Encrypt(message string) (string, error) {
	if message == "" {
		return "", fmt.Errorf("cannot encrypt empty message")
	}

	atomic.AddUint64(&e.encryptCount, 1)

	messageBytes := []byte(message)

	if e.requiresDoubleEnc {
		encoded := base64.StdEncoding.EncodeToString(messageBytes)
		messageBytes = []byte(encoded)
	}

	encrypted := e.xorEncrypt(messageBytes)
	encoded := base64.StdEncoding.EncodeToString(encrypted)

	if e.requiresDoubleEnc {
		return encoded, nil
	}

	return url.PathEscape(encoded), nil
}

func (e *Encryption) Decrypt(encrypted string) (string, error) {
	if encrypted == "" {
		return "", fmt.Errorf("cannot decrypt empty string")
	}

	trimmed := strings.TrimSpace(encrypted)

	if len(trimmed) > 0 && (trimmed[0] == '{' || trimmed[0] == '[') {
		return trimmed, nil
	}

	if strings.HasPrefix(trimmed, "<!DOCTYPE") || strings.HasPrefix(trimmed, "<html") {
		return trimmed, nil
	}

	atomic.AddUint64(&e.decryptCount, 1)

	if !e.requiresDoubleEnc {
		unquoted, err := url.PathUnescape(encrypted)
		if err != nil {
			unquoted, err = url.QueryUnescape(encrypted)
			if err != nil {
				unquoted = encrypted
			}
		}

		trimmedUnquoted := strings.TrimSpace(unquoted)

		if len(trimmedUnquoted) > 0 && (trimmedUnquoted[0] == '{' || trimmedUnquoted[0] == '[') {
			return trimmedUnquoted, nil
		}

		decoded, err := base64.StdEncoding.DecodeString(unquoted)
		if err != nil {
			atomic.AddUint64(&e.decryptErrors, 1)
			xorResult := e.xorDecrypt([]byte(unquoted))
			return string(xorResult), nil
		}

		decrypted := e.xorDecrypt(decoded)
		return string(decrypted), nil
	}

	decoded, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		atomic.AddUint64(&e.decryptErrors, 1)
		xorResult := e.xorDecrypt([]byte(encrypted))
		return string(xorResult), nil
	}

	decrypted := e.xorDecrypt(decoded)

	if e.requiresDoubleEnc {
		decodedStr, err := base64.StdEncoding.DecodeString(string(decrypted))
		if err != nil {
			return string(decrypted), nil
		}
		return string(decodedStr), nil
	}

	return string(decrypted), nil
}

func (e *Encryption) xorEncrypt(data []byte) []byte {
	keyLen := len(e.key)
	if keyLen == 0 {
		return data
	}

	result := make([]byte, len(data))
	for i := range data {
		result[i] = data[i] ^ e.key[i%keyLen]
	}

	return result
}

func (e *Encryption) xorDecrypt(data []byte) []byte {
	return e.xorEncrypt(data)
}

func (e *Encryption) MustEncrypt(message string) string {
	encrypted, err := e.Encrypt(message)
	if err != nil {
		panic(fmt.Sprintf("encryption failed: %v", err))
	}
	return encrypted
}

func (e *Encryption) MustDecrypt(encrypted string) string {
	decrypted, err := e.Decrypt(encrypted)
	if err != nil {
		panic(fmt.Sprintf("decryption failed: %v", err))
	}
	return decrypted
}

func (e *Encryption) Stats() (encryptCount, decryptCount, decryptErrors uint64) {
	return atomic.LoadUint64(&e.encryptCount),
		atomic.LoadUint64(&e.decryptCount),
		atomic.LoadUint64(&e.decryptErrors)
}

func (e *Encryption) ResetStats() {
	atomic.StoreUint64(&e.encryptCount, 0)
	atomic.StoreUint64(&e.decryptCount, 0)
	atomic.StoreUint64(&e.decryptErrors, 0)
}

func FastXOR(data, key []byte) {
	keyLen := len(key)
	if keyLen == 0 {
		return
	}

	switch keyLen {
	case 1:
		k := key[0]
		for i := range data {
			data[i] ^= k
		}
	default:
		for i := range data {
			data[i] ^= key[i%keyLen]
		}
	}
}

var (
	DefaultEncryption  = NewEncryption()
	SocketEncryption   = NewEncryption(WithSocketMode())
	Version2Encryption = NewEncryption(WithVersion(Version2))
)

func Encrypt(message string) (string, error) {
	return DefaultEncryption.Encrypt(message)
}

func Decrypt(encrypted string) (string, error) {
	return DefaultEncryption.Decrypt(encrypted)
}