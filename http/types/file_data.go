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
	ContentReader   io.ReadCloser
}

// Write отправляет файл в HTTP-ответ.
// Поддерживает полный файл или частичный Range.
func (file FileData) Write(w http.ResponseWriter) error {
	defer file.ContentReader.Close()

	if file.PartialDataInfo != nil {
		return file.writePartialData(w)
	}

	// Полный файл
	w.Header().Set("Content-Type", file.ContentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", file.TotalFileSize))
	w.WriteHeader(http.StatusOK)

	_, err := io.Copy(w, file.ContentReader)
	if err != nil {
		return errors.WithMessage(err, "error during content copying")
	}

	return nil
}

// writePartialData отправляет только указанный диапазон байт
func (file FileData) writePartialData(w http.ResponseWriter) error {
	partial := file.PartialDataInfo
	start := partial.RangeStartByte
	end := partial.RangeEndByte

	if end == 0 || end >= file.TotalFileSize {
		end = file.TotalFileSize - 1
	}

	contentLength := end - start + 1

	w.Header().Set("Content-Type", file.ContentType)
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, file.TotalFileSize))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", contentLength))
	w.WriteHeader(http.StatusPartialContent)

	_, err := io.Copy(w, file.ContentReader)
	if err != nil {
		return errors.WithMessage(err, "error during content copying")
	}

	return nil
}
