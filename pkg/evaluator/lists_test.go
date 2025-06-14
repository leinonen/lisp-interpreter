package evaluator

import (
	"testing"

	"github.com/leinonen/lisp-interpreter/pkg/types"
)

func TestEvaluatorListOperations(t *testing.T) {
	tests := []struct {
		name     string
		expr     types.Expr
		expected types.Value
	}{
		{
			name: "list creation with no arguments",
			expr: &types.ListExpr{
				Elements: []types.Expr{
					&types.SymbolExpr{Name: "list"},
				},
			},
			expected: &types.ListValue{Elements: []types.Value{}},
		},
		{
			name: "list creation with single element",
			expr: &types.ListExpr{
				Elements: []types.Expr{
					&types.SymbolExpr{Name: "list"},
					&types.NumberExpr{Value: 42},
				},
			},
			expected: &types.ListValue{Elements: []types.Value{types.NumberValue(42)}},
		},
		{
			name: "list creation with multiple elements",
			expr: &types.ListExpr{
				Elements: []types.Expr{
					&types.SymbolExpr{Name: "list"},
					&types.NumberExpr{Value: 1},
					&types.NumberExpr{Value: 2},
					&types.NumberExpr{Value: 3},
				},
			},
			expected: &types.ListValue{Elements: []types.Value{
				types.NumberValue(1),
				types.NumberValue(2),
				types.NumberValue(3),
			}},
		},
		{
			name: "list with mixed types",
			expr: &types.ListExpr{
				Elements: []types.Expr{
					&types.SymbolExpr{Name: "list"},
					&types.NumberExpr{Value: 42},
					&types.StringExpr{Value: "hello"},
					&types.BooleanExpr{Value: true},
				},
			},
			expected: &types.ListValue{Elements: []types.Value{
				types.NumberValue(42),
				types.StringValue("hello"),
				types.BooleanValue(true),
			}},
		},
		{
			name: "empty? on empty list",
			expr: &types.ListExpr{
				Elements: []types.Expr{
					&types.SymbolExpr{Name: "empty?"},
					&types.ListExpr{
						Elements: []types.Expr{
							&types.SymbolExpr{Name: "list"},
						},
					},
				},
			},
			expected: types.BooleanValue(true),
		},
		{
			name: "empty? on non-empty list",
			expr: &types.ListExpr{
				Elements: []types.Expr{
					&types.SymbolExpr{Name: "empty?"},
					&types.ListExpr{
						Elements: []types.Expr{
							&types.SymbolExpr{Name: "list"},
							&types.NumberExpr{Value: 1},
						},
					},
				},
			},
			expected: types.BooleanValue(false),
		},
		{
			name: "length of empty list",
			expr: &types.ListExpr{
				Elements: []types.Expr{
					&types.SymbolExpr{Name: "length"},
					&types.ListExpr{
						Elements: []types.Expr{
							&types.SymbolExpr{Name: "list"},
						},
					},
				},
			},
			expected: types.NumberValue(0),
		},
		{
			name: "length of non-empty list",
			expr: &types.ListExpr{
				Elements: []types.Expr{
					&types.SymbolExpr{Name: "length"},
					&types.ListExpr{
						Elements: []types.Expr{
							&types.SymbolExpr{Name: "list"},
							&types.NumberExpr{Value: 1},
							&types.NumberExpr{Value: 2},
							&types.NumberExpr{Value: 3},
						},
					},
				},
			},
			expected: types.NumberValue(3),
		},
		{
			name: "first of non-empty list",
			expr: &types.ListExpr{
				Elements: []types.Expr{
					&types.SymbolExpr{Name: "first"},
					&types.ListExpr{
						Elements: []types.Expr{
							&types.SymbolExpr{Name: "list"},
							&types.NumberExpr{Value: 1},
							&types.NumberExpr{Value: 2},
							&types.NumberExpr{Value: 3},
						},
					},
				},
			},
			expected: types.NumberValue(1),
		},
		{
			name: "rest of non-empty list",
			expr: &types.ListExpr{
				Elements: []types.Expr{
					&types.SymbolExpr{Name: "rest"},
					&types.ListExpr{
						Elements: []types.Expr{
							&types.SymbolExpr{Name: "list"},
							&types.NumberExpr{Value: 1},
							&types.NumberExpr{Value: 2},
							&types.NumberExpr{Value: 3},
						},
					},
				},
			},
			expected: &types.ListValue{Elements: []types.Value{
				types.NumberValue(2),
				types.NumberValue(3),
			}},
		},
		{
			name: "rest of single element list",
			expr: &types.ListExpr{
				Elements: []types.Expr{
					&types.SymbolExpr{Name: "rest"},
					&types.ListExpr{
						Elements: []types.Expr{
							&types.SymbolExpr{Name: "list"},
							&types.NumberExpr{Value: 1},
						},
					},
				},
			},
			expected: &types.ListValue{Elements: []types.Value{}},
		},
		{
			name: "cons element to list",
			expr: &types.ListExpr{
				Elements: []types.Expr{
					&types.SymbolExpr{Name: "cons"},
					&types.NumberExpr{Value: 0},
					&types.ListExpr{
						Elements: []types.Expr{
							&types.SymbolExpr{Name: "list"},
							&types.NumberExpr{Value: 1},
							&types.NumberExpr{Value: 2},
						},
					},
				},
			},
			expected: &types.ListValue{Elements: []types.Value{
				types.NumberValue(0),
				types.NumberValue(1),
				types.NumberValue(2),
			}},
		},
		{
			name: "cons element to empty list",
			expr: &types.ListExpr{
				Elements: []types.Expr{
					&types.SymbolExpr{Name: "cons"},
					&types.NumberExpr{Value: 42},
					&types.ListExpr{
						Elements: []types.Expr{
							&types.SymbolExpr{Name: "list"},
						},
					},
				},
			},
			expected: &types.ListValue{Elements: []types.Value{
				types.NumberValue(42),
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := NewEnvironment()
			evaluator := NewEvaluator(env)
			result, err := evaluator.Eval(tt.expr)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !valuesEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluatorListOperationErrors(t *testing.T) {
	tests := []struct {
		name string
		expr types.Expr
	}{
		{
			name: "first on empty list",
			expr: &types.ListExpr{
				Elements: []types.Expr{
					&types.SymbolExpr{Name: "first"},
					&types.ListExpr{
						Elements: []types.Expr{
							&types.SymbolExpr{Name: "list"},
						},
					},
				},
			},
		},
		{
			name: "rest on empty list",
			expr: &types.ListExpr{
				Elements: []types.Expr{
					&types.SymbolExpr{Name: "rest"},
					&types.ListExpr{
						Elements: []types.Expr{
							&types.SymbolExpr{Name: "list"},
						},
					},
				},
			},
		},
		{
			name: "first with wrong number of arguments",
			expr: &types.ListExpr{
				Elements: []types.Expr{
					&types.SymbolExpr{Name: "first"},
				},
			},
		},
		{
			name: "rest with wrong number of arguments",
			expr: &types.ListExpr{
				Elements: []types.Expr{
					&types.SymbolExpr{Name: "rest"},
					&types.ListExpr{
						Elements: []types.Expr{
							&types.SymbolExpr{Name: "list"},
							&types.NumberExpr{Value: 1},
						},
					},
					&types.NumberExpr{Value: 2}, // extra argument
				},
			},
		},
		{
			name: "cons with wrong number of arguments",
			expr: &types.ListExpr{
				Elements: []types.Expr{
					&types.SymbolExpr{Name: "cons"},
					&types.NumberExpr{Value: 1},
				},
			},
		},
		{
			name: "length with wrong number of arguments",
			expr: &types.ListExpr{
				Elements: []types.Expr{
					&types.SymbolExpr{Name: "length"},
				},
			},
		},
		{
			name: "empty? with wrong number of arguments",
			expr: &types.ListExpr{
				Elements: []types.Expr{
					&types.SymbolExpr{Name: "empty?"},
				},
			},
		},
		{
			name: "first on non-list",
			expr: &types.ListExpr{
				Elements: []types.Expr{
					&types.SymbolExpr{Name: "first"},
					&types.NumberExpr{Value: 42},
				},
			},
		},
		{
			name: "rest on non-list",
			expr: &types.ListExpr{
				Elements: []types.Expr{
					&types.SymbolExpr{Name: "rest"},
					&types.NumberExpr{Value: 42},
				},
			},
		},
		{
			name: "cons with non-list second argument",
			expr: &types.ListExpr{
				Elements: []types.Expr{
					&types.SymbolExpr{Name: "cons"},
					&types.NumberExpr{Value: 1},
					&types.NumberExpr{Value: 42},
				},
			},
		},
		{
			name: "length on non-list",
			expr: &types.ListExpr{
				Elements: []types.Expr{
					&types.SymbolExpr{Name: "length"},
					&types.NumberExpr{Value: 42},
				},
			},
		},
		{
			name: "empty? on non-list",
			expr: &types.ListExpr{
				Elements: []types.Expr{
					&types.SymbolExpr{Name: "empty?"},
					&types.NumberExpr{Value: 42},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := NewEnvironment()
			evaluator := NewEvaluator(env)
			_, err := evaluator.Eval(tt.expr)

			if err == nil {
				t.Errorf("expected error for %s", tt.name)
			}
		})
	}
}
