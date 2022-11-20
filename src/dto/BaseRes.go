package dto

import (
	"net/http"
	"strconv"
)

type BaseRes struct {
	Code string      `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func Success() *BaseRes {

	var baseRes BaseRes
	baseRes.Code = strconv.Itoa(http.StatusOK)
	baseRes.Msg = "success"

	return &baseRes
}

func SuccessForData(data interface{}) *BaseRes {

	var baseRes BaseRes
	baseRes.Code = strconv.Itoa(http.StatusOK)
	baseRes.Msg = "success"
	baseRes.Data = data

	return &baseRes
}

func Fail(code string, msg string) *BaseRes {

	var baseRes BaseRes
	baseRes.Code = code
	baseRes.Msg = msg

	return &baseRes
}
