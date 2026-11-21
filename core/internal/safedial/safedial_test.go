package safedial

import (
	"net"
	"net/http"
	"testing"
)

func TestValidateRejectsUnsafeAddresses(t *testing.T) {
	for _, source := range []string{
		"file:///etc/passwd",
		"javascript:alert(1)",
		"http://user:password@example.com/image.png",
		"http://example.com:70000/image.png",
		"http://example.com:0/image.png",
		"http://[fd00::1]/image.png",
	} {
		if Validate(source) == nil {
			t.Fatalf("Validate accepted %q", source)
		}
	}
	for _, source := range []string{
		"http://example.com/image.png",
		"https://example.com/image.png",
		"http://example.com:8080/image.png",
	} {
		if err := Validate(source); err != nil {
			t.Fatalf("Validate rejected %q: %v", source, err)
		}
	}
}

func TestPublicAddressRejectsPrivateNetworks(t *testing.T) {
	for _, source := range []string{
		"127.0.0.1",
		"10.0.0.1",
		"169.254.169.254",
		"100.64.0.1",
		"::1",
		"fe80::1",
	} {
		if PublicAddress(net.ParseIP(source)) {
			t.Fatalf("PublicAddress accepted %q", source)
		}
	}
	if !PublicAddress(net.ParseIP("93.184.216.34")) {
		t.Fatal("PublicAddress rejected a public address")
	}
}

func TestCheckRedirectRejectsPrivateTargetsAndDowngrades(t *testing.T) {
	target := func(raw string) *http.Request {
		t.Helper()
		req, err := http.NewRequest(http.MethodGet, raw, nil)
		if err != nil {
			t.Fatal(err)
		}
		return req
	}
	via := target("https://example.com/")
	if CheckRedirect(target("http://127.0.0.1/x"), []*http.Request{via}) == nil {
		t.Fatal("redirect to a loopback address was allowed")
	}
	if CheckRedirect(target("http://example.com/x"), []*http.Request{via}) == nil {
		t.Fatal("https to http downgrade was allowed")
	}
	if err := CheckRedirect(target("https://example.com/x"), []*http.Request{via}); err != nil {
		t.Fatalf("safe redirect rejected: %v", err)
	}
}
