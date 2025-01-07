package main

import (
	"time"

	"github.com/harnyk/gena"
)

type CurrentTimeHandler struct{}

func NewCurrentTimeHandler() *CurrentTimeHandler {
	return &CurrentTimeHandler{}
}

func (h CurrentTimeHandler) Execute(params gena.H) (any, error) {
	return time.Now().String(), nil
}
