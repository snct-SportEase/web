package handler

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNotificationRequestLengthLimitsUseRunes(t *testing.T) {
	assert.LessOrEqual(t, len([]rune(strings.Repeat("あ", maxNotificationRequestTitleRunes))), maxNotificationRequestTitleRunes)
	assert.Greater(t, len([]rune(strings.Repeat("あ", maxNotificationMessageRunes+1))), maxNotificationMessageRunes)
}
