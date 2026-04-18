package utilis

import (
	"crypto/rand"
	"math/big"
)

const (
	uppercaseCharset = "ABCDEFGHJKLMNPQRSTUVWXYZ"
	lowercaseCharset = "abcdefghijkmnpqrstuvwxyz"
	digitCharset     = "23456789"
	symbolCharset    = "!@#$%&*"
	allCharset       = uppercaseCharset + lowercaseCharset + digitCharset + symbolCharset
)

func GenerateRandomPassword(length int) (string, error) {
	if length < 8 {
		length = 10
	}

	password := make([]byte, length)

	categories := []string{uppercaseCharset, lowercaseCharset, digitCharset, symbolCharset}
	for i, set := range categories {
		ch, err := randomChar(set)
		if err != nil {
			return "", err
		}
		password[i] = ch
	}

	for i := len(categories); i < length; i++ {
		ch, err := randomChar(allCharset)
		if err != nil {
			return "", err
		}
		password[i] = ch
	}

	if err := shuffle(password); err != nil {
		return "", err
	}

	return string(password), nil
}

func randomChar(charset string) (byte, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
	if err != nil {
		return 0, err
	}
	return charset[n.Int64()], nil
}

func shuffle(b []byte) error {
	for i := len(b) - 1; i > 0; i-- {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return err
		}
		b[i], b[j.Int64()] = b[j.Int64()], b[i]
	}
	return nil
}
