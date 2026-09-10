package service

import "testing"

func TestValidatePetAvatarRange(t *testing.T) {
	for _, id := range []int16{1, 12} {
		if err := validatePetAvatar(id); err != nil {
			t.Fatalf("avatar %d should be valid: %v", id, err)
		}
	}
	for _, id := range []int16{0, 13} {
		if err := validatePetAvatar(id); err == nil {
			t.Fatalf("avatar %d should be invalid", id)
		}
	}
}
