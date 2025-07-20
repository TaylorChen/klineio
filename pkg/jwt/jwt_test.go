package jwt

import (
	"testing"
	"time"

	"github.com/spf13/viper"
)

func newTestJWT() *JWT {
	v := viper.New()
	v.Set("security.jwt.key", "testkey1234567890")
	return NewJwt(v)
}

func TestGenTokenAndParseToken(t *testing.T) {
	jwt := newTestJWT()
	token, err := jwt.GenToken("user1", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("GenToken error: %v", err)
	}
	claims, err := jwt.ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken error: %v", err)
	}
	if claims.UserId != "user1" {
		t.Errorf("UserId not match")
	}
}

func TestParseToken_Empty(t *testing.T) {
	jwt := newTestJWT()
	_, err := jwt.ParseToken("")
	if err == nil {
		t.Errorf("expected error for empty token")
	}
}

func TestParseToken_Invalid(t *testing.T) {
	jwt := newTestJWT()
	_, err := jwt.ParseToken("invalidtoken")
	if err == nil {
		t.Errorf("expected error for invalid token")
	}
}
