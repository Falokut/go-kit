package log

import (
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	logPackage string

	minimumCallerDepth int

	callerInitOnce sync.Once
)

const (
	maximumCallerDepth int = 25
	knownLogFrames     int = 4
)

type Entry struct {
	ReportCaller bool
	Data         Fields
	Time         time.Time
	Level        Level
	Caller       *runtime.Frame
	Message      string
}

func getPackageName(f string) string {
	for {
		lastPeriod := strings.LastIndex(f, ".")
		lastSlash := strings.LastIndex(f, "/")
		if lastPeriod > lastSlash {
			f = f[:lastPeriod]
		} else {
			break
		}
	}

	return f
}

func getCaller() *runtime.Frame {
	callerInitOnce.Do(func() {
		pcs := make([]uintptr, maximumCallerDepth)
		_ = runtime.Callers(0, pcs)

		for i := 0; i < maximumCallerDepth; i++ {
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
		pkg := getPackageName(f.Function)

		if pkg != logPackage {
			return &f //nolint:scopelint
		}
	}

	return nil
}

func (entry Entry) HasCaller() (has bool) {
	return entry.ReportCaller &&
		entry.Caller != nil
}
