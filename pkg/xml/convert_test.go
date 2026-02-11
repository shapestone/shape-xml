package xml

import (
	"reflect"
	"testing"

	"github.com/shapestone/shape-core/pkg/ast"
)

func TestInterfaceToNode(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
		check   func(*testing.T, ast.SchemaNode)
	}{
		{
			name:  "nil value",
			input: nil,
			check: func(t *testing.T, node ast.SchemaNode) {
				if lit, ok := node.(*ast.LiteralNode); !ok {
					t.Errorf("expected *ast.LiteralNode, got %T", node)
				} else if lit.Value() != nil {
					t.Errorf("expected nil value, got %v", lit.Value())
				}
			},
		},
		{
			name:  "string value",
			input: "hello",
			check: func(t *testing.T, node ast.SchemaNode) {
				lit, ok := node.(*ast.LiteralNode)
				if !ok {
					t.Fatalf("expected *ast.LiteralNode, got %T", node)
				}
				if lit.Value() != "hello" {
					t.Errorf("expected 'hello', got %v", lit.Value())
				}
			},
		},
		{
			name:  "int value",
			input: 42,
			check: func(t *testing.T, node ast.SchemaNode) {
				lit, ok := node.(*ast.LiteralNode)
				if !ok {
					t.Fatalf("expected *ast.LiteralNode, got %T", node)
				}
				if lit.Value() != int64(42) {
					t.Errorf("expected 42, got %v", lit.Value())
				}
			},
		},
		{
			name:  "int64 value",
			input: int64(123),
			check: func(t *testing.T, node ast.SchemaNode) {
				lit, ok := node.(*ast.LiteralNode)
				if !ok {
					t.Fatalf("expected *ast.LiteralNode, got %T", node)
				}
				if lit.Value() != int64(123) {
					t.Errorf("expected 123, got %v", lit.Value())
				}
			},
		},
		{
			name:  "float64 value",
			input: 3.14,
			check: func(t *testing.T, node ast.SchemaNode) {
				lit, ok := node.(*ast.LiteralNode)
				if !ok {
					t.Fatalf("expected *ast.LiteralNode, got %T", node)
				}
				if lit.Value() != 3.14 {
					t.Errorf("expected 3.14, got %v", lit.Value())
				}
			},
		},
		{
			name:  "bool value",
			input: true,
			check: func(t *testing.T, node ast.SchemaNode) {
				lit, ok := node.(*ast.LiteralNode)
				if !ok {
					t.Fatalf("expected *ast.LiteralNode, got %T", node)
				}
				if lit.Value() != true {
					t.Errorf("expected true, got %v", lit.Value())
				}
			},
		},
		{
			name:  "slice value",
			input: []interface{}{"a", "b", "c"},
			check: func(t *testing.T, node ast.SchemaNode) {
				// InterfaceToNode returns ArrayDataNode for slices
				_ = node // Just verify no error
			},
		},
		{
			name:  "map value",
			input: map[string]interface{}{"key": "value"},
			check: func(t *testing.T, node ast.SchemaNode) {
				obj, ok := node.(*ast.ObjectNode)
				if !ok {
					t.Fatalf("expected *ast.ObjectNode, got %T", node)
				}
				_ = obj // Just verify type
			},
		},
		{
			name:    "unsupported struct",
			input:   struct{ Name string }{Name: "test"},
			wantErr: true,
		},
		{
			name:    "unsupported pointer",
			input:   func() *string { s := "hello"; return &s }(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := InterfaceToNode(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("InterfaceToNode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				tt.check(t, node)
			}
		})
	}
}

