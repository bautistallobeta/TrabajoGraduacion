package models

import (
	"MSTransaccionesFinancieras/internal/infra/persistence"
	"database/sql"
	"errors"
)

type Parametros struct {
	Parametro     string `json:"parametro"`
	Valor         string `json:"valor"`
	Descripcion   string `json:"descripcion"`
	EsModificable string `json:"es_modificable"`
}

// Devuelve los datos de un parámetro específico por su clave.
// tsp_dame_parametro
// - tokenSesion: token de sesión del usuario
// - parametro: clave del parámetro a instanciar
func (p *Parametros) Dame(tokenSesion string) (string, error) {
	rows, err := persistence.ClienteMySQL.Query("CALL tsp_dame_parametro(?, ?)", tokenSesion, p.Parametro)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	var mensaje string
	var valor sql.NullString
	var descripcion sql.NullString
	var esModificable sql.NullString
	if rows.Next() {
		err = rows.Scan(&mensaje, &p.Parametro, &valor, &descripcion, &esModificable)
		if err != nil {
			return mensaje, err
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
		return mensaje, nil
	}
	if mensaje != "OK" {
		return mensaje, errors.New(mensaje)
	}
	return mensaje, nil
}

//	Permite buscar los parámetros del sistema según su nombre. Si pSoloModificables es 'S', muestra solo los
//	modificables desde el sitio administrativo. Ordena por nombre de parámetro.
//
// tsp_buscar_parametros
// - tokenSesion: token de sesión del usuario
func (p *Parametros) BuscarParametros(tokenSesion string, cadena string) ([]Parametros, error) {
	rows, err := persistence.ClienteMySQL.Query("CALL tsp_buscar_parametros(?, ?)", tokenSesion, cadena)
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
// Devuelve OK o el mensaje de error en Mensaje.
// tsp_modificar_parametro
// - tokenSesion: token de sesión del usuario
// - parametro: clave del parámetro a modificar
// - valor: nuevo valor del parámetro
func (p *Parametros) ModificarParametro(tokenSesion string) (string, error) {
	var mensaje string
	err := persistence.ClienteMySQL.QueryRow("CALL tsp_modificar_parametro(?, ?, ?)", tokenSesion, p.Parametro, p.Valor).Scan(&mensaje)
	if err != nil {
		return "", err
	}
	return mensaje, nil
}
