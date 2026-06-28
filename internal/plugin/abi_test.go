package plugin

import (
	"testing"
)

func TestPackUnpackPtrLen(t *testing.T) {
	tests := []struct {
		ptr    uint32
		length uint32
	}{
		{0, 0},
		{1024, 256},
		{0xFFFFFFFF, 0},
		{0, 0xFFFFFFFF},
		{0x7FFFFFFF, 0x7FFFFFFF},
		{65536, 4096},
	}

	for _, tt := range tests {
		packed := packPtrLen(tt.ptr, tt.length)
		ptr, length := unpackPtrLen(packed)

		if ptr != tt.ptr {
			t.Errorf("ptr: packPtrLen(%d,%d)=%d, unpack=%d, want=%d",
				tt.ptr, tt.length, packed, ptr, tt.ptr)
		}
		if length != tt.length {
			t.Errorf("len: packPtrLen(%d,%d)=%d, unpack=%d, want=%d",
				tt.ptr, tt.length, packed, length, tt.length)
		}
	}
}

func TestPackPtrLenRoundTrip(t *testing.T) {
	// QuickCheck-style
	for ptr := uint32(0); ptr < 10000; ptr += 73 {
		for length := uint32(0); length < 10000; length += 97 {
			packed := packPtrLen(ptr, length)
			p, l := unpackPtrLen(packed)
			if p != ptr || l != length {
				t.Fatalf("round-trip failed: (%d,%d) → %d → (%d,%d)", ptr, length, packed, p, l)
			}
		}
	}
}

func TestUnpackZeroes(t *testing.T) {
	p, l := unpackPtrLen(0)
	if p != 0 || l != 0 {
		t.Errorf("unpack(0) = (%d,%d), want (0,0)", p, l)
	}
}

func TestPackPtrLenMaxValues(t *testing.T) {
	// Max uint32 values should round-trip
	packed := packPtrLen(0xFFFFFFFF, 0xFFFFFFFF)
	p, l := unpackPtrLen(packed)
	// Note: packing loses some precision when both are max because they overlap in 64 bits.
	// ptr uses high 32, len uses low 32. Both 0xFFFFFFFF:
	// packed = (0xFFFFFFFF << 32) | 0xFFFFFFFF = 0xFFFFFFFF_FFFFFFFF
	// p = 0xFFFFFFFF >> 32 = 0xFFFFFFFF ✓
	// l = 0xFFFFFFFF_FFFFFFFF & 0xFFFFFFFF = 0xFFFFFFFF ✓
	if p != 0xFFFFFFFF {
		t.Errorf("max ptr: got %d, want %d", p, uint32(0xFFFFFFFF))
	}
	if l != 0xFFFFFFFF {
		t.Errorf("max len: got %d, want %d", l, uint32(0xFFFFFFFF))
	}
	_ = packed
}
