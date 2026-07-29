// Package analyzer reports source patterns that can affect Go binary size.
package analyzer

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/build/constraint"
	"go/token"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const (
	categoryBuildTag   = "build-tag"
	categoryCGO        = "cgo"
	categoryEmbed      = "embed"
	categoryImport     = "import"
	categoryResolver   = "resolver"
	categorySideEffect = "side-effect-import"
	categoryTimezone   = "timezone"
	categoryUserLookup = "user-lookup"
)

const doc = `report source patterns that may affect Go binary size

The analyzer reports audit leads, not defects. Each finding requires a measured
before-and-after build and runtime checks before a source or build change is
accepted.`

// Analyzer reports source patterns worth reviewing during binary-size work.
var Analyzer = &analysis.Analyzer{
	Name:             "binsize",
	Doc:              doc,
	Run:              run,
	RunDespiteErrors: true,
}

type sourceConstraint struct {
	text string
	pos  token.Pos
	end  token.Pos
}

type rawConstraint struct {
	text string
	line int
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if isNonProductionFile(pass, file) {
			continue
		}
		reportBuildConstraint(pass, activeConstraint(file))
		reportImports(pass, file)
		reportEmbedDirectives(pass, file)
	}

	if pass.ReadFile == nil {
		return nil, nil
	}
	for _, filename := range pass.IgnoredFiles {
		if filepath.Ext(filename) != ".go" || strings.HasSuffix(filename, "_test.go") {
			continue
		}
		if err := reportIgnoredBuildConstraint(pass, filename); err != nil {
			return nil, fmt.Errorf("read ignored file %q: %w", filename, err)
		}
	}

	return nil, nil
}

func isNonProductionFile(pass *analysis.Pass, file *ast.File) bool {
	tokenFile := pass.Fset.File(file.Pos())
	if tokenFile == nil {
		return true
	}
	filename := tokenFile.Name()
	return filepath.Ext(filename) != ".go" || strings.HasSuffix(filename, "_test.go")
}

func reportImports(pass *analysis.Pass, file *ast.File) {
	for _, spec := range file.Imports {
		importPath, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			continue
		}

		if category, message, ok := importDiagnostic(importPath); ok {
			pass.Report(analysis.Diagnostic{
				Pos:      spec.Pos(),
				End:      spec.End(),
				Category: category,
				Message:  message,
			})
			continue
		}

		if spec.Name != nil && spec.Name.Name == "_" && importPath != "embed" {
			pass.Report(analysis.Diagnostic{
				Pos:      spec.Pos(),
				End:      spec.End(),
				Category: categorySideEffect,
				Message: fmt.Sprintf(
					"side-effect import %q pulls its initialization path into the binary; confirm the registration is required, then measure",
					importPath,
				),
			})
		}
	}
}

func importDiagnostic(importPath string) (category, message string, ok bool) {
	switch importPath {
	case "C":
		return categoryCGO, `import "C" enables cgo for this package; if cgo is required and C source is compiled, benchmark CGO_CFLAGS="-Oz"`, true
	case "plugin":
		return categoryImport, `import "plugin" can weaken dead-code elimination; isolate optional plugin support behind a build tag or separate binary, then measure`, true
	case "reflect":
		return categoryImport, `import "reflect" may retain extra type and method metadata; check whether reflection is required on the shipped path, then measure`, true
	case "text/template":
		return categoryImport, `import "text/template" adds runtime parsing and reflection; gate optional template support or compare generated output, then measure`, true
	case "html/template":
		return categoryImport, `import "html/template" adds runtime parsing and contextual escaping; gate optional template support and preserve escaping behavior in any replacement`, true
	case "time/tzdata":
		return categoryTimezone, `import "time/tzdata" embeds the timezone database; remove it only when the deployment supplies zoneinfo, then measure`, true
	case "net":
		return categoryResolver, `import "net" makes resolver behavior part of any cgo or netgo size experiment; test DNS before accepting CGO_ENABLED=0 with -tags netgo,osusergo`, true
	case "os/user":
		return categoryUserLookup, `import "os/user" makes user lookup behavior part of any cgo or osusergo size experiment; test user and group lookup before accepting CGO_ENABLED=0 with -tags netgo,osusergo`, true
	default:
		return "", "", false
	}
}

