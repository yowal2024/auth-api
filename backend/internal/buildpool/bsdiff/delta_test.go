package bsdiff

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"os"
	"testing"
)

func TestDeltaRoundTripAndMalformed(t *testing.T) {
	for _, pair := range [][2][]byte{{[]byte("old"), []byte("new")}, {[]byte("prefix-old-tail"), []byte("prefix-new-and-longer-tail")}, {bytes.Repeat([]byte{255}, 8192), bytes.Repeat([]byte{1}, 4096)}} {
		patch, err := Build(pair[0], pair[1])
		if err != nil {
			t.Fatal(err)
		}
		got, err := Apply(pair[0], patch)
		if err != nil || !bytes.Equal(got, pair[1]) {
			t.Fatal("roundtrip", err)
		}
		for n := 0; n < len(patch); n++ {
			if _, err := Apply(pair[0], patch[:n]); err == nil {
				t.Fatal("accepted truncation", n)
			}
		}
		if _, err := Apply(pair[0], append(patch, 0)); err == nil {
			t.Fatal("accepted trailing bytes")
		}
	}
	patch, _ := Build([]byte("old"), []byte("new"))
	for _, offset := range []int{8, 16, 24, 32} {
		bad := append([]byte{}, patch...)
		binary.LittleEndian.PutUint64(bad[offset:], ^uint64(0))
		if _, err := Apply(nil, bad); err == nil {
			t.Fatal("accepted overflowing header", offset)
		}
	}
}

func FuzzApply(f *testing.F) {
	patch, _ := Build([]byte("old"), []byte("new"))
	f.Add(patch)
	var fixtures struct {
		Blobs map[string][]byte
		Cases []struct{ Patch string }
	}
	raw, err := os.ReadFile("testdata/original.json")
	if err != nil {
		f.Fatal(err)
	}
	if err := json.Unmarshal(raw, &fixtures); err != nil {
		f.Fatal(err)
	}
	for _, c := range fixtures.Cases {
		f.Add(fixtures.Blobs[c.Patch])
	}
	f.Fuzz(func(t *testing.T, patch []byte) {
		// Keep fuzz allocations bounded while exercising all parser paths.
		if len(patch) >= 16 && binary.LittleEndian.Uint64(patch[8:]) > 1<<16 {
			return
		}
		_, _ = Apply([]byte("old"), patch)
	})
}
