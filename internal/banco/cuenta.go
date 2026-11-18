// Package banco contiene la entidad de negocio descrita en los Feature Gherkin.
//
// Los montos se manejan en centavos (int64) para evitar errores de redondeo
// de punto flotante (p. ej. 234.56 - 200.00 debe ser exactamente 34.56).
package banco

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	// ErrMontoInvalido: el monto es cero, negativo o no es un número válido (flujo 4.b).
	ErrMontoInvalido = errors.New("monto inválido: debe ser mayor a 0")
	// ErrSaldoInsuficiente: RN-01 y RN-02 (flujo 4.a).
	ErrSaldoInsuficiente = errors.New("saldo insuficiente: intente con un monto menor")
)

// Cuenta representa la cuenta bancaria del cliente.
type Cuenta struct {
	titular string
	saldo   int64 // centavos
}

// NuevaCuenta crea una cuenta con saldo inicial en centavos.
func NuevaCuenta(titular string, saldoCentavos int64) *Cuenta {
	return &Cuenta{titular: titular, saldo: saldoCentavos}
}

func (c *Cuenta) Titular() string { return c.titular }

// Saldo devuelve el saldo actual en centavos.
func (c *Cuenta) Saldo() int64 { return c.saldo }

// Retirar descuenta el monto del saldo.
//   - RN-01: no se retira si el monto es mayor al saldo.
//   - RN-02: no se permiten saldos negativos.
//   - RN-03: operación atómica: si falla, el saldo no cambia.
func (c *Cuenta) Retirar(montoCentavos int64) error {
	if montoCentavos <= 0 {
		return ErrMontoInvalido
	}
	if montoCentavos > c.saldo {
		return ErrSaldoInsuficiente
	}
	c.saldo -= montoCentavos
	return nil
}

// Abonar incrementa el saldo con el monto depositado (US-01b).
// Solo acepta montos mayores a cero; si falla, el saldo no cambia.
func (c *Cuenta) Abonar(montoCentavos int64) error {
	if montoCentavos <= 0 {
		return ErrMontoInvalido
	}
	c.saldo += montoCentavos
	return nil
}

// ParseMonto convierte texto como "200", "200.5" o "200.00" a centavos.
func ParseMonto(s string) (int64, error) {
	s = strings.TrimSpace(strings.ReplaceAll(s, ",", "."))
	if s == "" {
		return 0, ErrMontoInvalido
	}
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")
	partes := strings.Split(s, ".")
	if len(partes) > 2 || partes[0] == "" && len(partes) == 1 {
		return 0, ErrMontoInvalido
	}
	enteros := partes[0]
	if enteros == "" {
		enteros = "0"
	}
	frac := "00"
	if len(partes) == 2 {
		frac = partes[1]
		if len(frac) == 0 || len(frac) > 2 {
			return 0, ErrMontoInvalido
		}
		if len(frac) == 1 {
			frac += "0"
		}
	}
	e, err := strconv.ParseInt(enteros, 10, 64)
	if err != nil || e < 0 {
		return 0, ErrMontoInvalido
	}
	f, err := strconv.ParseInt(frac, 10, 64)
	if err != nil || f < 0 {
		return 0, ErrMontoInvalido
	}
	total := e*100 + f
	if neg {
		total = -total
	}
	return total, nil
}

// FormatoMonto devuelve centavos como "234.56".
func FormatoMonto(centavos int64) string {
	signo := ""
	if centavos < 0 {
		signo = "-"
		centavos = -centavos
	}
	return fmt.Sprintf("%s%d.%02d", signo, centavos/100, centavos%100)
}
