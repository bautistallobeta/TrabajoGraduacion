package middlewares

import (
	"MSTransaccionesFinancieras/internal/models"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

const ClaveActor = "Actor"
const ClaveCredencial = "Credencial"

func AutenticacionDual() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			apiKey := c.Request().Header.Get("X-API-Key")
			authHeader := c.Request().Header.Get("Authorization")

			if apiKey != "" && authHeader != "" {
				return c.JSON(
					http.StatusBadRequest,
					models.NewErrorRespuesta("Solo se permite un método de autenticación"),
				)
			}

			if apiKey == "" && authHeader == "" {
				return c.JSON(
					http.StatusUnauthorized,
					models.NewErrorRespuesta("No Autorizado"),
				)
			}

			if apiKey != "" {
				c.Set(ClaveActor, "SISTEMA")
				c.Set(ClaveCredencial, apiKey)
			} else {
				partes := strings.SplitN(authHeader, " ", 2)
				if len(partes) != 2 || strings.ToLower(partes[0]) != "bearer" || partes[1] == "" {
					return c.JSON(
						http.StatusUnauthorized,
						models.NewErrorRespuesta("No Autorizado"),
					)
				}
				c.Set(ClaveActor, "USUARIO")
				c.Set(ClaveCredencial, partes[1])
			}

			return next(c)
		}
	}
}
