package apkzip

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"strings"
)

func Repack(apk, manifest []byte) ([]byte, error) {
	r, e := zip.NewReader(bytes.NewReader(apk), int64(len(apk)))
	if e != nil {
		return nil, e
	}
	var out bytes.Buffer
	w := zip.NewWriter(&out)
	found := false
	for _, f := range r.File {
		upper := strings.ToUpper(f.Name)
		if strings.HasPrefix(upper, "META-INF/") && (strings.HasSuffix(upper, ".SF") || strings.HasSuffix(upper, ".RSA") || strings.HasSuffix(upper, ".DSA") || strings.HasSuffix(upper, ".EC") || upper == "META-INF/MANIFEST.MF") {
			continue
		}
		h := f.FileHeader
		h.Extra = append([]byte{}, h.Extra...)
		h.Flags &^= 8
		var content io.Reader
		if f.Name == "AndroidManifest.xml" {
			if found {
				return nil, fmt.Errorf("duplicate manifest")
			}
			found = true
			var compressed bytes.Buffer
			fw, e := flate.NewWriter(&compressed, flate.DefaultCompression)
			if e != nil {
				return nil, e
			}
			if _, e = fw.Write(manifest); e != nil {
				return nil, e
			}
			if e = fw.Close(); e != nil {
				return nil, e
			}
			h.Method = zip.Deflate
			h.CRC32 = crc32.ChecksumIEEE(manifest)
			h.UncompressedSize64 = uint64(len(manifest))
			h.CompressedSize64 = uint64(compressed.Len())
			h.UncompressedSize = uint32(h.UncompressedSize64)
			h.CompressedSize = uint32(h.CompressedSize64)
			content = bytes.NewReader(compressed.Bytes())
		} else {
			content, e = f.OpenRaw()
			if e != nil {
				return nil, e
			}
		}
		if e = w.Flush(); e != nil {
			return nil, e
		}
		// Stored data aligns to four bytes; native libraries to 16 KiB.
		if h.Method == zip.Store && !strings.HasSuffix(h.Name, "/") {
			align := 4
			if strings.HasSuffix(h.Name, ".so") {
				align = 16384
			}
			offset := out.Len() + 30 + len(h.Name) + len(h.Extra)
			if offset%align != 0 {
				padding := (align - (offset+4)%align) % align
				extra := make([]byte, 4+padding)
				binary.LittleEndian.PutUint16(extra, 0xd935)
				binary.LittleEndian.PutUint16(extra[2:], uint16(padding))
				h.Extra = append(h.Extra, extra...)
			}
		}
		if len(h.Extra) > 65535 {
			return nil, fmt.Errorf("ZIP extra field too large")
		}
		dst, e := w.CreateRaw(&h)
		if e != nil {
			return nil, e
		}
		if _, e = io.Copy(dst, content); e != nil {
			return nil, e
		}
	}
	if !found {
		return nil, fmt.Errorf("AndroidManifest.xml missing")
	}
	if e = w.Close(); e != nil {
		return nil, e
	}
	return out.Bytes(), nil
}
