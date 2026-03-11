package models

type ErrorRespuesta struct {
	Error string `json:"error"`
}

func NewErrorRespuesta(errMsg string) ErrorRespuesta {
	return ErrorRespuesta{Error: errMsg}
}
