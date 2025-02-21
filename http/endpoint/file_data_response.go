package endpoint

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/pkg/errors"

	"github.com/Falokut/go-kit/http/types"
)

type FileDataResponseMapper struct {
}

func (m FileDataResponseMapper) Map(file types.FileData, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", file.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(file.ContentSize, 10))
	if file.PartialDataInfo != nil {
		lastByte := file.PartialDataInfo.RangeStartByte + file.ContentSize - 1
		if lastByte > file.PartialDataInfo.TotalDataSize {
			lastByte = file.PartialDataInfo.TotalDataSize - 1
		}
		w.Header().Set("Content-Range",
			fmt.Sprintf("bytes %d-%d/%d",
				file.PartialDataInfo.RangeStartByte,
				lastByte,
				file.PartialDataInfo.TotalDataSize,
			),
		)
		w.Header().Set("Accept-Ranges", "bytes")
		w.WriteHeader(http.StatusPartialContent)
	}

	_, err := w.Write(file.Content)
	if err != nil {
		return errors.WithMessage(err, "write file content")
	}
	return nil
}
