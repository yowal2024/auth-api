// Reconstructed declarations from Go runtime metadata; method bodies are not recovered.
package apkzip

// Recovered from runtime type 0x138dd80; original size 112 bytes.
type cdEntry struct {
	versionMadeBy uint16
	versionNeeded uint16
	gpFlag        uint16
	method        uint16
	dosTime       uint16
	dosDate       uint16
	crc32         uint32
	csize         uint32
	usize         uint32
	fname         []uint8
	extra         []uint8
	comment       []uint8
	diskNo        uint16
	internalAttr  uint16
	externalAttr  uint32
	lfhOffset     uint32
}
