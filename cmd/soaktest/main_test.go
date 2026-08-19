package main

import (
	"bytes"
	"testing"
	"time"
)

func TestRandomPayload(t *testing.T) {
	p1 := randomPayload(64)
	p2 := randomPayload(64)
	if len(p1) != 64 || len(p2) != 64 {
		t.Fatalf("expected payload length 64, got %d and %d", len(p1), len(p2))
	}
	if bytes.Equal(p1, p2) {
		t.Fatalf("expected two random payloads to differ")
	}
}

func TestChurn_Lifecycle(t *testing.T) {
	stop := make(chan struct{})
	churn(2, stop)
	time.Sleep(50 * time.Millisecond)
	close(stop)
	time.Sleep(50 * time.Millisecond)
}
