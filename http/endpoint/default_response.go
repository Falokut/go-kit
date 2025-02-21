package endpoint

import (
	"net/http"
)

type ResponseWriter interface {
	Write(w http.ResponseWriter) error
}

type DefaultResponseMapper struct {
	JsonResponseMapper
}

func (m DefaultResponseMapper) Map(result any, w http.ResponseWriter) error {
	if result == nil {
		return nil
	}
	writer, ok := result.(ResponseWriter)
	if ok {
		return writer.Write(w)
	}
	return m.JsonResponseMapper.Map(result, w)
}
