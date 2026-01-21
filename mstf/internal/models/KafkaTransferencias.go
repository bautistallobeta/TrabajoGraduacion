package models

// Mensaje JSON que se espera en el topic de kafka
type KafkaTransferencias struct {
	IdTransferencia string `json:"id_transferencia"`
	IdCuentaDebito  string `json:"id_cuenta_debito"`
	IdCuentaCredito string `json:"id_cuenta_credito"`
	Monto           uint64 `json:"monto"`
	Ledger          uint32 `json:"ledger"`
	IdCategoria     uint64 `json:"id_categoria"`
	Fecha           string `json:"fecha"`
}
