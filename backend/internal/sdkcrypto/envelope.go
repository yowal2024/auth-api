package sdkcrypto

import (
	"auth-api/internal/repo"
	"errors"
)

func DecryptPayload(e *repo.Encrypt, data string) ([]byte, error) {
	b, err := DecodeBlob(data, e.Encoding)
	if err != nil {
		return nil, err
	}
	switch e.Algorithm {
	case "aes", "session":
		k, ok := decodeKeyN(e.SymKey, 16, 24, 32)
		if !ok {
			return nil, errors.New("invalid AES key")
		}
		return DecryptAESCBC(k, b)
	case "des":
		k, ok := decodeKeyN(e.SymKey, 8)
		if !ok {
			return nil, errors.New("invalid DES key")
		}
		return DecryptDESCBC(k, b)
	case "rc4":
		return RC4Crypt([]byte(e.SymKey), b)
	case "rsa":
		return RSADecrypt(e.ServerPrivateKey, b)
	}
	return nil, errors.New("unsupported encryption")
}
func EncryptPayload(e *repo.Encrypt, data []byte) (string, error) {
	var b []byte
	var err error
	switch e.Algorithm {
	case "aes", "session":
		k, ok := decodeKeyN(e.SymKey, 16, 24, 32)
		if !ok {
			return "", errors.New("invalid AES key")
		}
		b, err = EncryptAESCBC(k, data)
	case "des":
		k, ok := decodeKeyN(e.SymKey, 8)
		if !ok {
			return "", errors.New("invalid DES key")
		}
		b, err = EncryptDESCBC(k, data)
	case "rc4":
		b, err = RC4Crypt([]byte(e.SymKey), data)
	case "rsa":
		b, err = RSAEncrypt(e.ClientPublicKey, data)
	default:
		return "", errors.New("unsupported encryption")
	}
	if err != nil {
		return "", err
	}
	return EncodeBlob(b, e.Encoding), nil
}
