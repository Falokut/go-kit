package types

import (
	"fmt"
	"io"
	"net/http"

	http2 "github.com/Falokut/go-kit/http"
	"github.com/pkg/errors"
)

type PartialDataInfo struct {
	RangeStartByte int64
	RangeEndByte   int64
	TotalDataSize  int64
}

type FileData struct {
	PartialDataInfo *PartialDataInfo
	ContentType     string
	TotalFileSize   int64
	ContentReader   io.Reader
}

func (file FileData) Write(w http.ResponseWriter) error {
	w.Header().Set(http2.ContentTypeHeader, file.ContentType)
	partialInfo := file.PartialDataInfo
	if partialInfo != nil {
		endByte := partialInfo.RangeEndByte
		if endByte >= file.TotalFileSize {
			endByte = file.TotalFileSize - 1
		}
		w.Header().Set(http2.ContentRangeHeader,
			fmt.Sprintf("%s %d-%d/%d",
				http2.BytesRange,
				partialInfo.RangeStartByte,
				endByte,
				file.TotalFileSize,
			),
		)
		w.Header().Set(http2.AcceptRangeHeader, http2.BytesRange)
		w.WriteHeader(http.StatusPartialContent)
	}
	_, err := io.Copy(w, file.ContentReader)
	if err != nil {
		return errors.WithMessage(err, "copy from content reader")
	}
	return nil
}
