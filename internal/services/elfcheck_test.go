package services

import (
	"os"
	"path/filepath"
	"testing"
)

// fakeELF writes a synthetic ELF64 header with configurable machine + class bytes.
func fakeELF(t *testing.T, name string, class, data byte, machine uint16) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	hdr := make([]byte, 64)
	copy(hdr[0:4], []byte{0x7f, 'E', 'L', 'F'})
	hdr[4] = class
	hdr[5] = data
	hdr[18] = byte(machine & 0xff)
	hdr[19] = byte(machine >> 8)
	if err := os.WriteFile(path, hdr, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestVerifyLinuxARM64_Valid(t *testing.T) {
	path := fakeELF(t, "good", 2, 1, 0xB7)
	if err := VerifyLinuxARM64(path); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVerifyLinuxARM64_RejectAmd64(t *testing.T) {
	path := fakeELF(t, "bad-amd64", 2, 1, 0x3E) // EM_X86_64
	err := VerifyLinuxARM64(path)
	if err == nil {
		t.Fatal("expected error for amd64 ELF")
	}
}

func TestVerifyLinuxARM64_RejectNonELF(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "not-elf")
	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := VerifyLinuxARM64(path); err == nil {
		t.Fatal("expected error for non-ELF file")
	}
}

func TestVerifyLinuxARM64_RejectELF32(t *testing.T) {
	path := fakeELF(t, "elf32", 1, 1, 0xB7)
	if err := VerifyLinuxARM64(path); err == nil {
		t.Fatal("expected error for ELF32")
	}
}

func TestVerifyLinuxARM64_MissingFile(t *testing.T) {
	if err := VerifyLinuxARM64("/does/not/exist"); err == nil {
		t.Fatal("expected error for missing file")
	}
}
