package requestid

import "github.com/google/uuid"

func Next() string {
	return uuid.NewString()
}
