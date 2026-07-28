package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nurullahgd/barrage.git/internal/client"
	"github.com/nurullahgd/barrage.git/internal/stats"
)

var (
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	labelStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	valueStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

type Model struct {
	Total       int
	Successful  int
	Failed      int
	Latencies   []time.Duration
	StartTime   time.Time
	Done        bool
	TargetURL   string
	Workers     int
	Rate        int
	TotalTarget int
}

type ResultMsg client.Result
type DoneMsg struct{}

func NewModel(url string, workers, rate, total int) Model {
	return Model{
		StartTime:   time.Now(),
		TargetURL:   url,
		Workers:     workers,
		Rate:        rate,
		TotalTarget: total,
		Latencies:   make([]time.Duration, 0, total),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ResultMsg:
		r := client.Result(msg)
		m.Total++
		if r.Error == nil && r.StatusCode >= 200 && r.StatusCode < 300 {
			m.Successful++
		} else {
			m.Failed++
		}
		m.Latencies = append(m.Latencies, r.Latency)
		return m, nil
	case DoneMsg:
		m.Done = true
		return m, tea.Quit
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.Done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	elapsed := time.Since(m.StartTime).Round(time.Millisecond)

	rps := 0.0
	if elapsed.Seconds() > 0 {
		rps = float64(m.Total) / elapsed.Seconds()
	}

	successRate := 0.0
	if m.Total > 0 {
		successRate = float64(m.Successful) / float64(m.Total) * 100
	}

	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString(titleStyle.Render(" ⚡ barrage"))
	sb.WriteString("\n\n")

	// config satırı
	sb.WriteString(dimStyle.Render(fmt.Sprintf("  %s   workers:%d",
		m.TargetURL, m.Workers)))
	if m.Rate > 0 {
		sb.WriteString(dimStyle.Render(fmt.Sprintf("   rate:%d/s", m.Rate)))
	}
	sb.WriteString("\n\n")

	// progress bar
	sb.WriteString(m.progressBar())
	sb.WriteString("\n\n")

	// metrikler
	sb.WriteString(fmt.Sprintf("  %s %s\n",
		labelStyle.Render("Requests/sec"),
		valueStyle.Render(fmt.Sprintf("%.1f", rps))))

	successStr := fmt.Sprintf("%d (%.1f%%)", m.Successful, successRate)
	sb.WriteString(fmt.Sprintf("  %s %s\n",
		labelStyle.Render("Successful  "),
		successStyle.Render(successStr)))

	failedStr := fmt.Sprintf("%d", m.Failed)
	failStyle := successStyle
	if m.Failed > 0 {
		failStyle = errorStyle
	}
	sb.WriteString(fmt.Sprintf("  %s %s\n",
		labelStyle.Render("Failed      "),
		failStyle.Render(failedStr)))

	sb.WriteString("\n")

	// latency
	sb.WriteString(labelStyle.Render("  Latency\n"))
	sb.WriteString(fmt.Sprintf("    %s  %s\n",
		labelStyle.Render("p50"),
		valueStyle.Render(m.p(50).Round(time.Microsecond*10).String())))
	sb.WriteString(fmt.Sprintf("    %s  %s\n",
		labelStyle.Render("p95"),
		valueStyle.Render(m.p(95).Round(time.Microsecond*10).String())))
	sb.WriteString(fmt.Sprintf("    %s  %s\n",
		labelStyle.Render("p99"),
		valueStyle.Render(m.p(99).Round(time.Microsecond*10).String())))

	sb.WriteString("\n")
	sb.WriteString(dimStyle.Render(fmt.Sprintf("  elapsed %s", elapsed)))
	sb.WriteString("\n")

	if !m.Done {
		sb.WriteString(dimStyle.Render("  press ctrl+c to stop"))
		sb.WriteString("\n")
	}

	return sb.String()
}

func (m Model) progressBar() string {
	const width = 30
	if m.TotalTarget == 0 {
		return ""
	}

	pct := float64(m.Total) / float64(m.TotalTarget)
	if pct > 1 {
		pct = 1
	}
	filled := int(pct * width)
	empty := width - filled

	bar := successStyle.Render(strings.Repeat("█", filled)) +
		dimStyle.Render(strings.Repeat("░", empty))

	return fmt.Sprintf("  [%s]  %d/%d  (%.0f%%)",
		bar, m.Total, m.TotalTarget, pct*100)
}

func (m Model) p(percent float64) time.Duration {
	if len(m.Latencies) == 0 {
		return 0
	}
	sorted := make([]time.Duration, len(m.Latencies))
	copy(sorted, m.Latencies)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	return stats.Percentile(sorted, percent)
}
