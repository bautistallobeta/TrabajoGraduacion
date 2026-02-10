package gestores

import (
	"database/sql"

	"MSTransaccionesFinancieras/internal/models"
)

type GestorMonedas struct {
	Db *sql.DB
}

func NewGestorMonedas(db *sql.DB) *GestorMonedas {
	return &GestorMonedas{Db: db}
}

// Da de alta una nueva moneda.
// tsp_alta_moneda
func (gm *GestorMonedas) Crear(tokenSesion string, idMoneda int, ledger int) (string, int, error) {
	var mensaje string
	var id *int

	err := gm.Db.QueryRow("CALL tsp_alta_moneda(?, ?, ?)", tokenSesion, idMoneda, ledger).Scan(&mensaje, &id)
	if err != nil {
		return "", 0, err
	}

	if id == nil {
		return mensaje, 0, nil
	}

	return mensaje, *id, nil
}

// Lista las monedas.
// tsp_buscar_monedas
func (gm *GestorMonedas) Buscar(tokenSesion string, incluyeBajas string) ([]*models.Monedas, error) {
	rows, err := gm.Db.Query("CALL tsp_buscar_monedas(?, ?)", tokenSesion, incluyeBajas)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var monedas []*models.Monedas

	for rows.Next() {
		var m models.Monedas
		err = rows.Scan(&m.IdMoneda, &m.Ledger, &m.Estado, &m.FechaAlta)
		if err != nil {
			return nil, err
		}
		monedas = append(monedas, &m)
	}

	return monedas, nil
}