func reportEmbedDirectives(pass *analysis.Pass, file *ast.File) {
	for _, group := range file.Comments {
		for _, comment := range group.List {
			patterns, ok := embedPatterns(comment.Text)
			if !ok {
				continue
			}

			message := "go:embed adds matched file bytes to the binary; audit or externalize the payload, then measure"
			if patterns != "" {
				message = fmt.Sprintf(
					"go:embed pattern %q adds matched file bytes to the binary; audit or externalize the payload, then measure",
					patterns,
				)
			}

			pass.Report(analysis.Diagnostic{
				Pos:      comment.Pos(),
				End:      comment.End(),
				Category: categoryEmbed,
				Message:  message,
			})
		}
	}
}

func embedPatterns(comment string) (string, bool) {
	const prefix = "//go:embed"
	if comment == prefix {
		return "", true
	}
	if !strings.HasPrefix(comment, prefix) {
		return "", false
	}
	rest := strings.TrimPrefix(comment, prefix)
	if rest[0] != ' ' && rest[0] != '\t' {
		return "", false
	}
	return strings.TrimSpace(rest), true
}

func activeConstraint(file *ast.File) sourceConstraint {
	var goBuild, plusBuild []sourceConstraint
	for _, group := range file.Comments {
		for _, comment := range group.List {
			if comment.Pos() > file.Package {
				continue
			}
			line := strings.TrimSpace(comment.Text)
			item := sourceConstraint{text: line, pos: comment.Pos(), end: comment.End()}
			switch {
			case constraint.IsGoBuild(line):
				goBuild = append(goBuild, item)
			case constraint.IsPlusBuild(line):
				plusBuild = append(plusBuild, item)
			}
		}
	}
	return mergeConstraints(goBuild, plusBuild)
}

func reportIgnoredBuildConstraint(pass *analysis.Pass, filename string) error {
	content, err := pass.ReadFile(filename)
	if err != nil {
		return err
	}

	constraintLine := rawFileConstraint(content)
	if constraintLine.text == "" {
		return nil
	}

	file := pass.Fset.AddFile(filename, -1, len(content))
	file.SetLinesForContent(content)
	reportBuildConstraint(pass, sourceConstraint{
		text: constraintLine.text,
		pos:  file.LineStart(constraintLine.line),
	})
	return nil
}

func rawFileConstraint(content []byte) rawConstraint {
	var goBuild, plusBuild []rawConstraint
	for index, contentLine := range bytes.Split(content, []byte{'\n'}) {
		line := strings.TrimSpace(string(contentLine))
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] == "package" {
			break
		}
		item := rawConstraint{text: line, line: index + 1}
		switch {
		case constraint.IsGoBuild(line):
			goBuild = append(goBuild, item)
		case constraint.IsPlusBuild(line):
			plusBuild = append(plusBuild, item)
		}
	}

	selected := goBuild
	if len(selected) == 0 {
		selected = plusBuild
	}
	if len(selected) == 0 {
		return rawConstraint{}
	}

	tags := tagsFromConstraintLines(rawConstraintTexts(selected))
	if len(tags) == 0 {
		return rawConstraint{}
	}
	return rawConstraint{
		text: buildTagMessage(tags),
		line: selected[0].line,
	}
}

func mergeConstraints(goBuild, plusBuild []sourceConstraint) sourceConstraint {
	selected := goBuild
	if len(selected) == 0 {
		selected = plusBuild
	}
	if len(selected) == 0 {
		return sourceConstraint{}
	}

	lines := make([]string, 0, len(selected))
	for _, item := range selected {
		lines = append(lines, item.text)
	}
	tags := tagsFromConstraintLines(lines)
	if len(tags) == 0 {
		return sourceConstraint{}
	}

	return sourceConstraint{
		text: buildTagMessage(tags),
		pos:  selected[0].pos,
		end:  selected[0].end,
	}
}

