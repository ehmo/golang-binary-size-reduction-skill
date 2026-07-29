package analyzer

import (
	"go/build/constraint"
	"reflect"
	"testing"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	t.Setenv("GOFLAGS", "")
	testAnalyzer := *Analyzer
	// analysistest only reads expectations from ignored Go files for the
	// standard build-tag analyzer names.
	testAnalyzer.Name = "buildtag"
	analysistest.Run(t, analysistest.TestData(), &testAnalyzer, "findings")
}

func TestAnalyzerConfiguration(t *testing.T) {
	if err := analysis.Validate([]*analysis.Analyzer{Analyzer}); err != nil {
		t.Fatalf("analysis.Validate: %v", err)
	}
}

func TestTagsFromConstraintLines(t *testing.T) {
	t.Parallel()

	got := tagsFromConstraintLines([]string{
		"//go:build (linux && enterprise) || (arm64 && feature_x)",
		"// +build darwin,debug",
		"//go:build go1.25 || goexperiment.greenteagc || arm64.v8.0",
	})
	want := []string{"debug", "enterprise", "feature_x"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("tagsFromConstraintLines() = %v, want %v", got, want)
	}
}

func TestCollectCustomTags(t *testing.T) {
	t.Parallel()

	expression, err := constraint.Parse("//go:build !production && (linux || custom)")
	if err != nil {
		t.Fatal(err)
	}
	got := make(map[string]struct{})
	collectCustomTags(expression, got)
	want := map[string]struct{}{"custom": {}, "production": {}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("collectCustomTags() = %v, want %v", got, want)
	}
}

func TestImportDiagnostic(t *testing.T) {
	t.Parallel()

	for _, importPath := range []string{
		"C",
		"plugin",
		"reflect",
		"text/template",
		"html/template",
		"time/tzdata",
		"net",
		"os/user",
	} {
		if _, _, ok := importDiagnostic(importPath); !ok {
			t.Errorf("importDiagnostic(%q) did not return a finding", importPath)
		}
	}

	if _, _, ok := importDiagnostic("fmt"); ok {
		t.Error(`importDiagnostic("fmt") returned a finding`)
	}
}

func TestEmbedPatterns(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		comment  string
		patterns string
		ok       bool
	}{
		{comment: "//go:embed asset.txt", patterns: "asset.txt", ok: true},
		{comment: "//go:embed\tasset.txt", patterns: "asset.txt", ok: true},
		{comment: "//go:embed", ok: true},
		{comment: "//go:embedder asset.txt", ok: false},
		{comment: "// go:embed asset.txt", ok: false},
	} {
		patterns, ok := embedPatterns(test.comment)
		if patterns != test.patterns || ok != test.ok {
			t.Errorf("embedPatterns(%q) = %q, %v; want %q, %v", test.comment, patterns, ok, test.patterns, test.ok)
		}
	}
}
