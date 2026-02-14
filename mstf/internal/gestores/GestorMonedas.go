package gestores

import (
	"MSTransaccionesFinancieras/internal/models"
	"database/sql"
)

type GestorMonedas struct {
	Db *sql.DB
}

func NewGestorMonedas(db *sql.DB) *GestorMonedas {
	return &GestorMonedas{Db: db}
}

// Crea una moneda en estado P: Pendiente y le asocia un ledger en TigerBeetle.
// tsp_crear_moneda
// - tokenSesion: token de sesión del usuario
// - idMoneda: Id de la moneda a crear (viene de MisGastos)
func (gm *GestorMonedas) Crear(tokenSesion string, idMoneda int) (string, error) {
	var mensaje string
	err := gm.Db.QueryRow("CALL tsp_crear_moneda(?, ?)", tokenSesion, idMoneda).Scan(&mensaje)
	if err != nil {
		return "", err
	}
	return mensaje, nil
}

//	Permite listar todas las monedas. Si pIncluyeInactivas es 'S', muestra todas.
//
// Si es 'N', muestra solo las activas. Ordena por IdMoneda.
// tsp_listar_monedas
// - tokenSesion: token de sesión del usuario
// - incluyeBajas: 'S' o 'N' para incluir o no las monedas bajas
func (gm *GestorMonedas) Listar(tokenSesion string, incluyeBajas string) ([]models.Monedas, error) {
	rows, err := gm.Db.Query("CALL tsp_listar_monedas(?, ?)", tokenSesion, incluyeBajas)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var monedas []models.Monedas
	for rows.Next() {
		var m models.Monedas
		err = rows.Scan(&m.IdMoneda, &m.Ledger, &m.IdCuentaEmpresa, &m.Estado, &m.FechaAlta)
		if err != nil {
			return nil, err
		}
		monedas = append(monedas, m)
	}
	return monedas, nil
}

// Borra una moneda únicamente si está en estado pendiente.
// tsp_borrar_moneda
// - tokenSesion: token de sesión del usuario
// - idMoneda: Id de la moneda a borrar
func (gm *GestorMonedas) Borrar(tokenSesion string, idMoneda int) (string, error) {
	var mensaje string
	err := gm.Db.QueryRow("CALL tsp_borrar_moneda(?, ?)", tokenSesion, idMoneda).Scan(&mensaje)
	if err != nil {
		return "", err
	}
	return mensaje, nil
}
