package models

import (
	"MSTransaccionesFinancieras/internal/infra/cache"
	"MSTransaccionesFinancieras/internal/infra/persistence"
	"database/sql"
	"errors"
	"log"
	"strconv"
	"time"
)

type Monedas struct {
	IdMoneda        int       `json:"IdMoneda"`
	IdCuentaEmpresa string    `json:"IdCuentaEmpresa"`
	Estado          string    `json:"Estado"`
	FechaAlta       time.Time `json:"FechaAlta"`
}

var CacheMonedas = cache.NewCache[Monedas](15 * time.Second)

// Instancia los atributos de la moneda desde la base de datos.
// tsp_dame_moneda
// - tokenSesion: token de sesión del usuario
func (m *Monedas) Dame(tokenSesion string) (string, error) {
	clave := strconv.Itoa(m.IdMoneda)
	if cached, ok := CacheMonedas.Dame(clave); ok {
		*m = cached
		log.Printf("[CACHE HIT] Moneda %s obtenida del caché", clave)
		return "OK", nil
	}

	rows, err := persistence.ClienteMySQL.Query("CALL tsp_dame_moneda(?, ?)", tokenSesion, m.IdMoneda)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	var mensaje string
	var idMoneda sql.NullInt32
	var idCuentaEmpresa sql.NullString
	var estado sql.NullString
	var fechaAlta sql.NullTime
	if rows.Next() {
		err = rows.Scan(&mensaje, &idMoneda, &idCuentaEmpresa, &estado, &fechaAlta)

		if idMoneda.Valid {
			m.IdMoneda = int(idMoneda.Int32)
		} else {
			m.IdMoneda = 0
		}
		if idCuentaEmpresa.Valid {
			m.IdCuentaEmpresa = idCuentaEmpresa.String
		} else {
			m.IdCuentaEmpresa = ""
		}
		if fechaAlta.Valid {
			m.FechaAlta = fechaAlta.Time
		} else {
			m.FechaAlta = time.Time{}
		}
		if estado.Valid {
			m.Estado = estado.String
		} else {
			m.Estado = ""
		}
		if err != nil {
			return mensaje, err
		}
		if mensaje != "OK" {
			return mensaje, errors.New(mensaje)
		}
		CacheMonedas.Guardar(clave, *m)
		log.Printf("[CACHE MISS] Moneda %s obtenida de MySQL y guardada en caché", clave)
	}
	return mensaje, nil
}

// Activa una moneda pendiente asignando la cuenta empresa.
// tsp_activar_moneda
// - idCuentaEmpresa: Id de la cuenta empresa en TigerBeetle
func (m *Monedas) Activar(idCuentaEmpresa string) (string, error) {
	var mensaje string
	err := persistence.ClienteMySQL.QueryRow("CALL tsp_activar_moneda(?, ?)", idCuentaEmpresa, m.IdMoneda).Scan(&mensaje)
	if err != nil {
		return "", err
	}
	CacheMonedas.Borrar(strconv.Itoa(m.IdMoneda))
	log.Printf("[CACHE INVALIDADO] Moneda %d borrada del caché por activación", m.IdMoneda)
	return mensaje, nil
}

// Desactiva una moneda activa.
// tsp_desactivar_moneda
// - tokenSesion: token de sesión del usuario
func (m *Monedas) Desactivar(tokenSesion string) (string, error) {
	var mensaje string
	err := persistence.ClienteMySQL.QueryRow("CALL tsp_desactivar_moneda(?, ?)", tokenSesion, m.IdMoneda).Scan(&mensaje)
	if err != nil {
		return "", err
	}
	CacheMonedas.Borrar(strconv.Itoa(m.IdMoneda))
	log.Printf("[CACHE INVALIDADO] Moneda %d borrada del caché por desactivación", m.IdMoneda)
	return mensaje, nil
}