func reportBuildConstraint(pass *analysis.Pass, finding sourceConstraint) {
	if finding.text == "" {
		return
	}
	pass.Report(analysis.Diagnostic{
		Pos:      finding.pos,
		End:      finding.end,
		Category: categoryBuildTag,
		Message:  finding.text,
	})
}

func rawConstraintTexts(items []rawConstraint) []string {
	lines := make([]string, 0, len(items))
	for _, item := range items {
		lines = append(lines, item.text)
	}
	return lines
}

func tagsFromConstraintLines(lines []string) []string {
	tags := make(map[string]struct{})
	for _, line := range lines {
		expression, err := constraint.Parse(line)
		if err != nil {
			continue
		}
		collectCustomTags(expression, tags)
	}

	result := make([]string, 0, len(tags))
	for tag := range tags {
		result = append(result, tag)
	}
	sort.Strings(result)
	return result
}

func collectCustomTags(expression constraint.Expr, tags map[string]struct{}) {
	switch expression := expression.(type) {
	case *constraint.TagExpr:
		if !isStandardBuildTag(expression.Tag) {
			tags[expression.Tag] = struct{}{}
		}
	case *constraint.NotExpr:
		collectCustomTags(expression.X, tags)
	case *constraint.AndExpr:
		collectCustomTags(expression.X, tags)
		collectCustomTags(expression.Y, tags)
	case *constraint.OrExpr:
		collectCustomTags(expression.X, tags)
		collectCustomTags(expression.Y, tags)
	}
}

func buildTagMessage(tags []string) string {
	quoted := make([]string, 0, len(tags))
	for _, tag := range tags {
		quoted = append(quoted, strconv.Quote(tag))
	}

	noun := "tag"
	verb := "gates"
	if len(tags) > 1 {
		noun = "tags"
		verb = "gate"
	}
	return fmt.Sprintf(
		"custom build %s %s %s this file; compare the feature-off build when the target does not need it",
		noun,
		strings.Join(quoted, ", "),
		verb,
	)
}

func isStandardBuildTag(tag string) bool {
	if strings.HasPrefix(tag, "go1.") || strings.HasPrefix(tag, "goexperiment.") {
		return true
	}
	for _, architecture := range architectures {
		if strings.HasPrefix(tag, architecture+".") {
			return true
		}
	}
	_, ok := standardBuildTags[tag]
	return ok
}

var architectures = []string{
	"386",
	"amd64",
	"amd64p32",
	"arm",
	"armbe",
	"arm64",
	"arm64be",
	"loong64",
	"mips",
	"mips64",
	"mips64le",
	"mips64p32",
	"mips64p32le",
	"mipsle",
	"ppc",
	"ppc64",
	"ppc64le",
	"riscv64",
	"s390x",
	"sparc",
	"sparc64",
	"wasm",
}

var standardBuildTags = map[string]struct{}{
	"386":         {},
	"aix":         {},
	"amd64":       {},
	"amd64p32":    {},
	"android":     {},
	"arm":         {},
	"armbe":       {},
	"arm64":       {},
	"arm64be":     {},
	"asan":        {},
	"cgo":         {},
	"darwin":      {},
	"dragonfly":   {},
	"freebsd":     {},
	"fuzz":        {},
	"gc":          {},
	"gccgo":       {},
	"hurd":        {},
	"illumos":     {},
	"ignore":      {},
	"ios":         {},
	"js":          {},
	"linux":       {},
	"loong64":     {},
	"mips":        {},
	"mips64":      {},
	"mips64le":    {},
	"mips64p32":   {},
	"mips64p32le": {},
	"mipsle":      {},
	"msan":        {},
	"nacl":        {},
	"netbsd":      {},
	"openbsd":     {},
	"plan9":       {},
	"ppc":         {},
	"ppc64":       {},
	"ppc64le":     {},
	"race":        {},
	"riscv64":     {},
	"s390x":       {},
	"solaris":     {},
	"sparc":       {},
	"sparc64":     {},
	"unix":        {},
	"wasip1":      {},
	"wasm":        {},
	"windows":     {},
	"zos":         {},
}
