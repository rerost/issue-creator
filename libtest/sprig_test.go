package libtest_test

import (
	"fmt"
	"reflect"
	"slices"
	"testing"

	"github.com/Masterminds/sprig/v3"
	"github.com/bradleyjkemp/cupaloy/v2"
)

func TestFuncMapSnapshot(t *testing.T) {
	t.Parallel()
	l := ToFunctionList(t, sprig.FuncMap())

	listStr := ""
	for _, sig := range l {
		listStr = listStr + "\n" + sig
	}

	cupaloy.SnapshotT(t, listStr)
}

func ToFunctionList(t *testing.T, funcMap map[string]any) []string {
	t.Helper()

	funcs := []string{}
	for funcName, f := range funcMap {
		funcs = append(
			funcs,
			fmt.Sprintf("%s: %s", funcName, FuncToSig(f)),
		)
	}

	slices.Sort(funcs)
	return funcs
}

// FuncToString(func Hoge(i int) string)
// "(int) -> string"
func FuncToSig(f any) string {
	funcType := reflect.TypeOf(f)

	sig := "("

	for i := 0; i < funcType.NumIn(); i++ {
		if i > 0 {
			sig = sig + ", "
		}
		sig = sig + fmt.Sprintf("%v", funcType.In(i))
	}
	sig = sig + ") -> "

	if funcType.NumOut() == 0 {
		return ""
	}
	for i := 0; i < funcType.NumOut(); i++ {
		if i > 0 {
			sig = sig + ", "
		}
		sig = sig + fmt.Sprintf("%v", funcType.Out(i))
	}

	return sig
}
