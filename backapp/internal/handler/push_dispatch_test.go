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

func TestDispatchPushBatchLogsServiceReason(t *testing.T) {
	const endpoint = "https://web.push.apple.com/push/sensitive-capability-token"
	subscription := models.PushSubscription{UserID: "user-1", Endpoint: endpoint}
	sender := sensitiveErrorSender{result: push.Result{
		Subscription:  subscription,
		StatusCode:    403,
		ServiceReason: "BadJwtToken",
	}}

	var logs bytes.Buffer
	previousWriter := log.Writer()
	log.SetOutput(&logs)
	t.Cleanup(func() { log.SetOutput(previousWriter) })

	dispatchPushBatch(sender, nil, []byte(`{"title":"test"}`), []models.PushSubscription{subscription}, 60, "notification")
	output := logs.String()
	for _, want := range []string{"[notification]", push.EndpointLogID(endpoint), "status=403", `reason="BadJwtToken"`} {
		if !strings.Contains(output, want) {
			t.Fatalf("log is missing %q: %s", want, output)
		}
	}
	if strings.Contains(output, "sensitive-capability-token") {
		t.Fatal("capability token was written to logs")
	}
}
