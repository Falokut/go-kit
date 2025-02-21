package types

import (
	http2 "github.com/Falokut/go-kit/http"
	"github.com/Falokut/go-kit/http/apierrors"
	"net/http"
	"strings"
)

type BearerToken struct {
	Token string
}

func (b *BearerToken) FromRequestHeader(r *http.Request) error {
	token := r.Header.Get(http2.AuthorizationHeader)
	tokenParts := strings.Split(token, " ")
	if len(tokenParts) != 2 ||
		tokenParts[0] != http2.BearerToken ||
		tokenParts[1] == "" {
		return apierrors.NewUnauthorizedError("invalid token")
	}
	b.Token = tokenParts[1]
	return nil
}
