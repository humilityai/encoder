package encoder

import (
	"testing"
)

func TestNewOrdinal(t *testing.T) {
	encoder := NewOrdinal(false)
	if encoder == nil {
		t.Error("new encoder not created")
	}

	e2 := NewOrdinal(true)
	if e2 == nil {
		t.Error("new encoder with init true was not created")
	}
}

func TestOrdinalEncode(t *testing.T) {
	encoder := NewOrdinal(false)
	code := encoder.Encode("hello world")
	if code != uint64(0) {
		t.Errorf("code was %d and not 0", code)
	}

	e2 := NewOrdinal(true)
	c2 := e2.Encode("hello world")
	if c2 != uint64(1) {
		t.Errorf("code 2 was %d and not 1", code)
	}
}

func TestOrdinalDecode(t *testing.T) {
	encoder := NewOrdinal(false)
	value := "hello world"
	code := encoder.Encode(value)
	if encoder.Decode(code) != value {
		t.Error("decoded value did not equal original value")
	}

	e2 := NewOrdinal(true)
	c2 := e2.Encode(value)
	if e2.Decode(c2) != value {
		t.Error("decoded value did not equal original value")
	}
}

func TestOrdinalJSON(t *testing.T) {
	encoder := NewOrdinal(false)
	value := "hello world"
	code := encoder.Encode(value)
	data, err := encoder.MarshalJSON()
	if err != nil {
		t.Errorf("json marshal error: %+v", err)
	}

	newEncoder := NewOrdinal(false)
	err = newEncoder.UnmarshalJSON(data)
	if err != nil {
		t.Errorf("json unmarshal error: %+v", err)
	}

	if newEncoder.Decode(code) != value {
		t.Error("decoded value did not equal original value")
	}
}

func TestOrdinalCSV(t *testing.T) {
	encoder := NewOrdinal(false)
	value := "hello world"
	code := encoder.Encode(value)
	data, err := encoder.MarshalCSV()
	if err != nil {
		t.Errorf("json marshal error: %+v", err)
	}

	newEncoder := NewOrdinal(false)
	err = newEncoder.UnmarshalCSV(data)
	if err != nil {
		t.Errorf("json unmarshal error: %+v", err)
	}

	if newEncoder.Decode(code) != value {
		t.Error("decoded value did not equal original value")
	}
}

func TestOrdinalGob(t *testing.T) {
	encoder := NewOrdinal(false)
	value := "hello world"
	code := encoder.Encode(value)
	data, err := encoder.GobEncode()
	if err != nil {
		t.Errorf("json marshal error: %+v", err)
	}

	newEncoder := NewOrdinal(false)
	err = newEncoder.GobDecode(data)
	if err != nil {
		t.Errorf("json unmarshal error: %+v", err)
	}

	if newEncoder.Decode(code) != value {
		t.Error("decoded value did not equal original value")
	}
}
