package fiberx

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v3"
	fbaerrors "github.com/yuWorm/fba-go/core/errors"
)

func ValidationMissingField(field string) error {
	return validationError(fmt.Sprintf("请求参数非法: %s 字段为必填项，输入：None", field))
}

func ValidationIntParsing(field string, input string) error {
	return validationError(fmt.Sprintf("请求参数非法: %s 输入应为有效的整数，无法将字符串解析为整数，输入：%s", field, input))
}

func ParseIntParam(field string, input string) (int, error) {
	value, err := strconv.Atoi(input)
	if err != nil {
		// Match FastAPI/Pydantic path validation for Annotated[int, Path(...)]
		// parameters before business handlers run.
		return 0, ValidationIntParsing(field, input)
	}
	return value, nil
}

func validationError(message string) error {
	return fbaerrors.New(fiber.StatusUnprocessableEntity, fiber.StatusUnprocessableEntity, message, nil)
}
