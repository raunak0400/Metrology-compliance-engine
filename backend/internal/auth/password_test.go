package auth

import "testing"

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("Correct-Horse-1")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if hash == "Correct-Horse-1" {
		t.Fatal("expected the hash to differ from the plaintext")
	}

	ok, err := VerifyPassword(hash, "Correct-Horse-1")
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if !ok {
		t.Fatal("expected the correct password to verify")
	}

	ok, err = VerifyPassword(hash, "wrong-password")
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if ok {
		t.Fatal("expected the wrong password to fail verification")
	}
}
