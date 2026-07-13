package handler

import (
	"bytes"
	"context"
	"errors"
	"log"
	"strings"
	"testing"

	"backapp/internal/models"
	"backapp/internal/push"
)

type sensitiveErrorSender struct {
	result push.Result
}

func (s sensitiveErrorSender) Enabled() bool {
	return true
}

func (s sensitiveErrorSender) ValidateSubscription(string, string, string) error {
	return nil
}

func (s sensitiveErrorSender) SendBatch(context.Context, []byte, []models.PushSubscription, int) []push.Result {
	return []push.Result{s.result}
}

func TestDispatchPushBatchDoesNotLogCapabilityURLOrErrorBody(t *testing.T) {
	const endpoint = "https://fcm.googleapis.com/fcm/send/sensitive-capability-token"
	const responseSecret = "push-provider-response-secret"
	subscription := models.PushSubscription{UserID: "user-1", Endpoint: endpoint}
	sender := sensitiveErrorSender{result: push.Result{
		Subscription: subscription,
		Err:          errors.New("POST " + endpoint + " failed: " + responseSecret),
	}}

	var logs bytes.Buffer
	previousWriter := log.Writer()
	log.SetOutput(&logs)
	t.Cleanup(func() {
		log.SetOutput(previousWriter)
	})

	dispatchPushBatch(sender, nil, []byte(`{"title":"test"}`), []models.PushSubscription{subscription}, 60, "test-push")

	logOutput := logs.String()
	for _, secret := range []string{endpoint, "sensitive-capability-token", responseSecret} {
		if strings.Contains(logOutput, secret) {
			t.Fatalf("sensitive Push value %q was written to logs", secret)
		}
	}
	if endpointID := push.EndpointLogID(endpoint); !strings.Contains(logOutput, endpointID) {
		t.Fatalf("expected safe endpoint correlation ID %q, got %q", endpointID, logOutput)
	}
}
