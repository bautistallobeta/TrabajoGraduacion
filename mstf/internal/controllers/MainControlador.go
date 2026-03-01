package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type MainControlador struct{}

func NewMainControlador() *MainControlador {
	return &MainControlador{}
}

func (mc *MainControlador) Ping(c echo.Context) error {
	return c.String(http.StatusOK, "API MSTF corriendo.")
}
