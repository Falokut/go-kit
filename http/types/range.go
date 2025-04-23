package types

import (
	"errors"
	"strconv"
	"strings"

	"github.com/Falokut/go-kit/http/apierrors"
)

var (
	ErrInvalidRange = errors.New("invalid range: End must be greater than Start")
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
		return -1, ErrInvalidRange
	}
	return length, nil
}

// nolint:cyclop,mnd
func (o *RangeOption) FromHeader(header string) error {
	if !strings.HasPrefix(header, "bytes=") {
		return apierrors.NewRangeUnacceptableError("invalid range format")
	}

	rangePart := strings.TrimPrefix(header, "bytes=")
	parts := strings.Split(rangePart, "-")
	if len(parts) != 2 {
		return apierrors.NewRangeUnacceptableError("invalid range format")
	}

	switch {
	case parts[0] == "" && parts[1] != "":
		endBytes, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
		if err != nil || endBytes <= 0 {
			return apierrors.NewRangeUnacceptableError("invalid end byte for suffix range")
		}
		o.Start = 0
		o.End = -endBytes
	case parts[0] != "" && parts[1] == "":
		startVal, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
		if err != nil || startVal < 0 {
			return apierrors.NewRangeUnacceptableError("invalid start byte")
		}
		o.Start = startVal
		o.End = 0
	case parts[0] != "" && parts[1] != "":
		startVal, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
		if err != nil || startVal < 0 {
			return apierrors.NewRangeUnacceptableError("invalid start byte")
		}
		endVal, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
		if err != nil || endVal < startVal {
			return apierrors.NewRangeUnacceptableError("invalid end byte or end < start")
		}
		o.Start = startVal
		o.End = endVal
	default:
		return apierrors.NewRangeUnacceptableError("invalid range format")
	}
	return nil
}
