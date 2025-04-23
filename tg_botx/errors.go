package tg_botx

type RetryWithDelayError struct {
	msg string
}

func NewRetryWithDelayError(msg string) RetryWithDelayError {
	return RetryWithDelayError{
		msg: msg,
	}
}

func (e RetryWithDelayError) Error() string {
	return e.msg
}
