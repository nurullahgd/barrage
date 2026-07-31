package ui

import (
	"testing"
	"time"

	"github.com/nurullahgd/barrage.git/internal/client"
)

func TestModel_Update_ResultMsg(t *testing.T) {
	m := NewModel("http://localhost", 5, 0, 100)

	msg := ResultMsg(client.Result{
		StatusCode: 200,
		Latency:    5 * time.Millisecond,
	})
	updated, _ := m.Update(msg)
	m = updated.(Model)

	if m.Total != 1 {
		t.Errorf("Total = %d, want 1", m.Total)
	}
	if m.Successful != 1 {
		t.Errorf("Successful = %d, want 1", m.Successful)
	}
	if m.Failed != 0 {
		t.Errorf("Failed = %d, want 0", m.Failed)
	}
	if len(m.Latencies) != 1 {
		t.Errorf("Latencies len = %d, want 1", len(m.Latencies))
	}
}

func TestModel_Update_FailedRequest(t *testing.T) {
	m := NewModel("http://localhost", 5, 0, 100)

	msg := ResultMsg(client.Result{
		StatusCode: 500,
		Category:   client.CategoryServerError,
		Latency:    10 * time.Millisecond,
	})
	updated, _ := m.Update(msg)
	m = updated.(Model)

	if m.Total != 1 {
		t.Errorf("Total = %d, want 1", m.Total)
	}
	if m.Successful != 0 {
		t.Errorf("Successful = %d, want 0", m.Successful)
	}
	if m.Failed != 1 {
		t.Errorf("Failed = %d, want 1", m.Failed)
	}
}

func TestModel_Update_DoneMsg(t *testing.T) {
	m := NewModel("http://localhost", 5, 0, 100)
	updated, cmd := m.Update(DoneMsg{})
	m = updated.(Model)

	if !m.Done {
		t.Error("Done = false, want true")
	}
	if cmd == nil {
		t.Error("expected tea.Quit command, got nil")
	}
}
