package models

// Mensaje JSON que se espera en el topic de kafka
type KafkaTransferencias struct {
	IdTransferencia string `json:"IdTransferencia"`
	IdUsuarioFinal  uint64 `json:"IdUsuarioFinal"`
	Monto           uint64 `json:"Monto"`
	IdMoneda        uint32 `json:"IdMoneda"`
	Tipo            string `json:"Tipo"`
	IdCategoria     uint64 `json:"IdCategoria"`
	Fecha           string `json:"Fecha"`
}
