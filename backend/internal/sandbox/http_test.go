package sandbox

import (
	"context"
	"net/netip"
	"testing"
)

func TestScriptHTTPRejectsPrivateDestinations(t *testing.T) {
	for _, address := range []string{"127.0.0.1", "10.0.0.1", "172.16.0.1", "192.168.0.1", "169.254.169.254", "100.100.100.200", "::1", "::ffff:127.0.0.1", "fd00::1", "64:ff9b::a00:1"} {
		if publicAddress(netip.MustParseAddr(address)) {
			t.Fatal("accepted", address)
		}
	}
	if !publicAddress(netip.MustParseAddr("8.8.8.8")) {
		t.Fatal("public address rejected")
	}
	r := Run(context.Background(), RunOptions{Code: `function handler(){return http.get("http://127.0.0.1:1/private");}`})
	if !r.OK || r.Result.(map[string]any)["ok"] != false {
		t.Fatal(r)
	}
}
