package test

import (
	"fmt"
	"testing"

	"github.com/google/cel-go/cel"
)

func buildHasProg(t *testing.T) cel.Program {
	t.Helper()
	env, err := cel.NewEnv(
		cel.Variable("params", cel.MapType(cel.StringType, cel.DynType)),
	)
	if err != nil {
		t.Fatal(err)
	}
	ast, iss := env.Compile(`has(params.age)`)
	if iss.Err() != nil {
		t.Fatal(iss.Err())
	}
	prg, err := env.Program(ast)
	if err != nil {
		t.Fatal(err)
	}
	return prg
}

func TestHasCase1(t *testing.T) {
	prg := buildHasProg(t)
	out, _, err := prg.Eval(map[string]any{
		"params": map[string]any{"age": 20},
	})
	fmt.Printf("Case1 out=%v  type(out)=%v  err=%q\n", out, out.Type(), err)
	// Case1 out=true  type(out)=bool  err=%!q(<nil>)
}

func TestHasCase2(t *testing.T) {
	prg := buildHasProg(t)
	out, _, err := prg.Eval(map[string]any{
		"params": map[string]any{"name": "frank"},
	})
	fmt.Printf("Case2 out=%v  type(out)=%v  err=%q\n", out, out.Type(), err)
	// Case2 out=false  type(out)=bool  err=%!q(<nil>)
}

func TestHasCase3(t *testing.T) {
	prg := buildHasProg(t)
	out, _, err := prg.Eval(map[string]any{})
	fmt.Printf("Case3 out=%v  type(out)=%v  err=%q\n", out, out.Type(), err)
	// Case3 out=no such attribute(s): params  type(out)=error  err="no such attribute(s): params"
}
