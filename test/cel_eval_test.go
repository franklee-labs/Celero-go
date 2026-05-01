package test

import (
	"fmt"
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/interpreter"
)

func buildProg(t *testing.T) cel.Program {
	t.Helper()
	env, err := cel.NewEnv(
		cel.Variable("params", cel.MapType(cel.StringType, cel.DynType)),
		cel.Variable("_SYS_KEY", cel.IntType),
	)
	if err != nil {
		t.Fatal(err)
	}
	ast, iss := env.Compile(`params.age == _SYS_KEY`)
	if iss.Err() != nil {
		t.Fatal(iss.Err())
	}
	prg, err := env.Program(ast)
	if err != nil {
		t.Fatal(err)
	}
	return prg
}

// Case1: params present but the "age" key is missing from the map.
func TestMissingMapKey(t *testing.T) {
	prg := buildProg(t)
	out, _, err := prg.Eval(map[string]any{
		"params":   map[string]any{"name": "frank"},
		"_SYS_KEY": int64(18),
	})
	fmt.Printf("Case1 (map key missing) out=%v  type(out)=%v  err=%q\n", out, out.Type(), err)
	// Case1 (map key missing) out=no such key: age  type(out)=error  err="no such key: age"
}

// Case1a: params present but the "age" value is nil.
func TestMapNilVal(t *testing.T) {
	prg := buildProg(t)
	out, _, err := prg.Eval(map[string]any{
		"params":   map[string]any{"age": nil},
		"_SYS_KEY": int64(18),
	})
	fmt.Printf("Case1a (map nil val) out=%v  type(out)=%v  err=%q\n", out, out.Type(), err)
	// Case1a (map nil val) out=false  type(out)=bool  err=%!q(<nil>)
}

// Case1b: params present but the value is nil.
func TestMapNilVal2(t *testing.T) {
	prg := buildProg(t)
	out, _, err := prg.Eval(map[string]any{
		"params":   nil,
		"_SYS_KEY": int64(18),
	})
	fmt.Printf("Case1b (map nil val) out=%v  type(out)=%v  err=%q\n", out, out.Type(), err)
	// Case1b (map nil val) out=no such key: age  type(out)=error  err="no such key: age"
}

// Case2: top-level variable "params" is missing from the activation entirely.
func TestMissingTopLevelVar(t *testing.T) {
	prg := buildProg(t)
	out, _, err := prg.Eval(map[string]any{
		"_SYS_KEY": int64(18),
	})
	fmt.Printf("Case2 (var missing)     out=%v    type(out)=%v  err=%q\n", out, out.Type(), err)
	// Case2 (var missing)     out=no such attribute(s): params    type(out)=error  err="no such attribute(s): params"
}

// Case3a: params present but key missing + wildcard partial activation.
func TestPartialActivationWildcardWithParams(t *testing.T) {
	prg := buildProg(t)
	pa, err := interpreter.NewPartialActivation(
		map[string]any{
			"params":   map[string]any{"name": "frank"},
			"_SYS_KEY": int64(18),
		},
		interpreter.NewAttributePattern("params").Wildcard(),
	)
	if err != nil {
		t.Fatal(err)
	}
	out, _, evalErr := prg.Eval(pa)
	fmt.Printf("Case3a (wildcard+params) out=%v  isUnknown=%v  err=%v\n",
		out, types.IsUnknown(out), evalErr)
}

// Case3: PartialActivation with Wildcard – marks all keys under params as unknown.
func TestPartialActivationWildcard(t *testing.T) {
	prg := buildProg(t)
	pa, err := interpreter.NewPartialActivation(
		map[string]any{"_SYS_KEY": int64(18)},
		interpreter.NewAttributePattern("params").Wildcard(),
	)
	if err != nil {
		t.Fatal(err)
	}
	out, _, evalErr := prg.Eval(pa)
	fmt.Printf("Case3 (wildcard partial) out=%v  isUnknown=%v  err=%v\n",
		out, types.IsUnknown(out), evalErr)
}

// Case4: PartialActivation marking the specific key "age" as unknown.
func TestPartialActivationSpecificKey(t *testing.T) {
	prg := buildProg(t)
	pa, err := interpreter.NewPartialActivation(
		map[string]any{
			"params":   map[string]any{"name": "frank"},
			"_SYS_KEY": int64(18),
		},
		interpreter.NewAttributePattern("params").QualString("age"),
	)
	if err != nil {
		t.Fatal(err)
	}
	out, _, evalErr := prg.Eval(pa)
	fmt.Printf("Case4 (specific key)    out=%v  isUnknown=%v  err=%v\n",
		out, types.IsUnknown(out), evalErr)
}
