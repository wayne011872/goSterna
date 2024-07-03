package util

import (
	"github.com/asaskevich/govalidator"
)

// InitValidator is used to initialize the validator.
// Call this funtion before using any validator.
func InitValidator() {
	govalidator.TagMap["required"] = govalidator.Validator(required)
	return
}

//Check if str is empty or not
func required(str string) bool {
	if len(str) > 0 {
		return true
	}
	return false
}

func IsStrInList(input string, target ...string) bool {
	for _, paramName := range target {
		if input == paramName {
			return true
		}
	}
	return false

}