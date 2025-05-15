// Package noexit определяет анализатор, проверяющий прямые вызовы os.Exit в функциях main.
//
// Этот анализатор выявляет случаи, когда os.Exit вызывается напрямую внутри функции main
// пакета main, что приводит к немедленному завершению программы без возможности
// корректной очистки ресурсов, освобождения памяти или плавного завершения работы.
package noexit

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "noexit",
	Doc:      "проверка прямых вызовов os.Exit в функции main пакета main",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspector := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
		(*ast.FuncDecl)(nil),
	}

	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	var inMainFunc bool

	inspector.Preorder(nodeFilter, func(node ast.Node) {
		if fd, ok := node.(*ast.FuncDecl); ok {
			inMainFunc = fd.Name.Name == "main"
			return
		}

		if !inMainFunc {
			return
		}

		call, ok := node.(*ast.CallExpr)
		if !ok {
			return
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return
		}

		ident, ok := sel.X.(*ast.Ident)
		if !ok {
			return
		}

		if ident.Name == "os" && sel.Sel.Name == "Exit" {
			if obj := pass.TypesInfo.Uses[ident]; obj != nil {
				if pkg, ok := obj.(*types.PkgName); ok && pkg.Imported().Path() == "os" {
					pass.Reportf(call.Pos(), "прямой вызов os.Exit в функции main запрещен")
				}
			}
		}
	})

	return nil, nil
}
