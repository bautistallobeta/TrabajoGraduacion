package models

import (
	"MSTransaccionesFinancieras/internal/auth"
	"MSTransaccionesFinancieras/internal/infra/cache"
	"MSTransaccionesFinancieras/internal/infra/persistence"
	"context"
	"database/sql"
	"time"
)

type Parametros struct {
	Parametro     string `json:"Parametro"`
	Valor         string `json:"Valor"`
	Descripcion   string `json:"Descripcion"`
	EsModificable string `json:"EsModificable"`
}

var CacheParametros = cache.NewCache[Parametros](5 * time.Minute)

// Instancia un parámetro específico por su clave.
// tsp_dame_parametro
func (p *Parametros) Dame() (string, error) {
	if cached, ok := CacheParametros.Dame(p.Parametro); ok {
		*p = cached
		return "OK", nil
	}

	rows, err := persistence.ClienteMySQL.Query("CALL tsp_dame_parametro(?)", p.Parametro)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	var mensaje string
	var param sql.NullString
	var valor sql.NullString
	var descripcion sql.NullString
	var esModificable sql.NullString
	if rows.Next() {
		err = rows.Scan(&mensaje, &param, &valor, &descripcion, &esModificable)
		if err != nil {
			return mensaje, err
		}
		if param.Valid {
			p.Parametro = param.String
		} else {
			p.Parametro = ""
		}
		if valor.Valid {
			p.Valor = valor.String
		} else {
			p.Valor = ""
		}
		if descripcion.Valid {
			p.Descripcion = descripcion.String
		} else {
			p.Descripcion = ""
		}
		if esModificable.Valid {
			p.EsModificable = esModificable.String
		} else {
			p.EsModificable = ""
		}
		CacheParametros.Guardar(p.Parametro, *p)
		return mensaje, nil
	}
	return mensaje, nil
}

// Permite buscar los parámetros del sistema según su nombre.
// tsp_buscar_parametros
// - Cadena: texto a buscar dentro del nombre del parámetro (puede ser parte del nombre o el nombre completo)
func (p *Parametros) BuscarParametros(Cadena string) ([]Parametros, error) {
	rows, err := persistence.ClienteMySQL.Query("CALL tsp_buscar_parametros(?, ?)", Cadena, "S")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parametros []Parametros
	for rows.Next() {
		var p Parametros
		err = rows.Scan(&p.Parametro, &p.Valor, &p.Descripcion, &p.EsModificable)
		if err != nil {
			return nil, err
		}
		parametros = append(parametros, p)
	}
	return parametros, nil
}

// Permite modificar el valor de un parámetro siempre y cuando exista y sea modificable.
// tsp_modificar_parametro
// - valor: nuevo valor del parámetro
func (p *Parametros) ModificarParametro(ctx context.Context, Valor string) (string, error) {
	credencial, actor := auth.CredencialDesdeCtx(ctx)
	var mensaje string
	err := persistence.ClienteMySQL.QueryRow("CALL tsp_modificar_parametro(?, ?, ?, ?)", credencial, actor, p.Parametro, Valor).Scan(&mensaje)
	if err != nil {
		return "", err
	}
	CacheParametros.Borrar(p.Parametro)
	return mensaje, nil
}
