// Package bsdiff implements the original DELTA1 suffix-array matching format.
package bsdiff

import (
	flate "auth-api/internal/buildpool/bsdiff/internal/flate125"
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"hash/adler32"
	"index/suffixarray"
	"io"
	"math"
)

var magic = []byte{'D', 'E', 'L', 'T', 'A', '1', 0, 0}

func compressed(b []byte) ([]byte, error) {
	var out bytes.Buffer
	// Original binary was built with Go 1.25.11. Go 1.26 changed DEFLATE
	// output, so pin that encoder and retain the RFC 1950 zlib envelope.
	out.Write([]byte{0x78, 0xda})
	z, err := flate.NewWriter(&out, flate.BestCompression)
	if err != nil {
		return nil, err
	}
	if _, e := z.Write(b); e != nil {
		return nil, e
	}
	if e := z.Close(); e != nil {
		return nil, e
	}
	binary.Write(&out, binary.BigEndian, adler32.Checksum(b))
	return out.Bytes(), nil
}
func writeOffset(b *bytes.Buffer, n int64) {
	v := uint64(n)
	if n < 0 {
		v = uint64(-n) | (1 << 63)
	}
	binary.Write(b, binary.LittleEndian, v)
}
func Build(old, new []byte) ([]byte, error) {
	control, diff, literal, e := bsdiffStreams(old, new)
	if e != nil {
		return nil, e
	}
	ctrl, e := compressed(control)
	if e != nil {
		return nil, e
	}
	delta, e := compressed(diff)
	if e != nil {
		return nil, e
	}
	extra, e := compressed(literal)
	if e != nil {
		return nil, e
	}
	var out bytes.Buffer
	out.Write(magic)
	for _, x := range []uint64{uint64(len(new)), uint64(len(ctrl)), uint64(len(delta)), uint64(len(extra))} {
		binary.Write(&out, binary.LittleEndian, x)
	}
	out.Write(ctrl)
	out.Write(delta)
	out.Write(extra)
	return out.Bytes(), nil
}

// Recovered from bsdiffStreams and longestMatch in the original Go executable:
// 32-byte seeds, at most 32 suffix-array candidates, first longest match wins.
// Separate control records flush literal bytes and seek before each copy span.
func bsdiffStreams(old, new []byte) ([]byte, []byte, []byte, error) {
	if len(old) == 0 || len(new) == 0 {
		return nil, nil, nil, fmt.Errorf("bsdiff: empty input")
	}
	index := suffixarray.New(old)
	var control, diff, extra, pending bytes.Buffer
	oldPos := 0
	for newPos := 0; newPos < len(new); {
		length, pos := longestMatch(index, old, new[newPos:])
		if length == 0 {
			pending.WriteByte(new[newPos])
			newPos++
			continue
		}
		seek := pos - oldPos
		if pending.Len() > 0 || seek != 0 || newPos == 0 {
			writeOffset(&control, 0)
			writeOffset(&control, int64(pending.Len()))
			writeOffset(&control, int64(seek))
			extra.Write(pending.Bytes())
			pending.Reset()
			oldPos = pos
		}
		for i := 0; i < length; i++ {
			diff.WriteByte(new[newPos+i] - old[pos+i])
		}
		writeOffset(&control, int64(length))
		writeOffset(&control, 0)
		writeOffset(&control, 0)
		oldPos += length
		newPos += length
	}
	if pending.Len() > 0 {
		writeOffset(&control, 0)
		writeOffset(&control, int64(pending.Len()))
		writeOffset(&control, 0)
		extra.Write(pending.Bytes())
	}
	return control.Bytes(), diff.Bytes(), extra.Bytes(), nil
}

func longestMatch(index *suffixarray.Index, old, next []byte) (length, pos int) {
	const seed = 32
	if len(next) < seed {
		return 0, 0
	}
	candidates := index.Lookup(next[:seed], 32)
	if len(candidates) == 0 {
		return 0, 0
	}
	length, pos = seed, candidates[0]
	for _, candidate := range candidates {
		n := 0
		for n < len(next) && candidate+n < len(old) && old[candidate+n] == next[n] {
			n++
		}
		if n > length {
			length, pos = n, candidate
		}
	}
	return length, pos
}
func Apply(old, patch []byte) ([]byte, error) {
	if len(patch) < 40 || !bytes.Equal(patch[:8], magic) {
		return nil, fmt.Errorf("invalid DELTA1 header")
	}
	size := binary.LittleEndian.Uint64(patch[8:])
	if size > 256<<20 {
		return nil, fmt.Errorf("patch output too large")
	}
	lengths := []uint64{binary.LittleEndian.Uint64(patch[16:]), binary.LittleEndian.Uint64(patch[24:]), binary.LittleEndian.Uint64(patch[32:])}
	streams := [][]byte{}
	offset := uint64(40)
	for i, n := range lengths {
		if n > uint64(len(patch))-offset {
			return nil, fmt.Errorf("truncated stream")
		}
		z, e := zlib.NewReader(bytes.NewReader(patch[offset : offset+n]))
		if e != nil {
			return nil, e
		}
		limit := int64(size) + 1
		if i == 0 {
			limit = 24 * int64(size+1)
		}
		if limit > 256<<20 {
			limit = 256 << 20
		}
		b, e := io.ReadAll(io.LimitReader(z, limit+1))
		z.Close()
		if e != nil || int64(len(b)) > limit {
			return nil, fmt.Errorf("invalid compressed stream")
		}
		streams = append(streams, b)
		offset += n
	}
	if offset != uint64(len(patch)) {
		return nil, fmt.Errorf("trailing patch bytes")
	}
	out := make([]byte, int(size))
	np, op, dp, ep, cp := int64(0), int64(0), int64(0), int64(0), 0
	read := func(b []byte) int64 {
		v := binary.LittleEndian.Uint64(b)
		n := int64(v &^ uint64(1<<63))
		if v>>63 != 0 {
			return -n
		}
		return n
	}
	for np < int64(size) {
		if cp+24 > len(streams[0]) {
			return nil, fmt.Errorf("truncated control")
		}
		x, y, z := read(streams[0][cp:]), read(streams[0][cp+8:]), read(streams[0][cp+16:])
		cp += 24
		// Zero-output seek records occur in original patches. The finite control
		// stream bounds their count; reject position overflow before arithmetic.
		if x < 0 || y < 0 || x > int64(size)-np || y > int64(size)-np-x || x > int64(len(streams[1]))-dp || y > int64(len(streams[2]))-ep {
			return nil, fmt.Errorf("invalid control")
		}
		if op > math.MaxInt64-x {
			return nil, fmt.Errorf("old position overflow")
		}
		for i := int64(0); i < x; i++ {
			var b byte
			if op+i >= 0 && op+i < int64(len(old)) {
				b = old[op+i]
			}
			out[np+i] = b + streams[1][dp+i]
		}
		np += x
		op += x
		dp += x
		copy(out[np:np+y], streams[2][ep:ep+y])
		np += y
		ep += y
		if z > 1<<40 || z < -(1<<40) {
			return nil, fmt.Errorf("invalid seek")
		}
		if (z > 0 && op > math.MaxInt64-z) || (z < 0 && op < math.MinInt64-z) {
			return nil, fmt.Errorf("old position overflow")
		}
		op += z
	}
	return out, nil
}
