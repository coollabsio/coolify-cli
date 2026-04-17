package services

import (
	"encoding/binary"
	"fmt"
	"os"
)

// VerifyLinuxARM64 returns an error if path is not a Linux aarch64 ELF64
// binary. It reads only the first 20 bytes (ELF ident + e_type + e_machine).
func VerifyLinuxARM64(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	var hdr [20]byte
	if _, err := f.ReadAt(hdr[:], 0); err != nil {
		return fmt.Errorf("read ELF header from %s: %w", path, err)
	}

	// Magic 0x7F 'E' 'L' 'F'.
	if hdr[0] != 0x7f || hdr[1] != 'E' || hdr[2] != 'L' || hdr[3] != 'F' {
		return fmt.Errorf("%s is not an ELF binary", path)
	}
	// EI_CLASS=2 (ELF64).
	if hdr[4] != 2 {
		return fmt.Errorf("%s is not ELF64", path)
	}
	// EI_DATA=1 (little-endian).
	if hdr[5] != 1 {
		return fmt.Errorf("%s is not little-endian", path)
	}
	// e_machine at offset 18, u16 LE. 0xB7 = EM_AARCH64.
	machine := binary.LittleEndian.Uint16(hdr[18:20])
	if machine != 0xB7 {
		return fmt.Errorf("%s is not aarch64 (e_machine=0x%x)", path, machine)
	}
	return nil
}
