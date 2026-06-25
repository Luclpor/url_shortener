package main

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// Analyzer reports calls that should not be used in project code.
var Analyzer = &analysis.Analyzer{
	Name: "panicexit",
	Doc:  "reports built-in panic calls and forbidden process termination calls",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		inspectFile(pass, file)
	}
	return nil, nil
}

func inspectFile(pass *analysis.Pass, file *ast.File) {
	var nodeStack []ast.Node
	var funcStack []*ast.FuncDecl

	ast.Inspect(file, func(node ast.Node) bool {
		if node == nil {
			last := nodeStack[len(nodeStack)-1]
			nodeStack = nodeStack[:len(nodeStack)-1]
			if _, ok := last.(*ast.FuncDecl); ok {
				funcStack = funcStack[:len(funcStack)-1]
			}
			return true
		}

		nodeStack = append(nodeStack, node)
		if fn, ok := node.(*ast.FuncDecl); ok {
			funcStack = append(funcStack, fn)
		}

		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}

		var enclosingFunc *ast.FuncDecl
		if len(funcStack) > 0 {
			enclosingFunc = funcStack[len(funcStack)-1]
		}
		checkCall(pass, call, enclosingFunc)

		return true
	})
}

func checkCall(pass *analysis.Pass, call *ast.CallExpr, enclosingFunc *ast.FuncDecl) {
	if isBuiltinPanic(pass, call) {
		pass.Reportf(call.Pos(), "use of built-in panic is prohibited")
		return
	}

	name, ok := processTerminator(pass, call)
	if !ok || isMainFunc(pass, enclosingFunc) {
		return
	}

	pass.Reportf(call.Pos(), "%s is allowed only in function main of package main", name)
}

func isBuiltinPanic(pass *analysis.Pass, call *ast.CallExpr) bool {
	ident, ok := call.Fun.(*ast.Ident)
	if !ok || ident.Name != "panic" {
		return false
	}

	obj, ok := pass.TypesInfo.Uses[ident].(*types.Builtin)
	return ok && obj.Name() == "panic"
}

func processTerminator(pass *analysis.Pass, call *ast.CallExpr) (string, bool) {
	var obj types.Object
	switch fun := call.Fun.(type) {
	case *ast.SelectorExpr:
		obj = pass.TypesInfo.Uses[fun.Sel]
	case *ast.Ident:
		obj = pass.TypesInfo.Uses[fun]
	default:
		return "", false
	}

	fn, ok := obj.(*types.Func)
	if !ok || fn.Pkg() == nil {
		return "", false
	}

	switch {
	case fn.Pkg().Path() == "log" && fn.Name() == "Fatal":
		return "log.Fatal", true
	case fn.Pkg().Path() == "os" && fn.Name() == "Exit":
		return "os.Exit", true
	default:
		return "", false
	}
}

func isMainFunc(pass *analysis.Pass, fn *ast.FuncDecl) bool {
	return pass.Pkg.Name() == "main" && fn != nil && fn.Recv == nil && fn.Name.Name == "main"
}
