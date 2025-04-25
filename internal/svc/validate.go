package svc

import (
	"reflect"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	"github.com/pkg/errors"
)

func (svc ServiceContext) Validate(dataStruct interface{}) error {
	enUs := en.New()
	validate := validator.New()
	//  RegisterTagNameFunc registers a function to get field name for error messages
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := fld.Tag.Get("label")
		return name
	})

	uni := ut.New(enUs)
	trans, _ := uni.GetTranslator("en")
	// RegisterDefaultTranslations registers the default translations
	_ = enTranslations.RegisterDefaultTranslations(validate, trans)
	err := validate.Struct(dataStruct)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			return errors.New(err.Translate(trans))
		}
	}
	return nil
}
