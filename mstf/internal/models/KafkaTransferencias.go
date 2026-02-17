package models

// Mensaje JSON que se espera en el topic de kafka
type KafkaTransferencias struct {
	IdTransferencia string `json:"IdTransferencia"`
	IdCuentaDebito  string `json:"IdCuentaDebito"`
	IdCuentaCredito string `json:"IdCuentaCredito"`
	IdUsuarioFinal  string `json:"IdUsuarioFinal"`
	Monto           uint64 `json:"Monto"`
	IdMoneda        uint32 `json:"IdMoneda"`
	IdCategoria     uint64 `json:"IdCategoria"`
	Fecha           string `json:"Fecha"`
}
