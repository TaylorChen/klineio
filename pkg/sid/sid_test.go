package sid

import (
	"testing"
)

func TestSid_GenUint64(t *testing.T) {
	sid := NewSid()
	id, err := sid.GenUint64()
	if err != nil {
		t.Errorf("GenUint64 error: %v", err)
	}
	if id == 0 {
		t.Errorf("GenUint64 returned 0")
	}
}

func TestSid_GenString(t *testing.T) {
	sid := NewSid()
	str, err := sid.GenString()
	if err != nil {
		t.Errorf("GenString error: %v", err)
	}
	if str == "" {
		t.Errorf("GenString returned empty string")
	}
}
