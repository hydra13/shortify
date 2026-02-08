package panicsanalyzer

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var PanicsAnalyzer = &analysis.Analyzer{
	Name: "panicscheck",
	Doc:  "check that code without panics",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}

			isMain := fn.Name.Name == "main"
			ast.Inspect(fn, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}

				checkPanicCall(pass, call)

				if !isMain {
					checkLogFatalCall(pass, call)
					checkOsExitCall(pass, call)
				}

				return true
			})
			return true
		})
	}
	return nil, nil
}

func checkPanicCall(pass *analysis.Pass, call *ast.CallExpr) {
	if fn, ok := call.Fun.(*ast.Ident); ok {
		if fn.Name == "panic" {
			pass.Reportf(fn.Pos(), "call of panic()")
		}
	}
}

func checkLogFatalCall(pass *analysis.Pass, call *ast.CallExpr) {
	if se, ok := call.Fun.(*ast.SelectorExpr); ok {
		if ident, ok := se.X.(*ast.Ident); ok && ident.Name == "log" {
			fnName := se.Sel.Name
			if fnName == "Fatalf" || fnName == "Fatal" || fnName == "Fatalln" {
				pass.Reportf(se.Pos(), "call of log.%s()", fnName)
			}
		}
	}
}

func checkOsExitCall(pass *analysis.Pass, call *ast.CallExpr) {
	if se, ok := call.Fun.(*ast.SelectorExpr); ok {
		if ident, ok := se.X.(*ast.Ident); ok && ident.Name == "os" && se.Sel.Name == "Exit" {
			pass.Reportf(se.Pos(), "call of os.Exit()")
		}
	}
}
