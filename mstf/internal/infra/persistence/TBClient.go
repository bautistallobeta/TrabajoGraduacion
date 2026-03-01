package persistence

import (
	"MSTransaccionesFinancieras/internal/config"
	"log"

	tigerbeetle "github.com/tigerbeetle/tigerbeetle-go"
	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

// instancia global de cliente de TB
var ClienteTB tigerbeetle.Client

func InitTBClient(cfg config.Config) error {
	var err error

	ClienteTB, err = tigerbeetle.NewClient(types.ToUint128(0), cfg.DireccionesTigerBeetle)
	if err != nil {
		log.Printf("Error al crear cliente de TigerBeetle: %v", err)
		return err
	}
	return nil
}

func CloseTBClient() {
	if ClienteTB != nil {
		ClienteTB.Close()
	}
}
