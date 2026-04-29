package arena

import "io"

// TraceFile is one decoded arena trace file selected for analysis.
type TraceFile struct {
	Name  string
	Trace TraceMatch
}

// TraceAnalysisInput is the generic input passed from the CLI to a
// game-specific analyzer.
type TraceAnalysisInput struct {
	TraceDir string
	Files    []TraceFile
}

// TraceAnalysis is a game-specific analysis report.
type TraceAnalysis interface {
	Write(io.Writer) error
}

// TraceAnalyzer is an optional game factory capability for trace analysis.
type TraceAnalyzer interface {
	AnalyzeTraces(TraceAnalysisInput) (TraceAnalysis, error)
}
