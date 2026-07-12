package auth

import "testing"

func TestPasswordRoundTrip(t *testing.T) {
	hash, err := HashPassword("A-strong-password")
	if err != nil {
		t.Fatal(err)
	}
	if !CheckPassword("A-strong-password", hash) {
		t.Fatal("expected password match")
	}
	if CheckPassword("wrong-password", hash) {
		t.Fatal("unexpected password match")
	}
}

func TestRefreshTokenIsHashed(t *testing.T) {
	raw, hash, err := NewRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	if raw == hash || HashRefreshToken(raw) != hash {
		t.Fatal("refresh token hash mismatch")
	}
}
