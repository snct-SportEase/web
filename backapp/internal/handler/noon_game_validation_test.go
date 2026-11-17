package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func intPointer(value int) *int {
	return &value
}

func TestValidateTemplateRankingPointOverrides(t *testing.T) {
	tests := []struct {
		name      string
		counts    map[int]int
		rankings  []yearRelayRankInput
		errorText string
	}{
		{
			name:     "accepts non-negative overrides for a tie",
			counts:   map[int]int{1: 2},
			rankings: []yearRelayRankInput{{Rank: 1, Points: intPointer(0)}, {Rank: 1, Points: intPointer(10)}},
		},
		{
			name:      "rejects a negative override",
			counts:    map[int]int{1: 2},
			rankings:  []yearRelayRankInput{{Rank: 1, Points: intPointer(-1)}, {Rank: 1, Points: intPointer(10)}},
			errorText: "points must not be negative",
		},
		{
			name:      "requires an override for every tied entry",
			counts:    map[int]int{1: 2},
			rankings:  []yearRelayRankInput{{Rank: 1}, {Rank: 1, Points: intPointer(10)}},
			errorText: "同順位があるため points の指定が必須です",
		},
		{
			name:      "rejects an override for a unique rank",
			counts:    map[int]int{1: 1},
			rankings:  []yearRelayRankInput{{Rank: 1, Points: intPointer(10)}},
			errorText: "同順位でない順位には points を指定できません",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateTemplateRankingPointOverrides(test.counts, test.rankings)
			if test.errorText == "" {
				assert.NoError(t, err)
				return
			}
			assert.EqualError(t, err, test.errorText)
		})
	}
}
