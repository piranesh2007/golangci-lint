package golinters

import (
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/golangci/golangci-lint/pkg/golinters/goanalysis"
	"github.com/romanyx/erris"
)

func NewErris() *goanalysis.Linter {
	return goanalysis.NewLinter(
		erris.Analyzer.Name,
		strings.SplitN(erris.Analyzer.Doc, "\n\n", 2)[0],
		[]*analysis.Analyzer{erris.Analyzer},
		nil,
	).WithLoadMode(goanalysis.LoadModeTypesInfo)
}
