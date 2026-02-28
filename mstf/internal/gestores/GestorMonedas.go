package gestores

import (
	"MSTransaccionesFinancieras/internal/auth"
	"MSTransaccionesFinancieras/internal/infra/persistence"
	"MSTransaccionesFinancieras/internal/models"
	"context"
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
// - Moneda.IdMoneda: Id de la moneda a crear (viene de MisGastos)
// - Moneda.IdCuentaEmpresa: Id de la cuenta empresa en TB asociada a esta moneda
func (gm *GestorMonedas) Crear(ctx context.Context, Moneda models.Monedas) (string, error) {
	credencial, actor := auth.CredencialDesdeCtx(ctx)
	var mensaje string
	err := persistence.ClienteMySQL.QueryRow("CALL tsp_crear_moneda(?, ?, ?, ?)", credencial, actor, Moneda.IdMoneda, Moneda.IdCuentaEmpresa).Scan(&mensaje)
	if err != nil {
		return "", err
	}
	return mensaje, nil
}

// Permite listar todas las monedas.
// tsp_listar_monedas
// - IncluyeInactivos: 'S' o 'N' para incluir o no las monedas inactivas
func (gm *GestorMonedas) Listar(IncluyeInactivos string) ([]models.Monedas, error) {
	rows, err := persistence.ClienteMySQL.Query("CALL tsp_listar_monedas(?)", IncluyeInactivos)
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
// - Moneda.IdMoneda: Id de la moneda a borrar
func (gm *GestorMonedas) Borrar(ctx context.Context, Moneda models.Monedas) (string, error) {
	// BORRAR TEST
	if Moneda.IdMoneda == 901 {
		return "", errors.New("error simulado: MySQL caido en rollback")
	}
	credencial, actor := auth.CredencialDesdeCtx(ctx)
	var mensaje string
	err := persistence.ClienteMySQL.QueryRow("CALL tsp_borrar_moneda(?, ?, ?)", credencial, actor, Moneda.IdMoneda).Scan(&mensaje)
	if err != nil {
		return "", err
	}
	models.CacheMonedas.Borrar(strconv.Itoa(Moneda.IdMoneda))
	return mensaje, nil
}
