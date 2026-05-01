package condition

import (
	"fmt"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/ast"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

const builtin_var string = "_SYS_BUILTIN_VAR"

const builtin_inner_key_prefix string = "_SYS_BUILTIN_KEY_"

func mapOfCelType(k string, v *cel.Type) map[string]*cel.Type {
	m := make(map[string]*cel.Type)
	m[k] = v
	return m
}

func mapOf(k string, v interface{}) map[string]interface{} {
	m := make(map[string]interface{})
	m[k] = v
	return m
}

// ExtractTopLevelVars parses a CEL expression and returns the unique list of
// top-level variable names it references. For a chained selector like a.b.c
// only "a" is returned. Comprehension-bound variables (iter/accu) are excluded.
func ExtractTopLevelVars(expression string) ([]string, error) {
	env, err := cel.NewEnv()
	if err != nil {
		return nil, err
	}
	parsed, issues := env.Parse(expression)
	if issues != nil && issues.Err() != nil {
		return nil, issues.Err()
	}
	seen := make(map[string]struct{})
	collectRoots(parsed.NativeRep().Expr(), seen, nil)
	vars := make([]string, 0, len(seen))
	for v := range seen {
		vars = append(vars, v)
	}
	return vars, nil
}

// collectRoots walks an AST expression and records the root identifier of every
// variable reference into seen. bound holds variable names introduced by
// enclosing comprehensions that should not be treated as top-level.
func collectRoots(expr ast.Expr, seen map[string]struct{}, bound map[string]struct{}) {
	switch expr.Kind() {
	case ast.IdentKind:
		name := expr.AsIdent()
		if _, ok := bound[name]; !ok {
			seen[name] = struct{}{}
		}

	case ast.SelectKind:
		// Only descend into the operand to find the root identifier.
		collectRoots(expr.AsSelect().Operand(), seen, bound)

	case ast.CallKind:
		call := expr.AsCall()
		if call.IsMemberFunction() {
			collectRoots(call.Target(), seen, bound)
		}
		for _, arg := range call.Args() {
			collectRoots(arg, seen, bound)
		}

	case ast.ListKind:
		for _, elem := range expr.AsList().Elements() {
			collectRoots(elem, seen, bound)
		}

	case ast.MapKind:
		for _, entry := range expr.AsMap().Entries() {
			me := entry.AsMapEntry()
			collectRoots(me.Key(), seen, bound)
			collectRoots(me.Value(), seen, bound)
		}

	case ast.ComprehensionKind:
		comp := expr.AsComprehension()
		// IterRange and AccuInit are evaluated in the outer scope.
		collectRoots(comp.IterRange(), seen, bound)
		collectRoots(comp.AccuInit(), seen, bound)
		// Inside the loop the iter/accu variables are bound — build an inner set.
		inner := make(map[string]struct{}, len(bound)+3)
		for k := range bound {
			inner[k] = struct{}{}
		}
		inner[comp.IterVar()] = struct{}{}
		inner[comp.AccuVar()] = struct{}{}
		if comp.HasIterVar2() {
			inner[comp.IterVar2()] = struct{}{}
		}
		collectRoots(comp.LoopCondition(), seen, inner)
		collectRoots(comp.LoopStep(), seen, inner)
		collectRoots(comp.Result(), seen, inner)

		// LiteralKind, StructKind, UnspecifiedExprKind: no variable references.
	}
}

func BuildCELProgram(expression string, builtinVars map[string]*cel.Type) (cel.Program, error) {
	vars, err := ExtractTopLevelVars(expression)
	if err != nil {
		return nil, err
	}
	prg, err := BuildCELProgramWithVars(expression, vars, builtinVars)
	if err != nil {
		return nil, err
	}
	return prg, nil
}

func BuildCELProgramWithVars(expression string, vars []string, celTypes map[string]*cel.Type) (cel.Program, error) {
	vs := make([]cel.EnvOption, 0, len(vars))
	for _, v := range vars {
		if celTypes == nil {
			vs = append(vs, cel.Variable(v, cel.DynType))
		} else {
			if t, ok := celTypes[v]; ok {
				vs = append(vs, cel.Variable(v, t))
			} else {
				vs = append(vs, cel.Variable(v, cel.DynType))
			}
		}
	}
	env, err := cel.NewEnv(vs...)
	// Check err for environment setup errors.
	if err != nil {
		return nil, err
	}
	ast, iss := env.Compile(expression)
	// Check iss for compilation errors.
	if iss.Err() != nil {
		return nil, iss.Err()
	}
	prg, err := env.Program(ast)
	if err != nil {
		return nil, err
	}
	return prg, nil
	// out, _, err := prg.Eval(map[string]interface{}{
	// 	"name": "CEL",
	// })
}

func IsKeyAbsent(out ref.Val, err error) bool {
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "no such key") || strings.Contains(msg, "no such attribute") {
			return true
		}
	} else {
		if types.IsError(out) {
			err = out.(*types.Err)
			msg := err.Error()
			if strings.Contains(msg, "no such key") || strings.Contains(msg, "no such attribute") {
				return true
			}
		}
	}
	return false
}

func Out(out ref.Val, err error) (bool, bool, error) {
	if IsKeyAbsent(out, err) {
		return false, true, &EvalError{Cause: err}
	} else if err != nil {
		return false, false, &EvalError{Cause: err}
	} else {
		b, ok := out.Value().(bool)
		if ok {
			return b, false, nil
		}
		return false, false, &EvalError{Cause: fmt.Errorf("result type is not bool. got:%s", out.Type().TypeName())}
	}
}
