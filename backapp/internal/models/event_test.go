package models

import "testing"

func TestIsValidEventStatus(t *testing.T) {
	for _, status := range []string{
		EventStatusPreparing,
		EventStatusTesting,
		EventStatusUpcoming,
		EventStatusActive,
		EventStatusArchived,
	} {
		if !IsValidEventStatus(status) {
			t.Errorf("expected %q to be a valid event status", status)
		}
	}

	if IsValidEventStatus("invalid") {
		t.Error("expected an unknown event status to be rejected")
	}
}
