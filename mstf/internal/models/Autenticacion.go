package models

import (
	"MSTransaccionesFinancieras/internal/infra/cache"
	"MSTransaccionesFinancieras/internal/infra/persistence"
	"errors"
	"time"
)

var cacheApiKeys = cache.NewCache[bool](5 * time.Minute)

// Valida las creadenciales del actor (SISTEMA o USUARIO) )
// Para SISTEMA: cachea la API key durante 5 minutos para reducir roundtrips.
// Para USUARIO: siempre consulta la DB
func Autenticar(credencial string, actor string) error {
	if actor == "SISTEMA" {
		if _, ok := cacheApiKeys.Dame(credencial); ok {
			return nil
		}
	}

	var mensaje string
	err := persistence.ClienteMySQL.QueryRow("CALL tsp_autenticar_actor(?, ?)", credencial, actor).Scan(&mensaje)
	if err != nil {
		return errors.New("Error de autenticación")
	}
	if mensaje != "OK" {
		return errors.New(mensaje)
	}

	if actor == "SISTEMA" {
		cacheApiKeys.Guardar(credencial, true)
	}

	return nil
}
