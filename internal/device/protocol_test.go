package device

import "testing"

func TestFrameRoundTrip(t *testing.T) {
	f := buildFrame(0x01, "CH123456", []byte("secret"))
	got, err := parseFrame(f)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got.Cmd != 0x01 || got.DeviceNo != "CH123456" || string(got.Data) != "secret" {
		t.Fatalf("unexpected frame: %+v", got)
	}
}

func TestCRCMismatch(t *testing.T) {
	f := buildFrame(0x02, "CH1", []byte{1, 2, 3})
	f[len(f)-1] ^= 0xFF
	if _, err := parseFrame(f); err == nil {
		t.Fatal("expected crc error")
	}
}
