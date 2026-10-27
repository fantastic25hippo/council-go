package meeting

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSahilLiteralsGuard(t *testing.T) {
	// Add a test that fails if internal/meeting non-test files compare Sahil's input to any literal other than quit/exit/c.
	// Scan for == " and case " on the lowered input variables in input.go, hello.go, split.go.
	// Failure message must name "send it to the members" as the fix.
	// Empty-line pause resume is trimmed t == "" (not a phrase); do not fail on == "".
	// Prompt strings (e.g. SAHIL_MEANS in member text) are not Sahil-input compares.

	targetFiles := []string{"input.go", "hello.go", "split.go"}
	allowedLiterals := map[string]bool{
		"":     true,
		"c":    true,
		"quit": true,
		"exit": true,
	}

	for _, fname := range targetFiles {
		path := filepath.Join(".", fname)
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("file not found: %s", path)
		}

		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			t.Fatalf("failed to parse %s: %v", path, err)
		}

		var currentSwitchTagExpr ast.Expr
		ast.Inspect(node, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.SwitchStmt:
				oldTag := currentSwitchTagExpr
				currentSwitchTagExpr = x.Tag
				if x.Body != nil {
					ast.Inspect(x.Body, func(bn ast.Node) bool {
						if cc, ok := bn.(*ast.CaseClause); ok {
							for _, expr := range cc.List {
								if lit, ok := expr.(*ast.BasicLit); ok && lit.Kind == token.STRING {
									val := strings.Trim(lit.Value, "\"")
									varName := exprIdent(currentSwitchTagExpr)
									if isSahilLoweredVar(varName) && !allowedLiterals[val] {
										t.Errorf("%s:%d: case clause with disallowed literal %q on %s; send it to the members instead",
											fname, fset.Position(expr.Pos()).Line, val, varName)
									}
								}
							}
						}
						return true
					})
				}
				currentSwitchTagExpr = oldTag
				return false
			case *ast.BinaryExpr:
				if x.Op == token.EQL || x.Op == token.NEQ {
					checkExpr := func(vExpr, litExpr ast.Expr) {
						lit, ok := litExpr.(*ast.BasicLit)
						if !ok || lit.Kind != token.STRING {
							return
						}
						val := strings.Trim(lit.Value, "\"")
						varName := exprIdent(vExpr)
						if isSahilLoweredVar(varName) {
							if !allowedLiterals[val] {
								t.Errorf("%s:%d: comparison with disallowed literal %q on %s; send it to the members instead",
									fname, fset.Position(x.Pos()).Line, val, varName)
							}
						}
					}
					checkExpr(x.X, x.Y)
					checkExpr(x.Y, x.X)
				}
			}
			return true
		})
	}
}

func exprIdent(e ast.Expr) string {
	if id, ok := e.(*ast.Ident); ok {
		return id.Name
	}
	return ""
}

func isSahilLoweredVar(name string) bool {
	switch name {
	case "low", "q", "ans", "line", "t":
		return true
	default:
		return false
	}
}
