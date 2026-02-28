package middlewares

import (
	"MSTransaccionesFinancieras/internal/auth"
	"MSTransaccionesFinancieras/internal/models"
	"context"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const ClaveActor = "Actor"
const ClaveCredencial = "Credencial"

func AutenticacionDual(skipper middleware.Skipper) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if skipper(c) {
				return next(c)
			}

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

			var credencial, actor string
			if apiKey != "" {
				actor = "SISTEMA"
				credencial = apiKey
			} else {
				partes := strings.SplitN(authHeader, " ", 2)
				if len(partes) != 2 || strings.ToLower(partes[0]) != "bearer" || partes[1] == "" {
					return c.JSON(
						http.StatusUnauthorized,
						models.NewErrorRespuesta("No Autorizado"),
					)
				}
				actor = "USUARIO"
				credencial = partes[1]
			}

			if err := models.Autenticar(credencial, actor); err != nil {
				return c.JSON(http.StatusUnauthorized, models.NewErrorRespuesta(err.Error()))
			}

			c.Set(ClaveActor, actor)
			c.Set(ClaveCredencial, credencial)

			ctx := context.WithValue(c.Request().Context(), auth.ClaveCredencial, credencial)
			ctx = context.WithValue(ctx, auth.ClaveActor, actor)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}
