package main

import (
	goc "github.com/go-critic/go-critic/checkers/analyzer"
	"go/ast"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
	"strings"
)

var ErrCheckAnalyzer = &analysis.Analyzer{
	Name: "errcheck",
	Doc:  "check for unchecked errors",
	Run:  run,
}

func main() {

	mychecks := []*analysis.Analyzer{
		shadow.Analyzer,          // Find shadowed variables
		structtag.Analyzer,       // Check struct tags
		assign.Analyzer,          // Check assignments
		atomic.Analyzer,          // Check for common mistakes using the sync/atomic package
		bools.Analyzer,           // Check for common mistakes involving boolean operators
		composite.Analyzer,       // Check for common mistakes involving composite literals
		copylock.Analyzer,        // Check for locks erroneously passed by value
		deepequalerrors.Analyzer, // Check for errors in deep equal comparisons
		defers.Analyzer,          // Check for common mistakes involving defer statements
		directive.Analyzer,       // Check for common mistakes involving build directives
		errorsas.Analyzer,        // Check for common mistakes involving errors.As
		nilfunc.Analyzer,         // Check for common mistakes involving nil function calls
		tests.Analyzer,           // Check for common mistakes in tests
		timeformat.Analyzer,      // Check for common mistakes involving time.Format
		unmarshal.Analyzer,       // Check for common mistakes involving encoding/json.Unmarshal
		unreachable.Analyzer,     // Check for unreachable code
		unusedresult.Analyzer,    // Check for unused results of calls to functions and methods
		goc.Analyzer,             // Check for common mistakes in Go code
	}

	// Include all SA checks
	for _, v := range staticcheck.Analyzers {
		if strings.HasPrefix(v.Analyzer.Name, "SA") {
			mychecks = append(mychecks, v.Analyzer)
		}
	}

	// Use plain channel send or receive instead of single-case select
	for _, v := range simple.Analyzers {
		if v.Analyzer.Name == "S1000" {
			mychecks = append(mychecks, v.Analyzer)
		}
	}

	// Incorrectly formatted error string
	for _, v := range stylecheck.Analyzers {
		if v.Analyzer.Name == "ST1005" {
			mychecks = append(mychecks, v.Analyzer)
		}
	}
	// Check os.Exit
	mychecks = append(mychecks, ErrCheckAnalyzer)
	multichecker.Main(
		mychecks...,
	)
}

func run(pass *analysis.Pass) (interface{}, error) {
	var err error
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			if x, ok := n.(*ast.CallExpr); ok {
				if exitOSChecker(pass, x) {
					pass.Reportf(x.Pos(), "os.Exit must not be used in main function")
				}
			}
			return true
		})
	}
	return nil, err
}

func exitOSChecker(pass *analysis.Pass, x *ast.CallExpr) bool {
	if selExpr, ok := x.Fun.(*ast.SelectorExpr); ok {
		if idExpr, ok := selExpr.X.(*ast.Ident); ok && idExpr.Name == "os" && selExpr.Sel.Name == "Exit" && pass.Pkg.Name() == "main" {
			for _, f := range pass.Files {
				if f.Name.Name == "main" {
					return true
				}
			}
		}
	}

	return false
}
