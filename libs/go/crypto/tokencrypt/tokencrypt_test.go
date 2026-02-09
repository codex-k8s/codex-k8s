package tokencrypt

import "testing"

func TestService_RoundTrip(t *testing.T) {
	svc, err := NewService("00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff")
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	enc, err := svc.EncryptString("secret-token")
	if err != nil {
		t.Fatalf("EncryptString: %v", err)
	}

	dec, err := svc.DecryptString(enc)
	if err != nil {
		t.Fatalf("DecryptString: %v", err)
	}
	if dec != "secret-token" {
		t.Fatalf("expected %q, got %q", "secret-token", dec)
	}
}
