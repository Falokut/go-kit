package validator

import (
	"errors"
	"fmt"
	en_translations "github.com/Falokut/go-kit/validator/translations/en"
	ru_translations "github.com/Falokut/go-kit/validator/translations/ru"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/ru"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"strings"
)

type Language string

const (
	En Language = "en"
	Ru Language = "ru"
)

type Adapter struct {
	validator  *validator.Validate
	translator ut.Translator
}

func New(lang Language) Adapter {
	validator := validator.New()
	err := registerCustomValidations(validator)
	if err != nil {
		panic(err)
	}

	// nolint:exhaustive
	switch lang {
	case Ru:
		ruTranslator := ru.New()
		uni := ut.New(ruTranslator, ruTranslator)
		translator, _ := uni.GetTranslator(string(lang))
		err = ru_translations.RegisterDefaultTranslations(validator, translator)
		if err != nil {
			panic(err)
		}
		return Adapter{
			validator:  validator,
			translator: translator,
		}
	default:
		lang = "en"
		enTranslator := en.New()
		uni := ut.New(enTranslator, enTranslator)
		translator, _ := uni.GetTranslator(string(lang))
		err = en_translations.RegisterDefaultTranslations(validator, translator)
		if err != nil {
			panic(err)
		}
		return Adapter{
			validator:  validator,
			translator: translator,
		}
	}
}

type wrapper struct {
	V any
}

func (a Adapter) Validate(v any) (bool, map[string]string) {
	err := a.validator.Struct(wrapper{v}) // hack
	if err == nil {
		return true, nil
	}
	details, err := a.collectDetails(err)
	if err != nil {
		return false, map[string]string{"#validator": err.Error()}
	}
	return false, details
}

func (a Adapter) ValidateToError(v any) error {
	ok, details := a.Validate(v)
	if ok {
		return nil
	}
	descriptions := make([]string, 0, len(details))
	for field, err := range details {
		descriptions = append(descriptions, fmt.Sprintf("%s -> %s", field, err))
	}
	err := strings.Join(descriptions, "; ")
	return errors.New(err) // nolint:err113
}

const (
	prefixToDelete = "wrapper.V"
)

func (a Adapter) collectDetails(err error) (map[string]string, error) {
	var e validator.ValidationErrors
	if !errors.As(err, &e) {
		return nil, err
	}
	result := make(map[string]string, len(e))
	for _, err := range e {
		field := []byte(err.Namespace())[len(prefixToDelete):]
		if field[0] == '.' {
			field = field[1:]
		}
		firstLetter := 0
		for i := 0; i < len(field); i++ {
			if field[i] == '.' {
				field[firstLetter] = strings.ToLower(string(field[firstLetter]))[0]
				firstLetter = i + 1
			}
		}
		field[firstLetter] = strings.ToLower(string(field[firstLetter]))[0]
		result[string(field)] = err.Translate(a.translator)
	}
	return result, nil
}
