package model

import "errors"

var ErrorConflict = errors.New("conflict")
var ErrorIncorrectRequest = errors.New("incorrect request")
var ErrorNotEqual = errors.New("not equal")
var ErrorNotAuthorization = errors.New("not authorization")
var ErrorNotValidNumber = errors.New("not valid order")
var ErrorDoubleOperation = errors.New("double operation")
var ErrorNotContent = errors.New("not content")
var ErrorInsufficientFunds = errors.New("insufficient funds")
