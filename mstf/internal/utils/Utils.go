package utils

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

// Convierte dos uint64 en un string concatenado.
// El 'low' se rellena con ceros a 20 dígitos para distinguirlo
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
	if len(bi.Bytes()) > 16 {
		return types.Uint128{}, errors.New("Formato de ID inválido: el valor excede el máximo de 128 bits")
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

	// Validar que entra en uint32 (hasta año 2106)
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

// Convierte un timestamp de TigerBeetle a string en hora Argentina (UTC-3).
func TimestampAFecha(timestamp uint64) string {
	art := time.FixedZone("ART", -3*60*60)
	t := time.Unix(0, int64(timestamp)).In(art)
	return t.Format("2006-01-02 15:04:05.999999999")
}

// Convierte un string de fecha a uint64 (nanosegundos desde epoch) para TimestampMin/Max de QueryFilter.
func FechaATimestampNS(s string) (uint64, error) {
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
	return uint64(t.UnixNano()), nil
}

// Valida el formato de una contraseña: mínimo 8 caracteres, al menos una mayúscula y al menos un dígito.
func ValidarFormatoPassword(p string) error {
	if len(p) < 8 {
		return errors.New("la contraseña debe tener al menos 8 caracteres")
	}
	tieneMayuscula := false
	tieneDigito := false
	for _, c := range p {
		if unicode.IsUpper(c) {
			tieneMayuscula = true
		}
		if unicode.IsDigit(c) {
			tieneDigito = true
		}
	}
	if !tieneMayuscula {
		return errors.New("la contraseña debe contener al menos una mayúscula")
	}
	if !tieneDigito {
		return errors.New("la contraseña debe contener al menos un número")
	}
	return nil
}

// Convierte string a hash md5
func MD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

// Convierte un monto decimal a unidad mínima (centavos) multiplicando por 100.
// Trunca en 2 decimales: 15.559 → 1555, 15.5 → 1550, 15 → 1500.
// Usa representación string para evitar errores de aritmética flotante.
func MontoDecimalAUnidadMinima(monto float64) uint64 {
	if monto <= 0 {
		return 0
	}
	s := strconv.FormatFloat(monto, 'f', 4, 64)
	partes := strings.SplitN(s, ".", 2)
	entera := partes[0]
	decimal := ""
	if len(partes) > 1 {
		decimal = partes[1]
	}
	for len(decimal) < 2 {
		decimal += "0"
	}
	decimal = decimal[:2]
	resultado, _ := strconv.ParseUint(entera+decimal, 10, 64)
	return resultado
}

// Funcion p ocultar algunos detalles de infraestructura en errores de red/conexión.
// Para errores de red (ej: MySQL o TigerBeetle inaccesibles), devuelve un mensaje genérico.
// Para otros errores, devuelve el mensaje original.
func SanitizarError(err error) string {
	var netErr *net.OpError
	if errors.As(err, &netErr) {
		return "No se pudo conectar al servicio de datos"
	}
	return err.Error()
}

// Convierte un monto en unidad mínima (centavos) almacenado en TB a string decimal con 2 cifras.
// Ej: 1550 → "15.50"
func Uint128ADecimalMoneda(monto types.Uint128) string {
	n, _ := new(big.Int).SetString(Uint128AStringDecimal(monto), 10)
	cien := big.NewInt(100)
	entero, resto := new(big.Int), new(big.Int)
	entero.DivMod(n, cien, resto)
	return fmt.Sprintf("%s.%02d", entero.String(), resto.Int64())
}
