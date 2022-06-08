package validator

import (
	"fmt"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

// Validator is a structure that wraps the third party validator package
// This package will allow for struct validation and pretty printing of error messages
type Validator struct {
	v               *validator.Validate
	errorTranslator *ut.Translator
}

func New() *Validator {

	validate := validator.New()
	english := en.New()
	uni := ut.New(english, english)
	translator, _ := uni.GetTranslator("en")
	en_translations.RegisterDefaultTranslations(validate, translator)

	return &Validator{
		v:               validate,
		errorTranslator: &translator,
	}
}

func (val *Validator) ValidateStruct(s interface{}) *string {
	return val.FormatErrors(val.v.Struct(s))
}

// FormatErrors translates the errors into human readible errors
func (val *Validator) FormatErrors(err error) *string {
	if err == nil {
		return nil
	}

	errStr := ""

	validatorErrs := err.(validator.ValidationErrors)
	for _, e := range validatorErrs {
		translatedErr := fmt.Errorf(e.Translate(*val.errorTranslator))
		errStr = errStr + fmt.Sprintf("%s. ", translatedErr.Error())
	}

	return &errStr
}
