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
	PrettyName      string
	ContentType     string
	TotalFileSize   int64
	ContentReader   io.ReadSeekCloser
}

// Write отправляет содержимое файла в ResponseWriter, поддерживая частичные загрузки.
func (file *FileData) Write(w http.ResponseWriter) error {
	if file.ContentReader == nil {
		return errors.New("ContentReader is nil")
	}
	defer file.ContentReader.Close()

	if file.PartialDataInfo != nil {
		return file.writePartialData(w)
	}

	file.setHeaders(w)
	w.WriteHeader(http.StatusOK)

	_, err := io.Copy(w, file.ContentReader)
	if err != nil {
		return errors.WithMessage(err, "error during content copying")
	}
	return nil
}

// writePartialData отправляет только часть файла, согласно PartialDataInfo.
func (file *FileData) writePartialData(w http.ResponseWriter) error {
	start := file.PartialDataInfo.RangeStartByte
	end := file.PartialDataInfo.RangeEndByte

	if start >= file.TotalFileSize {
		return errors.New("start byte out of range")
	}
	if end == 0 || end >= file.TotalFileSize {
		end = file.TotalFileSize - 1
	}

	contentLength := end - start + 1

	file.setHeaders(w)
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, file.TotalFileSize))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", contentLength))
	w.WriteHeader(http.StatusPartialContent)

	_, err := file.ContentReader.Seek(start, io.SeekStart)
	if err != nil {
		return errors.WithMessage(err, "failed to seek start byte")
	}

	_, err = io.CopyN(w, file.ContentReader, contentLength)
	if err != nil && !errors.Is(err, io.EOF) {
		return errors.WithMessage(err, "error during content copying")
	}

	return nil
}

// setHeaders устанавливает основные HTTP-заголовки для файла.
func (file *FileData) setHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", file.ContentType)
	if file.PrettyName != "" {
		w.Header().Set("Content-Disposition", fmt.Sprintf(`filename="%s"`, file.PrettyName))
	}
}
