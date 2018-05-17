package registry

import "github.com/tiglabs/baudengine/kernel/analysis"

var analyzers *Registry

func RegisterAnalyzer(name string, analyzer analysis.Analyzer) {
	analyzers.RegisterAnalyzer(name, analyzer)
}

func GetAnalyzer(name string) analysis.Analyzer {
	return analyzers.GetAnalyzer(name)
}

func RegisterTokenizer(name string, tokenizer analysis.Tokenizer) {
	analyzers.RegisterTokenizer(name, tokenizer)
}

func GetTokenizer(name string) analysis.Tokenizer {
	return analyzers.GetTokenizer(name)
}

type Registry struct {
	analyzers map[string]analysis.Analyzer
	tokenizer map[string]analysis.Tokenizer
}

func NewRegistry() *Registry {
	return &Registry{
		analyzers: make(map[string]analysis.Analyzer),
		tokenizer: make(map[string]analysis.Tokenizer),}
}

func (r *Registry) RegisterAnalyzer(name string, analyzer analysis.Analyzer) {
	if _, ok := r.analyzers[name]; ok {
		// TODO panic ??
		return
	}
	r.analyzers[name] = analyzer
}

func (r *Registry) GetAnalyzer(name string) analysis.Analyzer {
	if ar, ok := r.analyzers[name]; ok {
		return ar
	}
	// TODO panic ??
	return nil
}

func (r *Registry) RegisterTokenizer(name string, tokenizer analysis.Tokenizer) {
	if _, ok := r.tokenizer[name]; ok {
		// TODO panic ??
		return
	}
	r.tokenizer[name] = tokenizer
}

func (r *Registry) GetTokenizer(name string) analysis.Tokenizer {
	if ar, ok := r.tokenizer[name]; ok {
		return ar
	}
	// TODO panic ??
	return nil
}

func init() {
	analyzers = NewRegistry()
}