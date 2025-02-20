package types

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/Falokut/go-kit/http/apierrors"
)

type RangeOption struct {
	Start int64
	End   int64
}

func (o *RangeOption) Length(objSize int64) (int64, error) {
	var length int64
	switch {
	case o.End < 0:
		length = -o.End
	case o.End == 0:
		length = objSize - o.Start
	default:
		length = o.End - o.Start + 1
	}
	if length <= 0 {
		return -1, errors.New("invalid range: End must be greater than Start")
	}
	return length, nil
}

func (o *RangeOption) FromHeader(header string) error {
	if !strings.HasPrefix(header, "bytes=") {
		return rangeError("invalid range format")
	}

	rangePart := strings.TrimPrefix(header, "bytes=")
	parts := strings.Split(rangePart, "-")
	if len(parts) != 2 {
		return rangeError("invalid range format")
	}

	switch {
	case parts[0] == "" && parts[1] != "":
		endBytes, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
		if err != nil || endBytes <= 0 {
			return rangeError("invalid end byte for suffix range")
		}
		o.Start = 0
		o.End = -endBytes
	case parts[0] != "" && parts[1] == "":
		startVal, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
		if err != nil || startVal < 0 {
			return rangeError("invalid start byte")
		}
		o.Start = startVal
		o.End = 0
	case parts[0] != "" && parts[1] != "":
		startVal, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
		if err != nil || startVal < 0 {
			return rangeError("invalid start byte")
		}
		endVal, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
		if err != nil || endVal < startVal {
			return rangeError("invalid end byte or end < start")
		}
		o.Start = startVal
		o.End = endVal
	default:
		return rangeError("invalid range format")
	}
	return nil
}

func rangeError(errorMsg string) apierrors.Error {
	return apierrors.New(http.StatusRequestedRangeNotSatisfiable, apierrors.ErrCodeInvalidRange, errorMsg, errors.New(errorMsg))
}
