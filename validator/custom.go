package validator

import (
	"net"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

const (
	hostPortValidatorTag = "hostport"
)

func registerCustomValidations(validator *validator.Validate) error {
	err := validator.RegisterValidation(hostPortValidatorTag, ValidateHostPortOrPort)
	if err != nil {
		return errors.WithMessagef(err, "register '%s' validator tag", hostPortValidatorTag)
	}
	return nil
}

func ValidateHostPortOrPort(fl validator.FieldLevel) bool {
	value := fl.Field().String()

	// Check if it's ":port"
	if strings.HasPrefix(value, ":") {
		_, portErr := net.LookupPort("tcp", value[1:])
		return portErr == nil
	}

	// Check if it's "host:port" or "ip:port"
	host, port, err := net.SplitHostPort(value)
	if err != nil {
		return false
	}
	if port == "" {
		return false
	}
	_, portErr := net.LookupPort("tcp", port)
	if portErr != nil {
		return false
	}
	return host != ""
}
