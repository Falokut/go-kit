package healthcheck

const (
	StatusPass = "pass"
	StatusFail = "fail"
)

type Result struct {
	Status      string
	FailDetails map[string]any
}
