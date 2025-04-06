package mofu

import (
	"reflect"
	"runtime"
	"strings"
)

func funcName(fn reflect.Value) string {
	name := runtime.FuncForPC(fn.Pointer()).Name()

	// name = "(packagePath).(typeName).(funcName)-fm"
	i := strings.LastIndexByte(name, '.')
	if i >= 0 {
		name = name[i+1:]
	}
	s, _ := strings.CutSuffix(name, "-fm")
	return s
}
