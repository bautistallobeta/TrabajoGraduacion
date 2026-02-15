package utils

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

// Convierte dos uint64 en un string concatenado.
// El 'low' se rellena con ceros a 20 dígitos para distinguirlo
// TODO: EVALUAR REESCRIBIR SIN PADDING
func ConcatenarIDString(high uint64, low uint64) string {
	lowString := fmt.Sprintf("%020d", low)
	highString := fmt.Sprintf("%d", high)
	return highString + lowString
}

// Convierte una cadena de texto a un types.Uint128, aceptando solo la representación decimal
func ParsearUint128(s string) (types.Uint128, error) {
	ss := strings.TrimSpace(s)

	bi, ok := new(big.Int).SetString(ss, 10)
	if !ok || bi.Sign() < 0 {
		return types.Uint128{}, errors.New("Formato de ID inválido. Solo se acepta la representación decimal")
	}
	return types.BigIntToUint128(*bi), nil
}

// Convierte un types.Uint128 a su representación en string decimal
func Uint128AStringDecimal(id types.Uint128) string {
	byteSlice := id[:]

	// Invertir el orden de bytes (LE de TB a BE para big.Int)
	reversedBytes := make([]byte, len(byteSlice))
	for i, j := 0, len(byteSlice)-1; i < len(byteSlice); i, j = i+1, j-1 {
		reversedBytes[i] = byteSlice[j]
	}
	bigInt := new(big.Int).SetBytes(reversedBytes)

	return bigInt.String()
}

func FechaAUserData128(s string) (types.Uint128, error) {
	//TODO: revisar formateo enedge cases
	layouts := []string{
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05.999999",
		"2006-01-02 15:04:05.999",
		"2006-01-02 15:04:05",
		time.RFC3339Nano,
		time.RFC3339,
	}

	var t time.Time
	var err error
	for _, l := range layouts {
		t, err = time.ParseInLocation(l, s, time.UTC)
		if err == nil {
			break
		}
	}
	if err != nil {
		return types.Uint128{}, errors.New("Formato de fecha inválido; esperado 'YYYY-MM-DD HH:MM:SS[.fracción]' o RFC3339")
	}
	// nanosegundos a uint64
	ns := uint64(t.UnixNano())

	var b [16]byte
	binary.LittleEndian.PutUint64(b[0:8], ns)
	u := types.BytesToUint128(b)
	return u, nil
}

// Convierte un types.Uint128 de vuelta a string
func UserData128AFecha(u types.Uint128) (string, error) {
	b := u.Bytes()
	lo := binary.LittleEndian.Uint64(b[0:8])
	t := time.Unix(0, int64(lo)).UTC()
	return t.Format("2006-01-02 15:04:05.999999999"), nil
}

// Convierte un string de fecha a uint32 (segundos desde epoch)
func FechaAUserData32(s string) (uint32, error) {
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		time.RFC3339,
		"2006-01-02",
	}
	var t time.Time
	var err error
	for _, l := range layouts {
		t, err = time.ParseInLocation(l, s, time.UTC)
		if err == nil {
			break
		}
	}
	if err != nil {
		return 0, errors.New("Formato de fecha inválido; esperado 'YYYY-MM-DD HH:MM:SS' o similar")
	}
	segundos := t.Unix()

	// Validar que entra en uint32 (hasta año 2106) - TODO: revisar si conviene cambiar esto
	if segundos < 0 || segundos > 4294967295 {
		return 0, errors.New("Fecha fuera de rango válido para uint32 (1970-2106)")
	}

	return uint32(segundos), nil
}

// Convierte un uint32  de vuelta a string
func UserData32AFecha(u uint32) (string, error) {
	t := time.Unix(int64(u), 0).UTC()
	return t.Format("2006-01-02 15:04:05"), nil
}

// Convierte un timestamp de TigerBeetle  a string
func TimestampAFecha(timestamp uint64) string {
	t := time.Unix(0, int64(timestamp)).UTC()
	return t.Format("2006-01-02 15:04:05.999999999")
}

// Convierte string a hash md5
func MD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
