package edge

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

var Magic = []byte("APK Sig Block 42")

func FindEOCD(b []byte) (int, error) {
	for i := len(b) - 22; i >= 0 && i >= len(b)-65557; i-- {
		if binary.LittleEndian.Uint32(b[i:]) == 0x06054b50 && i+22+int(binary.LittleEndian.Uint16(b[i+20:])) == len(b) {
			return i, nil
		}
	}
	return 0, fmt.Errorf("EOCD not found")
}
func SigningBlock(b []byte) (start, cd, eocd int, err error) {
	eocd, err = FindEOCD(b)
	if err != nil {
		return
	}
	cd = int(binary.LittleEndian.Uint32(b[eocd+16:]))
	start = cd
	if cd < 0 || cd > eocd {
		err = fmt.Errorf("invalid central directory")
		return
	}
	if cd < 24 || !bytes.Equal(b[cd-16:cd], Magic) {
		return
	}
	size := binary.LittleEndian.Uint64(b[cd-24:])
	if size < 24 || size > uint64(cd-8) {
		err = fmt.Errorf("invalid signing block size")
		return
	}
	start = cd - int(size) - 8
	if binary.LittleEndian.Uint64(b[start:]) != size {
		err = fmt.Errorf("signing block size mismatch")
	}
	return
}
func Pairs(b []byte) (map[uint32][]byte, error) {
	start, cd, _, e := SigningBlock(b)
	if e != nil {
		return nil, e
	}
	out := map[uint32][]byte{}
	if start == cd {
		return out, nil
	}
	for p := start + 8; p < cd-24; {
		if p+12 > cd-24 {
			return nil, fmt.Errorf("invalid pair")
		}
		n := binary.LittleEndian.Uint64(b[p:])
		if n < 4 || n > uint64(cd-24-p-8) {
			return nil, fmt.Errorf("invalid pair size")
		}
		id := binary.LittleEndian.Uint32(b[p+8:])
		out[id] = append([]byte{}, b[p+12:p+8+int(n)]...)
		p += 8 + int(n)
	}
	return out, nil
}
func StripSigningBlock(b []byte) ([]byte, error) {
	start, cd, eocd, e := SigningBlock(b)
	if e != nil {
		return nil, e
	}
	if start == cd {
		return append([]byte{}, b...), nil
	}
	out := append(append([]byte{}, b[:start]...), b[cd:]...)
	binary.LittleEndian.PutUint32(out[eocd-(cd-start)+16:], uint32(start))
	return out, nil
}
func AddBlock(b, block []byte) ([]byte, error) {
	clean, e := StripSigningBlock(b)
	if e != nil {
		return nil, e
	}
	eocd, e := FindEOCD(clean)
	if e != nil {
		return nil, e
	}
	cd := int(binary.LittleEndian.Uint32(clean[eocd+16:]))
	out := append(append(append([]byte{}, clean[:cd]...), block...), clean[cd:]...)
	binary.LittleEndian.PutUint32(out[eocd+len(block)+16:], uint32(cd+len(block)))
	return out, nil
}
