package test

import (
	"fmt"
	"testing"

	"github.com/google/cel-go/cel"
)

func buildExistsProg(t *testing.T) cel.Program {
	t.Helper()
	env, err := cel.NewEnv(
		cel.Variable("list1", cel.ListType(cel.DynType)),
		cel.Variable("list2", cel.ListType(cel.DynType)),
	)
	if err != nil {
		t.Fatal(err)
	}
	ast, iss := env.Compile(`list1.exists(x, list2.exists(y, y == x))`)
	if iss.Err() != nil {
		t.Fatal(iss.Err())
	}
	prg, err := env.Program(ast)
	if err != nil {
		t.Fatal(err)
	}
	return prg
}

func TestExistsCase1(t *testing.T) {
	prg := buildExistsProg(t)
	out, _, err := prg.Eval(map[string]any{
		"list1": []string{"aa", "bb", "cc"},
		"list2": []string{"cc", "dd", "ee"},
	})
	fmt.Printf("Case1 out=%v  type(out)=%v  err=%q\n", out, out.Type(), err)
	// Case1 out=true  type(out)=bool  err=%!q(<nil>)
}

func TestExistsCase2(t *testing.T) {
	prg := buildExistsProg(t)
	out, _, err := prg.Eval(map[string]any{
		"list1": []string{"aa", "bb", "cc"},
		"list2": []string{"ff", "dd", "ee"},
	})
	fmt.Printf("Case2 out=%v  type(out)=%v  err=%q\n", out, out.Type(), err)
	// Case1 out=false  type(out)=bool  err=%!q(<nil>)
}

func TestExistsCase3(t *testing.T) {
	prg := buildExistsProg(t)
	out, _, err := prg.Eval(map[string]any{
		"list1": []interface{}{"aa", true, 1},
		"list2": []interface{}{true, "dd", 2.0},
	})
	fmt.Printf("Case3 out=%v  type(out)=%v  err=%q\n", out, out.Type(), err)
	// Case3 out=true  type(out)=bool  err=%!q(<nil>)
}

func TestExistsCase4(t *testing.T) {
	prg := buildExistsProg(t)
	out, _, err := prg.Eval(map[string]any{
		"list1": []interface{}{"aa", true, 2.01},
		"list2": []interface{}{false, "dd", 2.01},
	})
	fmt.Printf("Case4 out=%v  type(out)=%v  err=%q\n", out, out.Type(), err)
	// Case4 out=true  type(out)=bool  err=%!q(<nil>)
}

func TestExistsCase5(t *testing.T) {
	prg := buildExistsProg(t)
	out, _, err := prg.Eval(map[string]any{
		"list1": []interface{}{"aa", true, 2.01},
		"list2": []interface{}{false, "dd", 2.00},
	})
	fmt.Printf("Case5 out=%v  type(out)=%v  err=%q\n", out, out.Type(), err)
	// Case5 out=false  type(out)=bool  err=%!q(<nil>)
}
