package models

// Mensaje JSON que se espera en el topic de kafka
type KafkaTransferencias struct {
	IdTransferencia string `json:"IdTransferencia"`
	IdCuentaDebito  string `json:"IdCuentaDebito"`
	IdCuentaCredito string `json:"IdCuentaCredito"`
	Monto           uint64 `json:"Monto"`
	Ledger          uint32 `json:"Ledger"`
	IdCategoria     uint64 `json:"IdCategoria"`
	Fecha           string `json:"Fecha"`
}
