package stats

import (
	"testing"
	"time"

	"github.com/nurullahgd/barrage.git/internal/client"
)

func TestPercentile(t *testing.T) {
	tests := []struct {
		name   string
		sorted []int
		p      float64
		want   int
	}{
		{
			name:   "p0 returns minimum",
			sorted: []int{10, 20, 30, 40, 50},
			p:      0,
			want:   10,
		},
		{
			name:   "p100 returns maximum",
			sorted: []int{10, 20, 30, 40, 50},
			p:      100,
			want:   50,
		},
		{
			name:   "p50 returns median",
			sorted: []int{10, 20, 30, 40, 50},
			p:      50,
			want:   30,
		},
		{
			name:   "single element returns that element regardless of p",
			sorted: []int{42},
			p:      73,
			want:   42,
		},
		{
			name:   "zero value returns zero",
			sorted: []int{},
			p:      0,
			want:   0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Percentile(tt.sorted, tt.p)
			if got != tt.want {
				t.Errorf("Percentile(%v, %v) = %v, want %v", tt.sorted, tt.p, got, tt.want)
			}
		})
	}
}

func TestPercentile_Duration(t *testing.T) {
	sorted := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		30 * time.Millisecond,
	}
	got := Percentile(sorted, 100)
	want := 30 * time.Millisecond
	if got != want {
		t.Errorf("Percentile(%v, 100) = %v, want %v", sorted, got, want)
	}
}

func TestComputeStats(t *testing.T) {
	results := []client.Result{
		{StatusCode: 200, Latency: 10 * time.Millisecond},
		{StatusCode: 200, Latency: 20 * time.Millisecond},
		{StatusCode: 500, Latency: 30 * time.Millisecond, Category: client.CategoryServerError},
		{StatusCode: 400, Latency: 40 * time.Millisecond, Category: client.CategoryClientError},
		{StatusCode: 408, Latency: 50 * time.Millisecond, Category: client.CategoryTimeout},
	}

	got := ComputeStats(results, 1*time.Second)

	if got.TotalRequests != 5 {
		t.Errorf("TotalRequests = %d, want 5", got.TotalRequests)
	}
	if got.SuccessCount != 2 {
		t.Errorf("SuccessCount = %d, want 2", got.SuccessCount)
	}
	if got.ErrorCount != 3 {
		t.Errorf("ErrorCount = %d, want 3", got.ErrorCount)
	}
	if got.TimeoutCount != 1 {
		t.Errorf("TimeoutCount = %d, want 1", got.TimeoutCount)
	}
	if got.ClientErrorCount != 1 {
		t.Errorf("ClientErrorCount = %d, want 1", got.ClientErrorCount)
	}
	if got.ServerErrorCount != 1 {
		t.Errorf("ServerErrorCount = %d, want 1", got.ServerErrorCount)
	}
	if got.MinLatency != 10*time.Millisecond {
		t.Errorf("MinLatency = %v, want 10ms", got.MinLatency)
	}
	if got.MaxLatency != 50*time.Millisecond {
		t.Errorf("MaxLatency = %v, want 50ms", got.MaxLatency)
	}
	if got.AverageLatency != 30*time.Millisecond {
		t.Errorf("AverageLatency = %v, want 30ms", got.AverageLatency)
	}
	if got.RequestsPerSecond != 5 {
		t.Errorf("RequestsPerSecond = %v, want 5", got.RequestsPerSecond)
	}
}

func TestRound6(t *testing.T) {
	tests := []struct {
		input float64
		want  float64
	}{
		{0.0000001234567, 0.000000},
		{0.001234567, 0.001235},
		{1.9999999, 2.0},
	}
	for _, tt := range tests {
		got := round6(tt.input)
		if got != tt.want {
			t.Errorf("round6(%v) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestToJSON(t *testing.T) {
	s := Stats{
		TotalRequests:     10,
		SuccessCount:      9,
		ErrorCount:        1,
		TimeoutCount:      1,
		MinLatency:        1 * time.Millisecond,
		MaxLatency:        100 * time.Millisecond,
		AverageLatency:    50 * time.Millisecond,
		Percentile50:      45 * time.Millisecond,
		Percentile95:      90 * time.Millisecond,
		Percentile99:      99 * time.Millisecond,
		RequestsPerSecond: 100.0,
	}
	got := s.ToJSON()

	if got.TotalRequests != 10 {
		t.Errorf("TotalRequests = %d, want 10", got.TotalRequests)
	}
	if got.Latency.Min != round6(0.001) {
		t.Errorf("Latency.Min = %v, want %v", got.Latency.Min, round6(0.001))
	}
	if got.Latency.Max != round6(0.1) {
		t.Errorf("Latency.Max = %v, want %v", got.Latency.Max, round6(0.1))
	}
	if got.ErrorBreakdown.Timeouts != 1 {
		t.Errorf("Timeouts = %d, want 1", got.ErrorBreakdown.Timeouts)
	}
}

func TestComputeStats_Empty(t *testing.T) {
	got := ComputeStats([]client.Result{}, 1*time.Second)
	if got.TotalRequests != 0 {
		t.Errorf("TotalRequests = %d, want 0", got.TotalRequests)
	}
}
