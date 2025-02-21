package endpoint

import (
	"net/http"

	"github.com/Falokut/go-kit/http/types"
)

type TypeResponseMapper struct {
	JsonResponseMapper
	FileDataResponseMapper
}

func (m TypeResponseMapper) Map(result any, w http.ResponseWriter) error {
	switch result := result.(type) {
	case nil:
		return nil
	case types.FileData:
		return m.FileDataResponseMapper.Map(result, w)
	case *types.FileData:
		return m.FileDataResponseMapper.Map(*result, w)
	default:
		return m.JsonResponseMapper.Map(result, w)
	}
}
