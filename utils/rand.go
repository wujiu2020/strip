package utils

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"math/big"
)

var (
	HexAlphabets       = []byte("0123456789abcdef")
	LowercaseAlphabets = []byte("0123456789abcdefghijklmnopqrstuvwxyz")
	DefaultAlphabets   = []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")
)

// randomCreateBytes generate random []byte by specify chars.
func RandomCreateBytes(n int, alphabets ...byte) ([]byte, error) {
	if len(alphabets) == 0 {
		alphabets = DefaultAlphabets
	}
	var bytes = make([]byte, n)
	if num, err := rand.Read(bytes); num != n || err != nil {
		if err == nil {
			err = fmt.Errorf("random string not enough length: need %d but %d", n, num)
		}
		return nil, err
	}
	for i, b := range bytes {
		bytes[i] = alphabets[b%byte(len(alphabets))]
	}
	return bytes, nil
}

func RandomCreateString(n int, alphabets ...byte) (string, error) {
	bytes, err := RandomCreateBytes(n, alphabets...)
	return string(bytes), err
}

func MustRandomCreateBytes(n int, alphabets ...byte) []byte {
	b, err := RandomCreateBytes(n, alphabets...)
	if err != nil {
		panic(err)
	}
	return b
}

func MustRandomCreateString(n int, alphabets ...byte) string {
	return string(MustRandomCreateBytes(n, alphabets...))
}

func DefaultNumberEncode(number string) string {
	return NumberEncode(number, DefaultAlphabets)
}

func NumberEncode(number string, alphabet []byte) string {
	token := make([]byte, 0, 12)
	x, ok := new(big.Int).SetString(number, 10)
	if !ok {
		return ""
	}
	y := big.NewInt(int64(len(alphabet)))
	m := new(big.Int)
	for x.Sign() > 0 {
		x, m = x.DivMod(x, y, m)
		token = append(token, alphabet[int(m.Int64())])
	}
	return string(token)
}

func NumberDecode(token string, alphabet []byte) string {
	x := new(big.Int)
	y := big.NewInt(int64(len(alphabet)))
	z := new(big.Int)
	for i := len(token) - 1; i >= 0; i-- {
		v := bytes.IndexByte(alphabet, token[i])
		z.SetInt64(int64(v))
		x.Mul(x, y)
		x.Add(x, z)
	}
	return x.String()
}
