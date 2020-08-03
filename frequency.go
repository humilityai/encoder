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

import "github.com/humilityai/sam"

// Frequency is a one-way encoder.
// You cannot decode Frequency values
// as some values may be encoded with the same
// numerical value.
type Frequency struct {
	encoder sam.MapStringInt
}

// RollingFrequency is a one-war encoder.
// You cannot decode RollingFrequency values
// as some values may be encoded with the same
// numerical code.
type RollingFrequency struct {
	window int
	codes  sam.SliceInt
}

// NewFrequency will return a frequency encoder
// with the given values encoded.
func NewFrequency(values []string) *Frequency {
	encoder := make(sam.MapStringInt)
	for _, v := range values {
		encoder.Increment(v)
	}

	return &Frequency{
		encoder: encoder,
	}
}

// NewRollingFrequency will create a codeword for every value in the list of values
// in the order of those values.
// The list of values supplied to this function should not be a unique list of categorical
// values.
// The list should contain all the individual observation values found in the dataset/sample.
func NewRollingFrequency(window int, values []string) *RollingFrequency {
	codes := make(sam.SliceInt, len(values), len(values))

	encoder := make(sam.MapStringInt)
	for i := 0; i < len(values); i++ {
		if i%window == 0 {
			// zero
			encoder = make(sam.MapStringInt)
		}
		encoder.Increment(values[i])
		codes[i] = encoder[values[i]]
	}

	return &RollingFrequency{
		codes:  codes,
		window: window,
	}
}

// Codes will return the list of codes generated
// for the list of values provided in the creation
// of the RollingFrequency encoder.
func (e *RollingFrequency) Codes() sam.SliceInt {
	return e.codes
}

// Get will return the code for the given index, according
// to the original slice of values provided in the construction
// of the RollingFrequency encoder.
func (e *RollingFrequency) Get(index int) (int, error) {
	if index < 0 || index > len(e.codes)-1 {
		return 0, ErrBounds
	}

	return e.codes[index], nil
}

// Window will return the window used when
// creating the RollingFrequency encoder.
func (e *RollingFrequency) Window() int {
	return e.window
}

// Get ...
func (e *Frequency) Get(s string) (int, bool) {
	v, ok := e.encoder[s]
	return v, ok
}
