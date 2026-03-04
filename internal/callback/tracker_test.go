package callback

import (
	"testing"
)

func TestTrackerRecord(t *testing.T) {
	tracker := NewTracker()
	tracker.Record(DeliveryEvent{
		PayloadID:  "p1",
		RemoteAddr: "10.0.0.1:4444",
		Method:     "https",
		Size:       1024,
		Success:    true,
	})

	events := tracker.List()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].PayloadID != "p1" {
		t.Errorf("expected p1, got %s", events[0].PayloadID)
	}
	// RemoteAddr should be stripped of port
	if events[0].RemoteAddr != "10.0.0.1" {
		t.Errorf("expected 10.0.0.1, got %s", events[0].RemoteAddr)
	}
	if events[0].Timestamp.IsZero() {
		t.Error("timestamp should be set")
	}
}

func TestTrackerListCopy(t *testing.T) {
	tracker := NewTracker()
	tracker.Record(DeliveryEvent{PayloadID: "p1", RemoteAddr: "1.2.3.4"})

	events := tracker.List()
	events[0].PayloadID = "modified"

	// Original should not be modified
	original := tracker.List()
	if original[0].PayloadID != "p1" {
		t.Error("List should return a copy, not a reference to internal slice")
	}
}

func TestTrackerMaxEvents(t *testing.T) {
	tracker := NewTracker()
	for i := 0; i < maxEvents+100; i++ {
		tracker.Record(DeliveryEvent{
			PayloadID:  "p1",
			RemoteAddr: "1.2.3.4",
		})
	}

	events := tracker.List()
	if len(events) != maxEvents {
		t.Errorf("expected %d events (capped), got %d", maxEvents, len(events))
	}
}

func TestTrackerRemoteAddrNoPort(t *testing.T) {
	tracker := NewTracker()
	tracker.Record(DeliveryEvent{
		PayloadID:  "p1",
		RemoteAddr: "10.0.0.1", // no port
	})

	events := tracker.List()
	if events[0].RemoteAddr != "10.0.0.1" {
		t.Errorf("expected 10.0.0.1, got %s", events[0].RemoteAddr)
	}
}
