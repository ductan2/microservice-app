package utils

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"errors"
	"math"
	"time"
)

var (
	ErrMFARequired    = errors.New("mfa required")
	ErrInvalidMFACode = errors.New("invalid mfa code")
	ErrEmailNotVerified = errors.New("email not verified")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

// VerifyTOTP verifies a 6-digit TOTP code against the given base32 secret.
// RFC 6238, 30s window, allows +/- one step for clock skew.
func VerifyTOTP(secretBase32 string, code string, now time.Time) bool {
	secret, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secretBase32)
	if err != nil {
		return false
	}

	timestep := int64(30)
	counter := now.Unix() / timestep
	// allow -1, 0, +1 windows
	for i := int64(-1); i <= 1; i++ {
		if hotp(secret, uint64(counter+i)) == code {
			return true
		}
	}
	return false
}

func hotp(key []byte, counter uint64) string {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], counter)
	mac := hmac.New(sha1.New, key)
	mac.Write(buf[:])
	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 0x0F
	binaryCode := (int(sum[offset])&0x7f)<<24 |
		(int(sum[offset+1])&0xff)<<16 |
		(int(sum[offset+2])&0xff)<<8 |
		(int(sum[offset+3]) & 0xff)
	otp := binaryCode % int(math.Pow10(6))
	return leftPad6(otp)
}

func leftPad6(n int) string {
	s := "000000"
	x := []byte(s)
	i := 5
	for n > 0 && i >= 0 {
		x[i] = byte('0' + n%10)
		n /= 10
		i--
	}
	return string(x)
}
