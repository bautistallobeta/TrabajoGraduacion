package models

import (
	"MSTransaccionesFinancieras/internal/infra/persistence"
	"database/sql"
	"errors"
	"time"
)

type Monedas struct {
	IdMoneda        int            `json:"IdMoneda"`
	Ledger          int            `json:"Ledger"`
	IdCuentaEmpresa string         `json:"IdCuentaEmpresa"`
	Estado          string         `json:"Estado"`
	FechaAlta       time.Time      `json:"FechaAlta"`
}

// Instancia los atributos de la moneda desde la base de datos.
// tsp_dame_moneda
// - tokenSesion: token de sesión del usuario
func (m *Monedas) Dame(db *sql.DB, tokenSesion string) (string, error) {
	rows, err := persistence.ClienteMySQL.Query("CALL tsp_dame_moneda(?)", tokenSesion)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	var mensaje string
	var idMoneda sql.NullInt32
	var ledger sql.NullInt32
	var idCuentaEmpresa sql.NullString
	var estado sql.NullString
	var fechaAlta sql.NullTime
	if rows.Next() {
		err = rows.Scan(&mensaje, &idMoneda, &ledger, &idCuentaEmpresa, &estado, &fechaAlta)

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
	}
	return mensaje, nil
}

// Instancia los atributos de la moneda desde la base de datos a partir de su ledger.
// tsp_dame_moneda_por_ledger
// - tokenSesion: token de sesión del usuario
// - ledger: ledger de TigerBeetle
func (m *Monedas) DamePorLedger(db *sql.DB, tokenSesion string, ledger int) (string, error) {
	var idMoneda sql.NullInt32
	var mensaje string
	var idCuentaEmpresa sql.NullString
	var estado sql.NullString
	var fechaAlta sql.NullTime
	err := db.QueryRow("CALL tsp_dame_moneda_por_ledger(?, ?)", tokenSesion, ledger).Scan(&mensaje, &idMoneda, &m.Ledger, &idCuentaEmpresa, &estado, &fechaAlta)
	if err != nil {
		return "", err
	}
	if mensaje != "OK" {
		return mensaje, errors.New(mensaje)
	}
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
	if estado.Valid {
		m.Estado = estado.String
	} else {
		m.Estado = ""
	}
	if fechaAlta.Valid {
		m.FechaAlta = fechaAlta.Time
	} else {
		m.FechaAlta = time.Time{}
	}

	return mensaje, nil
}

// Activa una moneda pendiente asignando la cuenta empresa.
// tsp_activar_moneda
// - idCuentaEmpresa: Id de la cuenta empresa en TigerBeetle
func (m *Monedas) Activar(db *sql.DB, idCuentaEmpresa string) (string, error) {
	var mensaje string
	err := db.QueryRow("CALL tsp_activar_moneda(?, ?)", idCuentaEmpresa, m.IdMoneda).Scan(&mensaje)
	if err != nil {
		return "", err
	}
	return mensaje, nil
}

// Desactiva una moneda activa.
// tsp_desactivar_moneda
// - tokenSesion: token de sesión del usuario
func (m *Monedas) Desactivar(db *sql.DB, tokenSesion string) (string, error) {
	var mensaje string
	err := db.QueryRow("CALL tsp_desactivar_moneda(?, ?)", tokenSesion, m.IdMoneda).Scan(&mensaje)
	if err != nil {
		return "", err
	}
	return mensaje, nil
}
