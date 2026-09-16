package push

import (
	"encoding/json"
	"io"
)

// readServiceReason は既知のエラーコードだけを取り出し、本文や購読URLをログへ渡さない。
func readServiceReason(body io.Reader) (string, error) {
	data, err := io.ReadAll(io.LimitReader(body, maxErrorBodyBytes+1))
	if err != nil {
		return "response_read_failed", err
	}
	if int64(len(data)) > maxErrorBodyBytes {
		return "response_too_large", nil
	}

	var response struct {
		Reason string `json:"reason"`
		Error  struct {
			Status string `json:"status"`
		} `json:"error"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return "unknown", nil
	}

	// Apple Web Push の reason と FCM の error.status を記録する。
	// 任意文字列へのフォールバックは、機密情報や改行を出力しないため行わない。
	for _, reason := range []string{response.Reason, response.Error.Status} {
		switch reason {
		case "BadTtl", "BadUrgency", "BadWebPushRequest", "BadWebPushTopic",
			"VapidPkHashMismatch", "IdleTimeout", "BadAuthorizationHeader",
			"BadJwtToken", "BadVapidPublicKey", "BadPath", "MethodNotAllowed",
			"PayloadTooLarge", "TooManyRequests", "InternalServerError",
			"ServiceUnavailable", "Shutdown", "Unregistered",
			"INVALID_ARGUMENT", "UNAUTHENTICATED", "PERMISSION_DENIED",
			"NOT_FOUND", "RESOURCE_EXHAUSTED", "INTERNAL", "UNAVAILABLE":
			return reason, nil
		}
	}
	return "unknown", nil
}
