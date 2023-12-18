package types

import (
	"encoding/json"
)

// TODO: IS this type still needed when im using jsonb

// AllowedCredentialIds is a custom type for representing the allowed credential IDs.
type AllowedCredentialIds [][]byte

// UnmarshalJSON implements the custom JSON unmarshaling for AllowedCredentialIds.
func (ids *AllowedCredentialIds) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		*ids = nil
		return nil
	}

	var rawIds [][]byte
	if err := json.Unmarshal(b, &rawIds); err != nil {
		return err
	}

	// Validate the individual credential IDs if needed.
	// Here, we simply copy the raw IDs to the target field.
	*ids = rawIds

	return nil
}
