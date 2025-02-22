package attributedstring

import (
	"bytes"
	"fmt"

	"github.com/BurntSushi/toml"
)

// Slice allows for extending a TOML string array with custom
// attributes that control how the array is marshaled into a Go string.
//
// Specifically, an Slice can be configured to avoid it being
// overridden by a subsequent unmarshal sequence.  When the `append` attribute
// is specified, the array will be appended instead (e.g., `array=["9",
// {append=true}]`).
type Slice struct { // A "mixed-type array" in TOML.
	// Note that the fields below _must_ be exported.  Otherwise the TOML
	// encoder would fail during type reflection.
	Values     []string
	Attributes struct { // Using a struct allows for adding more attributes in the future.
		Append *bool // Nil if not set by the user
	}
}

// Get returns the Slice values or an empty string slice.
func (a *Slice) Get() []string {
	if a.Values == nil {
		return []string{}
	}
	return a.Values
}

// UnmarshalTOML is the custom unmarshal method for Slice.
func (a *Slice) UnmarshalTOML(data interface{}) error {
	iFaceSlice, ok := data.([]interface{})
	if !ok {
		return fmt.Errorf("unable to cast to interface array: %v", data)
	}

	var loadedStrings []string
	for _, x := range iFaceSlice { // Iterate over each item in the slice.
		switch val := x.(type) {
		case string: // Strings are directly appended to the slice.
			loadedStrings = append(loadedStrings, val)
		case map[string]interface{}: // The attribute struct is represented as a map.
			for k, v := range val { // Iterate over all _supported_ keys.
				switch k {
				case "append":
					boolVal, ok := v.(bool)
					if !ok {
						return fmt.Errorf("unable to cast append to bool: %v", k)
					}
					a.Attributes.Append = &boolVal
				default: // Unsupported map key.
					return fmt.Errorf("unsupported key %q in map: %v", k, val)
				}
			}
		default: // Unsupported item.
			return fmt.Errorf("unsupported item in attributed string slice: %v", x)
		}
	}

	if a.Attributes.Append != nil && *a.Attributes.Append { // If _explicitly_ configured, append the loaded slice.
		a.Values = append(a.Values, loadedStrings...)
	} else { // Default: override the existing Slice.
		a.Values = loadedStrings
	}
	return nil
}

// MarshalTOML is the custom marshal method for Slice.
func (a *Slice) MarshalTOML() ([]byte, error) {
	iFaceSlice := make([]interface{}, 0, len(a.Values))

	for _, x := range a.Values {
		iFaceSlice = append(iFaceSlice, x)
	}

	if a.Attributes.Append != nil {
		Attributes := make(map[string]any)
		Attributes["append"] = *a.Attributes.Append
		iFaceSlice = append(iFaceSlice, Attributes)
	}

	buf := new(bytes.Buffer)
	enc := toml.NewEncoder(buf)
	if err := enc.Encode(iFaceSlice); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
