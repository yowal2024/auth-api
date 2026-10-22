package axml

import (
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

func u16(b []byte, o int) uint16 { return binary.LittleEndian.Uint16(b[o:]) }
func u32(b []byte, o int) uint32 { return binary.LittleEndian.Uint32(b[o:]) }

type pool struct {
	start, end, header int
	flags              uint32
	values             []string
	styles             []byte
	styleOffsets       []byte
}

func parse(b []byte) (pool, error) {
	var p pool
	if len(b) < 36 || u16(b, 0) != 3 || int(u32(b, 4)) != len(b) {
		return p, fmt.Errorf("invalid Android binary XML")
	}
	p.start = int(u16(b, 2))
	if p.start+28 > len(b) || u16(b, p.start) != 1 {
		return p, fmt.Errorf("string pool missing")
	}
	p.header = int(u16(b, p.start+2))
	size := int(u32(b, p.start+4))
	p.end = p.start + size
	if p.header < 28 || p.end > len(b) || size < p.header {
		return p, fmt.Errorf("invalid string pool")
	}
	count := int(u32(b, p.start+8))
	styles := int(u32(b, p.start+12))
	p.flags = u32(b, p.start+16)
	data := p.start + int(u32(b, p.start+20))
	styleStart := int(u32(b, p.start+24))
	if count < 0 || styles < 0 || count > (size-p.header)/4 || styles > (size-p.header-count*4)/4 {
		return p, fmt.Errorf("invalid string count")
	}
	limit := p.end
	if styleStart > 0 {
		limit = p.start + styleStart
		if limit < data || limit > p.end {
			return p, fmt.Errorf("invalid styles")
		}
		p.styles = append([]byte{}, b[limit:p.end]...)
	}
	p.styleOffsets = append([]byte{}, b[p.start+p.header+count*4:p.start+p.header+(count+styles)*4]...)
	for i := 0; i < count; i++ {
		o := data + int(u32(b, p.start+p.header+i*4))
		if o < data || o >= limit {
			return p, fmt.Errorf("invalid string offset")
		}
		readLen8 := func() (int, error) {
			if o >= limit {
				return 0, io.ErrUnexpectedEOF
			}
			v := int(b[o])
			o++
			if v&128 != 0 {
				if o >= limit {
					return 0, io.ErrUnexpectedEOF
				}
				v = (v&127)<<8 | int(b[o])
				o++
			}
			return v, nil
		}
		if p.flags&0x100 != 0 {
			if _, e := readLen8(); e != nil {
				return p, e
			}
			n, e := readLen8()
			if e != nil || n > limit-o-1 {
				return p, fmt.Errorf("invalid UTF8 string")
			}
			p.values = append(p.values, string(b[o:o+n]))
		} else {
			if o+2 > limit {
				return p, io.ErrUnexpectedEOF
			}
			n := int(u16(b, o))
			o += 2
			if n&0x8000 != 0 {
				if o+2 > limit {
					return p, io.ErrUnexpectedEOF
				}
				n = (n&0x7fff)<<16 | int(u16(b, o))
				o += 2
			}
			if n > (limit-o-2)/2 {
				return p, fmt.Errorf("invalid UTF16 string")
			}
			v := make([]uint16, n)
			for j := range v {
				v[j] = u16(b, o+j*2)
			}
			p.values = append(p.values, string(utf16.Decode(v)))
		}
	}
	return p, nil
}
func packageIndex(b []byte, p pool) (int, error) {
	for o := p.end; o+8 <= len(b); {
		size := int(u32(b, o+4))
		if size < 8 || size > len(b)-o {
			return 0, fmt.Errorf("invalid XML chunk")
		}
		if u16(b, o) == 0x102 {
			h := int(u16(b, o+2))
			if h < 16 || o+h+20 > o+size {
				return 0, fmt.Errorf("invalid element")
			}
			name := int(u32(b, o+h+4))
			if name < len(p.values) && p.values[name] == "manifest" {
				start := int(u16(b, o+h+8))
				stride := int(u16(b, o+h+10))
				count := int(u16(b, o+h+12))
				if stride < 20 {
					return 0, fmt.Errorf("invalid attribute")
				}
				for i := 0; i < count; i++ {
					a := o + h + start + i*stride
					if a+20 > o+size {
						return 0, io.ErrUnexpectedEOF
					}
					n := int(u32(b, a+4))
					if n < len(p.values) && p.values[n] == "package" {
						idx := u32(b, a+8)
						if idx == 0xffffffff && b[a+15] == 3 {
							idx = u32(b, a+16)
						}
						if int(idx) < len(p.values) {
							return int(idx), nil
						}
					}
				}
			}
		}
		o += size
	}
	return 0, fmt.Errorf("manifest package missing")
}
func ReadOriginalPackage(b []byte) (string, error) {
	if bytes.HasPrefix(bytes.TrimSpace(b), []byte("<")) {
		d := xml.NewDecoder(bytes.NewReader(b))
		for {
			t, e := d.Token()
			if e != nil {
				return "", e
			}
			if start, ok := t.(xml.StartElement); ok && start.Name.Local == "manifest" {
				for _, a := range start.Attr {
					if a.Name.Local == "package" {
						return a.Value, nil
					}
				}
			}
		}
	}
	p, e := parse(b)
	if e != nil {
		return "", e
	}
	i, e := packageIndex(b, p)
	if e != nil {
		return "", e
	}
	return p.values[i], nil
}
func PatchManifest(b []byte, old, new string, classes ...map[string]bool) ([]byte, error) {
	p, e := parse(b)
	if e != nil {
		return nil, e
	}
	idx, e := packageIndex(b, p)
	if e != nil {
		return nil, e
	}
	if p.values[idx] != old {
		return nil, fmt.Errorf("original package mismatch")
	}
	// Rewrite individual attributes, preserving shared string-pool values.
	tail := append([]byte{}, b[p.end:]...)
	value := func(n uint32) string {
		if int(n) < len(p.values) {
			return p.values[n]
		}
		return ""
	}
	for o := 0; o+8 <= len(tail); {
		size := int(u32(tail, o+4))
		if size < 8 || size > len(tail)-o {
			return nil, fmt.Errorf("invalid XML chunk")
		}
		if u16(tail, o) == 0x102 {
			h := int(u16(tail, o+2))
			if h < 16 || h+20 > size {
				return nil, fmt.Errorf("invalid XML element")
			}
			tag := value(u32(tail, o+h+4))
			start := int(u16(tail, o+h+8))
			stride := int(u16(tail, o+h+10))
			count := int(u16(tail, o+h+12))
			if start < 20 || stride < 20 || h+start+count*stride > size {
				return nil, fmt.Errorf("invalid XML attributes")
			}
			for j := 0; j < count; j++ {
				a := o + h + start + j*stride
				name := value(u32(tail, a+4))
				ns := value(u32(tail, a))
				rawIdx := u32(tail, a+8)
				typed := tail[a+15] == 3
				vi := rawIdx
				if vi == 0xffffffff && typed {
					vi = u32(tail, a+16)
				}
				if int(vi) >= len(p.values) {
					continue
				}
				v := value(vi)
				replacement := v
				if tag == "manifest" && name == "package" && ns == "" {
					replacement = new
				} else if ns == "http://schemas.android.com/apk/res/android" {
					component := tag == "application" || tag == "activity" || tag == "activity-alias" || tag == "service" || tag == "receiver" || tag == "provider" || tag == "instrumentation"
					classAttr := (component && name == "name") || name == "targetActivity" || name == "parentActivityName" || name == "backupAgent" || name == "appComponentFactory" || name == "manageSpaceActivity" || name == "zygotePreloadName"
					if classAttr {
						if strings.HasPrefix(v, ".") {
							replacement = old + v
						} else if !strings.Contains(v, ".") && v != "" {
							replacement = old + "." + v
						}
					} else {
						identifier := name == "authorities" || name == "permission" || name == "readPermission" || name == "writePermission" || name == "taskAffinity" || name == "process" || name == "targetPackage" || name == "sharedUserId" || (name == "name" && (tag == "permission" || tag == "permission-tree" || tag == "permission-group" || tag == "uses-permission" || tag == "action" || tag == "category"))
						if identifier {
							parts := strings.Split(v, ";")
							for k, s := range parts {
								if s == old || strings.HasPrefix(s, old+".") {
									parts[k] = new + strings.TrimPrefix(s, old)
								}
							}
							replacement = strings.Join(parts, ";")
						}
					}
				}
				if replacement != v {
					ni := uint32(len(p.values))
					p.values = append(p.values, replacement)
					if rawIdx != 0xffffffff {
						binary.LittleEndian.PutUint32(tail[a+8:], ni)
					}
					if typed {
						binary.LittleEndian.PutUint32(tail[a+16:], ni)
					}
				}
			}
		}
		o += size
	}
	var data bytes.Buffer
	offsets := make([]uint32, len(p.values))
	writeLen8 := func(n int) error {
		if n > 32767 {
			return fmt.Errorf("string too long")
		}
		if n > 127 {
			data.WriteByte(byte(n>>8) | 128)
		}
		data.WriteByte(byte(n))
		return nil
	}
	for i, s := range p.values {
		offsets[i] = uint32(data.Len())
		if p.flags&0x100 != 0 {
			if e = writeLen8(len(utf16.Encode([]rune(s)))); e != nil {
				return nil, e
			}
			if e = writeLen8(len(s)); e != nil {
				return nil, e
			}
			data.WriteString(s)
			data.WriteByte(0)
		} else {
			v := utf16.Encode([]rune(s))
			if len(v) > 32767 {
				binary.Write(&data, binary.LittleEndian, uint16(len(v)>>16)|0x8000)
			}
			binary.Write(&data, binary.LittleEndian, uint16(len(v)))
			for _, u := range v {
				binary.Write(&data, binary.LittleEndian, u)
			}
			binary.Write(&data, binary.LittleEndian, uint16(0))
		}
	}
	for data.Len()%4 != 0 {
		data.WriteByte(0)
	}
	header := append([]byte{}, b[p.start:p.start+p.header]...)
	dataOffset := p.header + len(offsets)*4 + len(p.styleOffsets)
	size := dataOffset + data.Len() + len(p.styles)
	binary.LittleEndian.PutUint32(header[4:], uint32(size))
	binary.LittleEndian.PutUint32(header[8:], uint32(len(p.values)))
	binary.LittleEndian.PutUint32(header[16:], p.flags&^1)
	binary.LittleEndian.PutUint32(header[20:], uint32(dataOffset))
	if len(p.styles) > 0 {
		binary.LittleEndian.PutUint32(header[24:], uint32(dataOffset+data.Len()))
	}
	var chunk bytes.Buffer
	chunk.Write(header)
	for _, o := range offsets {
		binary.Write(&chunk, binary.LittleEndian, o)
	}
	chunk.Write(p.styleOffsets)
	chunk.Write(data.Bytes())
	chunk.Write(p.styles)
	out := append(append(append([]byte{}, b[:p.start]...), chunk.Bytes()...), tail...)
	binary.LittleEndian.PutUint32(out[4:], uint32(len(out)))
	return out, nil
}
func DexClasses(b []byte) map[string]bool {
	out := map[string]bool{}
	if len(b) < 112 || !bytes.HasPrefix(b, []byte("dex\n")) {
		return out
	}
	count, offset := int(u32(b, 56)), int(u32(b, 60))
	if count < 0 || offset < 0 || offset > len(b) || count > (len(b)-offset)/4 {
		return out
	}
	for i := 0; i < count; i++ {
		o := int(u32(b, offset+4*i))
		if o < 0 || o >= len(b) {
			continue
		}
		for n := 0; n < 5 && o < len(b); n++ {
			v := b[o]
			o++
			if v < 128 {
				break
			}
		}
		end := o
		for end < len(b) && b[end] != 0 && end-o < 4096 {
			end++
		}
		if end >= len(b) {
			continue
		}
		s := string(b[o:end])
		if strings.HasPrefix(s, "L") && strings.HasSuffix(s, ";") && utf8.ValidString(s) {
			out[strings.ReplaceAll(s[1:len(s)-1], "/", ".")] = true
		}
	}
	return out
}
