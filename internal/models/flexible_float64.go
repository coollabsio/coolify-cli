package models

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// FlexibleFloat64 accepts JSON numbers and quoted numeric strings.
type FlexibleFloat64 float64

func (f *FlexibleFloat64) UnmarshalJSON(data []byte) error {
	var value float64
	if len(data) > 0 && data[0] == '"' {
		var encoded string
		if err := json.Unmarshal(data, &encoded); err != nil {
			return err
		}

		parsed, err := strconv.ParseFloat(encoded, 64)
		if err != nil {
			return fmt.Errorf("invalid numeric string %q: %w", encoded, err)
		}
		value = parsed
	} else if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	*f = FlexibleFloat64(value)

	return nil
}
