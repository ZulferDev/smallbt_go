package compiler

import (
	"fmt"

	"github.com/1jehuang/backtest/internal/strategy/ast"
)

// Compile compiles a strategy AST into an executable form.
func Compile(ast *ast.Strategy) (interface{}, error) {
	return nil, fmt.Errorf("compiler not implemented yet")
}
