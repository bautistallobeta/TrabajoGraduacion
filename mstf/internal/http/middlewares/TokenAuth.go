package middlewares

import (
	"MSTransaccionesFinancieras/internal/models"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func TokenAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(
					http.StatusUnauthorized,
					models.NewErrorRespuesta("Falta header Authorization"),
				)
			}

			partes := strings.Split(authHeader, " ")
			if len(partes) != 2 || strings.ToLower(partes[0]) != "bearer" {
				return c.JSON(
					http.StatusUnauthorized,
					models.NewErrorRespuesta("Formato de Authorization inválido"),
				)
			}

			token := partes[1]
			if token == "" {
				return c.JSON(
					http.StatusUnauthorized,
					models.NewErrorRespuesta("Token vacío"),
				)
			}

			c.Set("adminToken", token)

			return next(c)
		}
	}
}
