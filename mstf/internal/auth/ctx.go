package auth

import "context"

// tipo privado para evitar colisiones con otras claves de context.
type claveCtx string

const (
	ClaveCredencial claveCtx = "Credencial"
	ClaveActor      claveCtx = "Actor"
)

// Eextrae la credencial y el actor del contexto de la request
// Strings vacíos si los valores no están presentes
func CredencialDesdeCtx(ctx context.Context) (credencial string, actor string) {
	credencial, _ = ctx.Value(ClaveCredencial).(string)
	actor, _ = ctx.Value(ClaveActor).(string)
	return
}
