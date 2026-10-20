package security_test

import (
	"testing"

	"automarket/internal/security"
)

func TestPBKDF2Hasher(t *testing.T) {
	hasher, err := security.NewPBKDF2Hasher(1_000)
	if err != nil {
		t.Fatalf("new hasher: %v", err)
	}

	encoded, err := hasher.Hash("correct horse battery staple")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if !hasher.Verify("correct horse battery staple", encoded) {
		t.Fatal("expected password to verify")
	}
	if hasher.Verify("wrong password", encoded) {
		t.Fatal("wrong password must not verify")
	}
	if hasher.Verify("correct horse battery staple", "malformed") {
		t.Fatal("malformed hash must not verify")
	}
}

func TestDummyPasswordHashIsValid(t *testing.T) {
	hasher, err := security.NewPBKDF2Hasher(security.DefaultIterations)
	if err != nil {
		t.Fatalf("new hasher: %v", err)
	}
	if !hasher.Verify("not-a-real-password", security.DummyPasswordHash) {
		t.Fatal("dummy hash does not match its documented password")
	}
}

func TestTokenIssuer(t *testing.T) {
	issuer := security.NewTokenIssuer()

	first, err := issuer.Issue()
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	second, err := issuer.Issue()
	if err != nil {
		t.Fatalf("issue second token: %v", err)
	}
	if first == second {
		t.Fatal("tokens must be unique")
	}
	if issuer.Fingerprint(first) == first {
		t.Fatal("fingerprint must differ from the raw token")
	}
	if issuer.Fingerprint(first) != issuer.Fingerprint(first) {
		t.Fatal("fingerprint must be deterministic")
	}
}
