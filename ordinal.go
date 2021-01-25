// Copyright 2020 Humility AI Incorporated, All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package encoder

import (
	"bytes"
	"encoding/csv"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/humilityai/sam"
)

// Ordinal will encode string values into
// a unique integer value.
// The empty string is ALWAYS the 0 value.
// It will also allow for string values to be decoded.
type Ordinal struct {
	encoder map[string]uint
	decoder sam.SliceString
}

// NewOrdinal will create a new ordinal encoder.
// If the `init` boolean is specified as true,
// then the encoder will intialize with the
// empty string `""` encoded as the `0` value.
func NewOrdinal(init bool) *Ordinal {
	e := &Ordinal{
		encoder: make(map[string]uint),
		decoder: make(sam.SliceString, 0),
	}

	// set empty string as 0
	if init {
		e.Encode("")
	}

	return e
}

// Contains will return whether or not a string
// has been assigned an ordinal code or not.
func (e *Ordinal) Contains(s string) bool {
	_, ok := e.encoder[s]
	return ok
}

// ContainsCode ...
func (e *Ordinal) ContainsCode(code int) bool {
	if len(e.decoder) >= code {
		return false
	}

	return true
}

// Encode ...
func (e *Ordinal) Encode(s string) uint {
	v, ok := e.encoder[s]
	if !ok {
		code := uint(len(e.decoder))
		e.decoder = append(e.decoder, s)
		e.encoder[s] = code
		return code
	}

	return v
}

// Decode will return an empty string if supplied integer
// argument is not a valid code.
func (e *Ordinal) Decode(i uint) string {
	if i > uint(len(e.decoder)-1) || i < 0 {
		return ""
	}

	return e.decoder[i]
}

// DecodeSlice will decode all the values in
// the slice of integers provided as an argument.
// If a string value has no existing encoding then
// it will be returned as the empty string.
func (e *Ordinal) DecodeSlice(s sam.SliceInt) sam.SliceString {
	values := make(sam.SliceString, len(s), len(s))
	for i, v := range s {
		values[i] = e.Decode(uint(v))
	}

	return values
}

// EncodeSlice will encode all the values in the slice of strings
// provided as an argument.
func (e *Ordinal) EncodeSlice(s sam.SliceString) []uint {
	codes := make([]uint, len(s), len(s))
	for i, v := range s {
		codes[i] = e.Encode(v)
	}

	return codes
}

// Length ...
func (e *Ordinal) Length() int {
	return len(e.decoder)
}

// List ...
func (e *Ordinal) List() sam.SliceString {
	return e.decoder
}

// MarshalJSON ...
func (e *Ordinal) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.encoder)
}

// UnmarshalJSON ...
func (e *Ordinal) UnmarshalJSON(data []byte) error {
	m := make(map[string]uint)
	err := json.Unmarshal(data, &m)
	if err != nil {
		return err
	}

	decoder := make(sam.SliceString, len(m), len(m))
	for value, code := range m {
		if code >= uint(len(m)) || code < 0 {
			return fmt.Errorf("value %+v with code %+v falls outside bounds", value, code)
		}

		decoder[code] = value
	}

	e = &Ordinal{
		encoder: m,
		decoder: decoder,
	}

	return nil
}

// MarshalCSV ...
func (e *Ordinal) MarshalCSV() ([]byte, error) {
	var lines [][]string

	// header
	lines = append(lines, []string{"value", "code"})

	for value, code := range e.encoder {
		line := []string{value, strconv.Itoa(int(code))}
		lines = append(lines, line)
	}

	var b bytes.Buffer
	w := csv.NewWriter(&b)
	err := w.WriteAll(lines)
	if err != nil {
		return []byte{}, err
	}

	return b.Bytes(), nil
}

// UnmarshalCSV ...
func (e *Ordinal) UnmarshalCSV(data []byte) error {
	var b bytes.Buffer
	_, err := b.Write(data)
	if err != nil {
		return err
	}

	r := csv.NewReader(&b)
	lines, err := r.ReadAll()
	if err != nil {
		return err
	}

	s := lines[1:]
	decoder := make(sam.SliceString, len(s), len(s))
	for _, line := range s {
		if len(line) == 2 {
			code, err := strconv.Atoi(line[1])
			if err == nil {
				e.encoder[line[0]] = uint(code)
				decoder[code] = line[0]
			}
		}
	}
	e.decoder = decoder

	return nil
}

// MarshalBinary ...
func (e *Ordinal) MarshalBinary() ([]byte, error) {
	var b bytes.Buffer
	fmt.Fprintln(&b, e)
	return b.Bytes(), nil
}

// UnmarshalBinary ...
func (e *Ordinal) UnmarshalBinary(data []byte) error {
	b := bytes.NewBuffer(data)
	_, err := fmt.Fscanln(b, e)
	return err
}

// MarshalGob ...
func (e *Ordinal) MarshalGob() ([]byte, error) {
	var buf bytes.Buffer

	enc := gob.NewEncoder(&buf)
	err := enc.Encode(e)
	if err != nil {
		return []byte{}, err
	}

	return buf.Bytes(), nil
}

// UnmarshalGob ...
func (e *Ordinal) UnmarshalGob(data []byte) error {
	var buf bytes.Buffer
	_, err := buf.Write(data)
	if err != nil {
		return err
	}

	dec := gob.NewDecoder(&buf)
	return dec.Decode(e)
}
