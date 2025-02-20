package endpoint

import (
	"context"
	"net/http"

	"github.com/Falokut/go-kit/http/types"
	"github.com/pkg/errors"
)

func ContextParam() ParamMapper {
	return ParamMapper{
		Type: "context.Context",
		Builder: func(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error) {
			return ctx, nil
		},
	}
}

func ResponseWriterParam() ParamMapper {
	return ParamMapper{
		Type: "http.ResponseWriter",
		Builder: func(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error) {
			return w, nil
		},
	}
}

func RequestParam() ParamMapper {
	return ParamMapper{
		Type: "*http.Request",
		Builder: func(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error) {
			return r, nil
		},
	}
}

func RangeParam() ParamMapper {
	return ParamMapper{
		Type: "*types.RangeOption",
		Builder: func(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error) {
			var rangeOption *types.RangeOption
			rangeHeader := r.Header.Get("Range")
			if rangeHeader == "" {
				return rangeOption, nil
			}

			rangeOption = &types.RangeOption{}
			err := rangeOption.FromHeader(rangeHeader)
			if err != nil {
				return nil, errors.WithMessage(err, "parse range header")
			}
			return rangeOption, nil
		},
	}
}
