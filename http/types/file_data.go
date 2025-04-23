package types

import (
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

type PartialDataInfo struct {
	RangeStartByte int64
	RangeEndByte   int64
}

type FileData struct {
	PartialDataInfo *PartialDataInfo
	ContentType     string
	TotalFileSize   int64
	ContentReader   io.ReadSeekCloser
}

func (file FileData) Write(w http.ResponseWriter) error {
	defer file.ContentReader.Close()
	if file.PartialDataInfo != nil {
		return file.writePartialData(w)
	}

	w.Header().Set("Content-Type", file.ContentType)
	w.WriteHeader(http.StatusOK)

	_, err := io.Copy(w, file.ContentReader)
	if err != nil {
		return errors.WithMessage(err, "error during content copying")
	}
	return nil
}

func (file FileData) writePartialData(w http.ResponseWriter) error {
	partialInfo := file.PartialDataInfo
	startByte := partialInfo.RangeStartByte
	endByte := partialInfo.RangeEndByte

	if endByte == 0 || endByte >= file.TotalFileSize {
		endByte = file.TotalFileSize - 1
	}

	contentLength := endByte - startByte + 1

	w.Header().Set("Content-Type", file.ContentType)
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", startByte, endByte, file.TotalFileSize))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", contentLength))
	w.WriteHeader(http.StatusPartialContent)

	_, err := file.ContentReader.Seek(startByte, io.SeekStart)
	if err != nil {
		return errors.WithMessage(err, "failed to seek start byte")
	}

	_, err = io.CopyN(w, file.ContentReader, contentLength)
	if err != nil {
		return errors.WithMessage(err, "error during content copying")
	}

	return nil
}
