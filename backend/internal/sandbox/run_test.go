package sandbox

import (
	"context"
	"testing"
	"time"
)

func TestFunctionContextCryptoAndTimeout(t *testing.T) {
	r := Run(context.Background(), RunOptions{Code: `function handler(ctx){console.log(ctx.payload.name);return {hash:crypto.sha256("abc"),value:ctx.payload.n+1};}`, Payload: map[string]any{"name": "fixture", "n": 41}})
	if !r.OK {
		t.Fatal(r.Error)
	}
	v := r.Result.(map[string]any)
	if v["hash"] != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" || len(r.Logs) != 1 {
		t.Fatal(r)
	}
	start := time.Now()
	r = Run(context.Background(), RunOptions{Code: `function handler(){while(true){}}`, Timeout: 20 * time.Millisecond})
	if r.OK || time.Since(start) > time.Second {
		t.Fatal("timeout did not interrupt")
	}
}
func TestBusinessErrorMessage(t *testing.T) {
	r := Run(context.Background(), RunOptions{Code: "function handler(){throw new Error('业务错误:余额不足');}"})
	if r.OK || r.Error != "业务错误:余额不足" {
		t.Fatalf("business error lost: %+v", r)
	}
}
