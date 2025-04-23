package log

import (
	"runtime"
	"strings"
	"sync"
)

// nolint:gochecknoglobals
var (
	logPackage         string
	minimumCallerDepth int
	callerInitOnce     sync.Once
)

const (
	maximumCallerDepth int = 25
	knownLogFrames     int = 4
)

func getPackageName(f string) string {
	lastPeriod := strings.LastIndex(f, ".")
	lastSlash := strings.LastIndex(f, "/")
	if lastPeriod > lastSlash {
		return f[:lastPeriod]
	}
	return f
}

func getCaller() *runtime.Frame {
	callerInitOnce.Do(func() {
		pcs := make([]uintptr, maximumCallerDepth)
		_ = runtime.Callers(0, pcs)

		for i := range maximumCallerDepth {
			funcName := runtime.FuncForPC(pcs[i]).Name()
			if strings.Contains(funcName, "getCaller") {
				logPackage = getPackageName(funcName)
				break
			}
		}

		minimumCallerDepth = knownLogFrames
	})

	pcs := make([]uintptr, maximumCallerDepth)
	depth := runtime.Callers(minimumCallerDepth, pcs)
	frames := runtime.CallersFrames(pcs[:depth])

	for f, again := frames.Next(); again; f, again = frames.Next() {
		if pkg := getPackageName(f.Function); pkg != logPackage {
			return &f
		}
	}

	return nil
}
