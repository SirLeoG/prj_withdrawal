// Package datos expone los datos iniciales (formato JSON incrustado con go:embed).
package datos

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"prjwithdrawal/internal/banco"
)

//go:embed cuentas.json
var cuentasJSON []byte

type cuentaDato struct {
	Titular string `json:"titular"`
	Saldo   string `json:"saldo"`
}

// Titulares devuelve los nombres de las cuentas disponibles.
func Titulares() ([]string, error) {
	var ds []cuentaDato
	if err := json.Unmarshal(cuentasJSON, &ds); err != nil {
		return nil, err
	}
	var out []string
	for _, d := range ds {
		out = append(out, d.Titular)
	}
	return out, nil
}

// BuscarCuenta crea la cuenta del titular a partir del JSON incrustado.
func BuscarCuenta(titular string) (*banco.Cuenta, error) {
	var ds []cuentaDato
	if err := json.Unmarshal(cuentasJSON, &ds); err != nil {
		return nil, err
	}
	for _, d := range ds {
		if d.Titular == titular {
			saldo, err := banco.ParseMonto(d.Saldo)
			if err != nil {
				return nil, err
			}
			return banco.NuevaCuenta(d.Titular, saldo), nil
		}
	}
	return nil, fmt.Errorf("titular %q no encontrado en cuentas.json", titular)
}
