package endpoint

import (
	"context"
	"net/http"
	"strings"

	http2 "github.com/Falokut/go-kit/http"
	"github.com/Falokut/go-kit/http/apierrors"
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
			rangeHeader := r.Header.Get(http2.RangeHeader)
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

func BearerTokenParam() ParamMapper {
	return ParamMapper{
		Type: "types.BearerToken",
		Builder: func(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error) {
			token := r.Header.Get(http2.AuthorizationHeader)
			tokenParts := strings.Split(token, " ")
			if len(tokenParts) != 2 ||
				tokenParts[0] != http2.BearerToken ||
				tokenParts[1] == "" {
				return types.BearerToken{}, apierrors.NewUnauthorizedError("invalid token")
			}
			return types.BearerToken{Token: tokenParts[1]}, nil
		},
	}
}
