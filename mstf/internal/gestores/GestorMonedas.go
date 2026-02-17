package gestores

import (
	"MSTransaccionesFinancieras/internal/infra/persistence"
	"MSTransaccionesFinancieras/internal/models"
	"errors"
	"strconv"
)

type GestorMonedas struct {
}

func NewGestorMonedas() *GestorMonedas {
	return &GestorMonedas{}
}

// Crea una moneda en estado P: Pendiente.
// tsp_crear_moneda
// - tokenSesion: token de sesión del usuario
// - idMoneda: Id de la moneda a crear (viene de MisGastos)
// - idCuentaEmpresa: Id de la cuenta empresa en TB asociada a esta moneda
func (gm *GestorMonedas) Crear(tokenSesion string, idMoneda int, idCuentaEmpresa string) (string, error) {
	var mensaje string
	err := persistence.ClienteMySQL.QueryRow("CALL tsp_crear_moneda(?, ?, ?)", tokenSesion, idMoneda, idCuentaEmpresa).Scan(&mensaje)
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
	rows, err := persistence.ClienteMySQL.Query("CALL tsp_listar_monedas(?, ?)", tokenSesion, incluyeBajas)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var monedas []models.Monedas
	for rows.Next() {
		var m models.Monedas
		err = rows.Scan(&m.IdMoneda, &m.IdCuentaEmpresa, &m.Estado, &m.FechaAlta)
		if err != nil {
			return nil, err
		}
		monedas = append(monedas, m)
	}
	return monedas, nil
}

// Borra una moneda únicamente si está en estado Inactivo.
// tsp_borrar_moneda
// - tokenSesion: token de sesión del usuario
// - idMoneda: Id de la moneda a borrar
func (gm *GestorMonedas) Borrar(tokenSesion string, idMoneda int) (string, error) {
	// BORRAR TEST
	if idMoneda == 901 {
		return "", errors.New("error simulado: MySQL caido en rollback")
	}
	var mensaje string
	err := persistence.ClienteMySQL.QueryRow("CALL tsp_borrar_moneda(?, ?)", tokenSesion, idMoneda).Scan(&mensaje)
	if err != nil {
		return "", err
	}
	models.CacheMonedas.Borrar(strconv.Itoa(idMoneda))
	return mensaje, nil
}
