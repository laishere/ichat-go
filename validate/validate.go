package validate

import (
	"errors"
	"github.com/gin-gonic/gin/binding"
	zh2 "github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/go-playground/validator/v10/translations/zh"
	"ichat-go/errs"
	"reflect"
	"strings"
)

var validate *validator.Validate
var translator ut.Translator

func init() {
	binding.Validator = &myValidator{}
	validate = validator.New(validator.WithRequiredStructEnabled())
	_ = validate.RegisterValidation("url", validateUrl)
	z := zh2.New()
	uni := ut.New(z)
	trans, _ := uni.GetTranslator("zh")
	_ = zh.RegisterDefaultTranslations(validate, trans)
	translator = trans
}

func HandleError(e error) {
	var ve validator.ValidationErrors
	if errors.As(e, &ve) {
		var _errs []string
		for _, v := range ve.Translate(translator) {
			_errs = append(_errs, v)
		}
		panic(errs.NewVerificationError(strings.Join(_errs, ",")))
	}
}

type myValidator struct {
}

func (*myValidator) ValidateStruct(obj any) error {
	return validate.Struct(obj)
}

func (*myValidator) Engine() any {
	return validate
}

func validateUrl(fl validator.FieldLevel) bool {
	filed := fl.Field()
	if filed.Kind() != reflect.String {
		return false
	}
	s := filed.String()
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		return true
	}
	return strings.HasPrefix(s, "file/")
}
