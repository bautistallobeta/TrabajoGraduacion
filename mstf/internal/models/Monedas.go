package models

import (
	"time"

	"MSTransaccionesFinancieras/internal/infra/persistence"
)

type Monedas struct {
	IdMoneda  int       `json:"IdMoneda"`
	Ledger    int       `json:"Ledger"`
	Estado    string    `json:"Estado"`
	FechaAlta time.Time `json:"FechaAlta"`
}

// Instancia los atributos de la moneda desde la base de datos.
// tsp_dame_moneda
func (m *Monedas) Dame(tokenSesion string) error {
	rows, err := persistence.ClienteMySQL.Query("CALL tsp_dame_moneda(?, ?)", tokenSesion, m.IdMoneda)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		return rows.Scan(&m.IdMoneda, &m.Ledger, &m.Estado, &m.FechaAlta)
	}

	return nil
}

// Cambia el estado de la moneda a baja.
// tsp_darbaja_moneda
func (m *Monedas) DarBaja(tokenSesion string) (string, error) {
	var mensaje string
	err := persistence.ClienteMySQL.QueryRow("CALL tsp_darbaja_moneda(?, ?)", tokenSesion, m.IdMoneda).Scan(&mensaje)
	return mensaje, err
}

// Cambia el estado de la moneda a activa.
// tsp_activar_moneda
func (m *Monedas) Activar(tokenSesion string) (string, error) {
	var mensaje string
	err := persistence.ClienteMySQL.QueryRow("CALL tsp_activar_moneda(?, ?)", tokenSesion, m.IdMoneda).Scan(&mensaje)
	return mensaje, err
}
