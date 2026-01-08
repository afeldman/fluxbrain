package analysis

// MockErrorbrainInput represents the input structure for Errorbrain
// This is a placeholder until the real Errorbrain library is integrated
type MockErrorbrainInput struct {
	Source  string
	Payload []byte
}

// MockErrorbrainResult represents the analysis result from Errorbrain
// This is a placeholder until the real Errorbrain library is integrated
type MockErrorbrainResult struct {
	Summary         string
	RootCause       string
	Recommendations []string
	RetrySafe       bool
	Confidence      float64
}

// MockErrorbrainAnalyzer is a placeholder for the real Errorbrain analyzer
// Remove this when integrating the actual github.com/afeldman/errorbrain library
type MockErrorbrainAnalyzer interface {
	Analyze(input MockErrorbrainInput) (MockErrorbrainResult, error)
}
