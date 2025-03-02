package libtest_test

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"

	sprigv3 "github.com/Masterminds/sprig/v3"
	cupaloyv2 "github.com/bradleyjkemp/cupaloy/v2"
)

func TestFuncMapSnapshot(t *testing.T) {
	t.Parallel()
	l := ToFunctionList(t, sprigv3.FuncMap())

	listStr := ""
	for _, sig := range l {
		listStr = listStr + "\n" + sig
	}

	cupaloyv2.SnapshotT(t, listStr)
}

//	map[string]any{
//	  "Hoge": func() {},
//	  "AddDateAndFormat": func(format string, d int) string {}
//	}
//
// -> []string{"AddDateAndFormat: (string, int) -> string", "Hoge: () -> ()"}
func ToFunctionList(t *testing.T, funcMap map[string]any) []string {
	t.Helper()
	funcs := []string{}

	for funcName, f := range funcMap {
		sig := FuncToSig(t, f)
		funcs = append(funcs, fmt.Sprintf("%s: %s", funcName, sig))
	}

	// 結果が安定しないためソート
	sort.StringSlice(funcs).Sort()
	return funcs
}

// FuncToString(func Hoge(i int) string)
// "(int) -> string"
// FuncToString(func Hoge(i int) *string)
// "(int) -> *string"
func FuncToSig(t *testing.T, f any) string {
	t.Helper()
	if f == nil {
		t.Error("function is nil")
		return ""
	}

	ft := reflect.TypeOf(f)
	if ft.Kind() != reflect.Func {
		t.Error("not a function")
		return ""
	}

	in := []string{}
	for i := 0; i < ft.NumIn(); i++ {
		in = append(in, ft.In(i).String())
	}

	out := []string{}
	for i := 0; i < ft.NumOut(); i++ {
		out = append(out, ft.Out(i).String())
	}

	return fmt.Sprintf("(%s) -> %s", strings.Join(in, ", "), strings.Join(out, ", "))
}
