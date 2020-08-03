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
	"encoding/json"
	"strconv"

	"github.com/humilityai/sam"
)

// OneHot will encode string values into
// a unique one-hot vector (binary vector with a single 1).
// The empty string is ALWAYS the 0-vector.
// It will also allow for string values to be decoded.
type OneHot struct {
	encoder sam.MapStringInt
	decoder sam.SliceString
}

// NewOneHot will return a one-hot encoder
// that will set the empty string as the first
// dimension of every one-hot binary codeword.
// "Binary" here means that every value in the
// codeword (integer slice) will be either a 0
// or a 1.
func NewOneHot() *OneHot {
	e := &OneHot{
		encoder: make(sam.MapStringInt),
		decoder: make(sam.SliceString, 0),
	}

	// set empty string as first dimension
	e.Encode("")

	return e
}

// Encode will return the integer slice that represents
// the binary encoding of the given string argument.
// If the string argument does not already have a code
// it will generate a new codeword for the given string
// argument and add it to the encoder.
func (e *OneHot) Encode(s string) []uint8 {
	_, ok := e.encoder[s]
	if !ok {
		e.decoder = append(e.decoder, s)
		e.encoder[s] = len(e.decoder)

		return e.code(s)
	}

	return e.code(s)
}

// Decode will return the string for the given binary
// codeword (one-hot code).
// If the codeword argument is longer than the encoders codewords
// then an `ErrLength` error will be returned.
func (e *OneHot) Decode(code []uint8) (string, error) {
	if len(code) > len(e.decoder) {
		return "", ErrLength
	}

	var dim int
	for i, v := range code {
		if v == 1 {
			dim = i
			break
		}
	}

	return e.decoder[dim], nil
}

// Contains will check if a string has been assigned
// a one-hot code or not.
func (e *OneHot) Contains(s string) bool {
	_, ok := e.encoder[s]
	return ok
}

// ContainsCode will check if a codeword is a valid
// codeword or not.
func (e *OneHot) ContainsCode(code []uint8) bool {
	if len(e.decoder) > len(code) {
		return false
	}

	return containsOne(code)
}

// Dimension returns the current dimension of
// each one-hot codeword. The dimension increases
// with every new string that gets encoded.
func (e *OneHot) Dimension() int {
	return len(e.decoder)
}

// MarshalJSON ...
func (e *OneHot) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.decoder)
}

// UnmarshalJSON ...
func (e *OneHot) UnmarshalJSON(data []byte) error {
	s := make(sam.SliceString, 0)
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	encoder := make(sam.MapStringInt)
	for i, v := range s {
		encoder[v] = i + 1
	}

	e = &OneHot{
		encoder: encoder,
		decoder: s,
	}

	return nil
}

// MarshalCSV ...
func (e *OneHot) MarshalCSV() ([]byte, error) {
	var lines [][]string

	// header
	lines = append(lines, []string{"value", "code"})

	for value, code := range e.encoder {
		line := []string{value, strconv.Itoa(code)}
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
func (e *OneHot) UnmarshalCSV(data []byte) error {
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
				e.encoder[line[0]] = code
				decoder[code-1] = line[0]
			}
		}
	}
	e.decoder = decoder

	return nil
}

func (e *OneHot) code(s string) (code []uint8) {
	code = make([]uint8, len(e.decoder), len(e.decoder))
	dim := e.encoder[s]

	code[dim] = 1
	return
}

func containsOne(code []uint8) bool {
	contains := false
	for _, v := range code {
		if v == 1 {
			if contains {
				return false
			}

			contains = true
		}
	}

	return contains
}
