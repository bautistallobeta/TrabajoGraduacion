package models

type ErrorRespuesta struct {
	Error string `json:"Error"`
}

func NewErrorRespuesta(errMsg string) ErrorRespuesta {
	return ErrorRespuesta{Error: errMsg}
}
