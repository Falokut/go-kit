package types

import (
	"fmt"
	"net/http"
	"strconv"

	http2 "github.com/Falokut/go-kit/http"
	"github.com/pkg/errors"
)

type FileData struct {
	PartialDataInfo *PartialDataInfo
	ContentType     string
	ContentSize     int64
	Content         []byte
}

type PartialDataInfo struct {
	RangeStartByte int64
	TotalDataSize  int64
}

func (file FileData) Write(w http.ResponseWriter) error {
	w.Header().Set(http2.ContentTypeHeader, file.ContentType)
	w.Header().Set(http2.ContentLengthHeader, strconv.FormatInt(file.ContentSize, 10))
	if file.PartialDataInfo != nil {
		lastByte := file.PartialDataInfo.RangeStartByte + file.ContentSize - 1
		if lastByte > file.PartialDataInfo.TotalDataSize {
			lastByte = file.PartialDataInfo.TotalDataSize - 1
		}
		w.Header().Set(http2.ContentRangeHeader,
			fmt.Sprintf("%s %d-%d/%d",
				http2.BytesRange,
				file.PartialDataInfo.RangeStartByte,
				lastByte,
				file.PartialDataInfo.TotalDataSize,
			),
		)
		w.Header().Set(http2.AcceptRangeHeader, http2.BytesRange)
		w.WriteHeader(http.StatusPartialContent)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	_, err := w.Write(file.Content)
	if err != nil {
		return errors.WithMessage(err, "write file content")
	}
	return nil
}
