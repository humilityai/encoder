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
	"hash/fnv"
	"io"
	"strconv"
	"sync"

	"github.com/humilityai/sam"
)

// Ordinal will encode string values into
// a unique integer value.
// The empty string is ALWAYS the 0 value.
// It will also allow for string values to be decoded.
type Ordinal struct {
	encoder map[uint64]uint64
	decoder sam.SliceString
	*sync.RWMutex
}

// NewOrdinal will create a new ordinal encoder.
// If the `init` boolean is specified as true,
// then the encoder will intialize with the
// empty string `""` encoded as the `0` value.
func NewOrdinal(init bool) *Ordinal {
	e := &Ordinal{
		encoder: make(map[uint64]uint64),
		decoder: make(sam.SliceString, 0),
		RWMutex: &sync.RWMutex{},
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
	e.RLock()
	defer e.RUnlock()

	hasher := fnv.New64a()
	_, err := hasher.Write([]byte(s))
	if err != nil {
		return false
	}
	hashedKey := hasher.Sum64()

	_, ok := e.encoder[hashedKey]
	return ok
}

// ContainsCode ...
func (e *Ordinal) ContainsCode(code int) bool {
	e.RLock()
	defer e.RUnlock()

	if len(e.decoder) >= code {
		return false
	}

	return true
}

// Encode ...
func (e *Ordinal) Encode(s string) uint64 {
	e.Lock()
	defer e.Unlock()

	hasher := fnv.New64a()
	_, err := hasher.Write([]byte(s))
	if err != nil {
		return 0
	}
	hashedKey := hasher.Sum64()

	v, ok := e.encoder[hashedKey]
	if !ok {
		code := uint64(len(e.decoder))
		e.decoder = append(e.decoder, s)
		e.encoder[hashedKey] = code
		return code
	}

	return v
}

// EncodeStringer --
func (e *Ordinal) EncodeStringer(s fmt.Stringer) uint64 {
	return e.Encode(s.String())
}

// EncodeBytes --
func (e *Ordinal) EncodeBytes(b []byte) uint64 {
	return e.Encode(string(b[:]))
}

// Decode will return an empty string if supplied integer
// argument is not a valid code.
func (e *Ordinal) Decode(i uint64) string {
	e.RLock()
	defer e.RUnlock()

	if i > uint64(len(e.decoder)-1) || i < 0 {
		return ""
	}

	return e.decoder[i]
}

// DecodeSlice will decode all the values in
// the slice of integers provided as an argument.
// If a string value has no existing encoding then
// it will be returned as the empty string.
func (e *Ordinal) DecodeSlice(s sam.SliceInt) sam.SliceString {
	e.RLock()
	defer e.RUnlock()

	values := make(sam.SliceString, len(s), len(s))
	for i, v := range s {
		values[i] = e.Decode(uint64(v))
	}

	return values
}

// EncodeSlice will encode all the values in the slice of strings
// provided as an argument.
func (e *Ordinal) EncodeSlice(s sam.SliceString) []uint64 {
	e.Lock()
	defer e.Unlock()

	codes := make([]uint64, len(s), len(s))
	for i, v := range s {
		codes[i] = e.Encode(v)
	}

	return codes
}

// Length ...
func (e *Ordinal) Length() int {
	e.RLock()
	defer e.RUnlock()

	return len(e.decoder)
}

// List ...
func (e *Ordinal) List() sam.SliceString {
	e.RLock()
	defer e.RUnlock()

	return e.decoder
}

// MarshalJSON ...
func (e *Ordinal) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.decoder)
}

// UnmarshalJSON ...
func (e *Ordinal) UnmarshalJSON(data []byte) error {
	s := make(sam.SliceString, 0)
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	hasher := fnv.New64a()
	encoder := make(map[uint64]uint64)
	for idx, str := range s {
		_, err := hasher.Write([]byte(str))
		if err != nil {
			return err
		}
		hashedKey := hasher.Sum64()
		encoder[hashedKey] = uint64(idx)
		hasher.Reset()
	}

	e.encoder = encoder
	e.decoder = s

	return nil
}

// MarshalCSV ...
func (e *Ordinal) MarshalCSV() ([]byte, error) {
	var lines [][]string

	// header
	lines = append(lines, []string{"value", "code"})

	for idx, value := range e.decoder {
		line := []string{value, strconv.Itoa(idx)}
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
	r := csv.NewReader(bytes.NewReader(data))

	decoder := make(sam.SliceString, 0)
	for i := 0; ; i++ {
		if i == 0 {
			_, err := r.Read()
			if err != nil && err != io.EOF {
				return err
			}
			continue
		}

		line, err := r.Read()
		if err == io.EOF {
			break
		} else if err == nil {
			code, err := strconv.Atoi(line[1])
			if err == nil {
				hasher := fnv.New64a()
				_, err := hasher.Write([]byte(line[0]))
				if err != nil {
					return err
				}
				hashedKey := hasher.Sum64()
				e.encoder[hashedKey] = uint64(code)
				if code > len(decoder)-1 {
					newCap := len(decoder) + (code - (len(decoder) - 1))
					newArray := make(sam.SliceString, newCap, newCap)
					copy(newArray, decoder)
					decoder = newArray
				}
				decoder[code] = line[0]
			} else {
				return err
			}
		}
	}

	e.decoder = decoder

	return nil
}

// GobEncode ...
func (e *Ordinal) GobEncode() ([]byte, error) {
	e.Lock()
	defer e.Unlock()

	var buf bytes.Buffer

	enc := gob.NewEncoder(&buf)

	eCopy := struct {
		Encoder map[uint64]uint64
		Decoder []string
	}{
		Encoder: e.encoder,
		Decoder: e.decoder,
	}

	err := enc.Encode(eCopy)
	if err != nil {
		return []byte{}, err
	}

	return buf.Bytes(), nil
}

// GobDecode ...
func (e *Ordinal) GobDecode(data []byte) error {
	var buf bytes.Buffer
	_, err := buf.Write(data)
	if err != nil {
		return err
	}

	var eCopy struct {
		Encoder map[uint64]uint64
		Decoder []string
	}

	dec := gob.NewDecoder(&buf)
	err = dec.Decode(&eCopy)
	if err != nil {
		return err
	}

	e.encoder = eCopy.Encoder
	e.decoder = sam.SliceString(eCopy.Decoder)
	return nil
}