func TestNodeToInterface(t *testing.T) {
	tests := []struct {
		name  string
		node  ast.SchemaNode
		want  interface{}
	}{
		{
			name: "literal node string",
			node: ast.NewLiteralNode("hello", ast.Position{}),
			want: "hello",
		},
		{
			name: "literal node int",
			node: ast.NewLiteralNode(int64(42), ast.Position{}),
			want: int64(42),
		},
		{
			name: "literal node float",
			node: ast.NewLiteralNode(3.14, ast.Position{}),
			want: 3.14,
		},
		{
			name: "literal node bool",
			node: ast.NewLiteralNode(true, ast.Position{}),
			want: true,
		},
		{
			name: "literal node nil",
			node: ast.NewLiteralNode(nil, ast.Position{}),
			want: nil,
		},
		{
			name: "object node",
			node: func() ast.SchemaNode {
				props := make(map[string]ast.SchemaNode)
				props["name"] = ast.NewLiteralNode("test", ast.Position{})
				return ast.NewObjectNode(props, ast.Position{})
			}(),
			want: map[string]interface{}{"name": "test"},
		},
		{
			name: "array data node",
			node: func() ast.SchemaNode {
				elements := []ast.SchemaNode{
					ast.NewLiteralNode("a", ast.Position{}),
					ast.NewLiteralNode("b", ast.Position{}),
				}
				return ast.NewArrayDataNode(elements, ast.Position{})
			}(),
			want: []interface{}{"a", "b"},
		},
		{
			name: "whole number float to int64",
			node: ast.NewLiteralNode(float64(42.0), ast.Position{}),
			want: int64(42),
		},
		{
			name: "nil node returns nil",
			node: nil,
			want: nil,
		},
		{
			name: "object with sequential numeric keys (legacy array)",
			node: func() ast.SchemaNode {
				props := map[string]ast.SchemaNode{
					"0": ast.NewLiteralNode("x", ast.Position{}),
					"1": ast.NewLiteralNode("y", ast.Position{}),
				}
				return ast.NewObjectNode(props, ast.Position{})
			}(),
			want: []interface{}{"x", "y"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NodeToInterface(tt.node)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeToInterface() = %v (%T), want %v (%T)", got, got, tt.want, tt.want)
			}
		})
	}
}

func TestInterfaceToNode_AdditionalTypes(t *testing.T) {
	checkLiteral := func(t *testing.T, node ast.SchemaNode, want interface{}) {
		t.Helper()
		lit, ok := node.(*ast.LiteralNode)
		if !ok {
			t.Fatalf("expected *ast.LiteralNode, got %T", node)
		}
		if !reflect.DeepEqual(lit.Value(), want) {
			t.Errorf("got %v (%T), want %v (%T)", lit.Value(), lit.Value(), want, want)
		}
	}

	t.Run("int32", func(t *testing.T) {
		node, err := InterfaceToNode(int32(99))
		if err != nil {
			t.Fatal(err)
		}
		checkLiteral(t, node, int64(99))
	})

	t.Run("int16", func(t *testing.T) {
		node, err := InterfaceToNode(int16(16))
		if err != nil {
			t.Fatal(err)
		}
		checkLiteral(t, node, int64(16))
	})

	t.Run("int8", func(t *testing.T) {
		node, err := InterfaceToNode(int8(8))
		if err != nil {
			t.Fatal(err)
		}
		checkLiteral(t, node, int64(8))
	})

	t.Run("uint", func(t *testing.T) {
		node, err := InterfaceToNode(uint(10))
		if err != nil {
			t.Fatal(err)
		}
		checkLiteral(t, node, int64(10))
	})

	t.Run("uint64", func(t *testing.T) {
		node, err := InterfaceToNode(uint64(64))
		if err != nil {
			t.Fatal(err)
		}
		checkLiteral(t, node, int64(64))
	})

	t.Run("uint32", func(t *testing.T) {
		node, err := InterfaceToNode(uint32(32))
		if err != nil {
			t.Fatal(err)
		}
		checkLiteral(t, node, int64(32))
	})

	t.Run("uint16", func(t *testing.T) {
		node, err := InterfaceToNode(uint16(16))
		if err != nil {
			t.Fatal(err)
		}
		checkLiteral(t, node, int64(16))
	})

	t.Run("uint8", func(t *testing.T) {
		node, err := InterfaceToNode(uint8(8))
		if err != nil {
			t.Fatal(err)
		}
		checkLiteral(t, node, int64(8))
	})

	t.Run("float32", func(t *testing.T) {
		node, err := InterfaceToNode(float32(1.5))
		if err != nil {
			t.Fatal(err)
		}
		checkLiteral(t, node, float64(float32(1.5)))
	})
}

func TestReleaseTree(t *testing.T) {
	t.Run("object with children", func(t *testing.T) {
		inner := ast.NewLiteralNode("val", ast.Position{})
		elements := []ast.SchemaNode{inner}
		arr := ast.NewArrayDataNode(elements, ast.Position{})
		props := map[string]ast.SchemaNode{"arr": arr}
		root := ast.NewObjectNode(props, ast.Position{})
		// Should not panic
		ReleaseTree(root)
	})

	t.Run("nil node", func(t *testing.T) {
		// Should not panic
		ReleaseTree(nil)
	})

	t.Run("literal node", func(t *testing.T) {
		node := ast.NewLiteralNode("test", ast.Position{})
		ReleaseTree(node)
	})
}
