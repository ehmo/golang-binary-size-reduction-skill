// Command gosizeaudit reports source patterns that may affect Go binary size.
package main

import (
	"github.com/ehmo/golang-binary-size-reduction-skill/analyzer"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(analyzer.Analyzer)
}
