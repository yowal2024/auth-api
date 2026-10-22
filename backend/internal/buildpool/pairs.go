package buildpool

import (
	"auth-api/internal/buildpool/edge"
	"bytes"
	"encoding/binary"
)

func preservePairs(apk []byte, original map[uint32][]byte) ([]byte, error) {
	pairs, e := edge.Pairs(apk)
	if e != nil {
		return nil, e
	}
	for id, v := range original {
		if id != 0x7109871a && id != 0xf05368c0 && id != 0x1b93ad61 && id != 0x6dff800d {
			pairs[id] = v
		}
	}
	var body bytes.Buffer
	for id, v := range pairs {
		binary.Write(&body, binary.LittleEndian, uint64(4+len(v)))
		binary.Write(&body, binary.LittleEndian, id)
		body.Write(v)
	}
	size := uint64(body.Len() + 24)
	var block bytes.Buffer
	binary.Write(&block, binary.LittleEndian, size)
	block.Write(body.Bytes())
	binary.Write(&block, binary.LittleEndian, size)
	block.Write(edge.Magic)
	return edge.AddBlock(apk, block.Bytes())
}
