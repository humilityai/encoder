package encoder

import (
	"github.com/humilityai/sam"
)

// JamesSteinRegression is a one way encoder.
// You cannot decode JamesSteinRegression values
// as some values may be encoded with the same
// numerical code.
// JamesSteinRegression is a target-based encoder.
type JamesSteinRegression struct {
	encoder map[string]float64
}

// JamesSteinClassification is a one way encoder.
// You cannot decode JamesSteinClassification values
// as some values may be encoded with the same
// numerical code.
// JamesSteinClassification is a target-based encoder.
type JamesSteinClassification struct {
	encodedValues sam.SliceFloat64
}

// NewJamesSteinRegression will create a JamesSteinRegression encoder
func NewJamesSteinRegression(values []string, target []float64) (*JamesSteinRegression, error) {
	if len(target) != len(values) {
		return &JamesSteinRegression{}, ErrTargetLength
	}

	targetValues := make(map[string]sam.SliceFloat64)
	for i := 0; i < len(values); i++ {
		targetValues[values[i]] = append(targetValues[values[i]], target[i])
	}

	encoder := make(map[string]float64)
	for k, v := range targetValues {
		encoder[k] = v.Avg()
	}

	return &JamesSteinRegression{
		encoder: encoder,
	}, nil
}

// NewJamesSteinClassification will create a JamesSteinClassification encoder
func NewJamesSteinClassification(values []string, target []string) (*JamesSteinClassification, error) {
	if len(target) != len(values) {
		return &JamesSteinClassification{}, ErrTargetLength
	}

	groupCounts := make(sam.MapStringInt)
	classCounts := make(sam.MapStringInt)
	groupClassCounts := make(map[string]sam.MapStringInt)
	for i := 0; i < len(values); i++ {
		group := values[i]
		class := target[i]
		groupCounts.Increment(group)
		classCounts.Increment(class)
		groupClassCounts[group].Increment(class)
	}

	groupClassBValues := make(map[string]map[string]float64)
	for group, classCounts := range groupClassCounts {
		groupCount := groupCounts[group]
		for class, count := range classCounts {
			classCount := classCounts[class]
			groupClassPercentage := float64(count) / float64(classCount)
			classPercentage := float64(classCount) / float64(len(target))
			groupClassValue := (groupClassPercentage * (1 - groupClassPercentage)) / float64(groupCount)
			classValue := (classPercentage * (1 - classPercentage)) / float64(len(target))

			groupClassBValues[group][class] = groupClassValue / (groupClassValue + classValue)
		}
	}

	encodedValues := make(sam.SliceFloat64, len(target), len(target))
	for i := 0; i < len(values); i++ {
		group := values[i]
		class := target[i]
		encodedValues[i] = groupClassBValues[group][class]
	}

	return &JamesSteinClassification{
		encodedValues: encodedValues,
	}, nil
}

// Get will retrieve the code for the given categorical value.
func (e *JamesSteinRegression) Get(s string) (float64, bool) {
	v, ok := e.encoder[s]
	return v, ok
}

// Codes will return the slice of codes for all of the values
// used in the construction of the JamesSteinClassification encoder.
func (e *JamesSteinClassification) Codes() sam.SliceFloat64 {
	return e.encodedValues
}

// Get will retrieve the code for the given categorical value.
func (e *JamesSteinClassification) Get(index int) (float64, error) {
	if index < 0 || index > len(e.encodedValues)-1 {
		return 0, ErrBounds
	}

	return e.encodedValues[index], nil
}
