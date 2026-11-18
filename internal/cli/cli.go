// Package cli implementa el menú interactivo por consola.
// Recibe io.Reader / io.Writer para poder probarlo sin terminal real.
package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"prjwithdrawal/internal/banco"
)

// Run ejecuta el bucle del menú hasta que el usuario elige salir o termina la entrada.
func Run(in io.Reader, out io.Writer, cajero *banco.Cajero) {
	lector := bufio.NewScanner(in)
	fmt.Fprintf(out, "=== Cajero CLI === Titular: %s\n", cajero.Cuenta().Titular())

	for {
		fmt.Fprintln(out, "\n1) Consultar saldo")
		fmt.Fprintln(out, "2) Retirar efectivo")
		fmt.Fprintln(out, "3) Abonar efectivo")
		fmt.Fprintln(out, "4) Ver log de transacciones (JSON)")
		fmt.Fprintln(out, "0) Salir")
		fmt.Fprint(out, "Opción: ")

		opcion, ok := leerLinea(lector)
		if !ok {
			return
		}

		switch opcion {
		case "1":
			mostrarSaldo(out, cajero.Saldo())
		case "2":
			operar(lector, out, cajero, banco.TipoRetiro)
		case "3":
			operar(lector, out, cajero, banco.TipoAbono)
		case "4":
			verLog(out, cajero)
		case "0":
			fmt.Fprintln(out, "Hasta pronto.")
			return
		default:
			fmt.Fprintln(out, "Opción no válida.")
		}
	}
}

func leerLinea(s *bufio.Scanner) (string, bool) {
	if !s.Scan() {
		return "", false
	}
	return strings.TrimSpace(s.Text()), true
}

func mostrarSaldo(out io.Writer, saldo int64) {
	fmt.Fprintf(out, "Saldo actual: S/ %s\n", banco.FormatoMonto(saldo))
}

// operar atiende retiro y abono con el mismo flujo: leer monto, ejecutar, informar.
func operar(s *bufio.Scanner, out io.Writer, k *banco.Cajero, tipo string) {
	etiqueta, exito := "retirar", "Retiro exitoso"
	if tipo == banco.TipoAbono {
		etiqueta, exito = "abonar", "Abono exitoso"
	}
	fmt.Fprintf(out, "Monto a %s: ", etiqueta)
	texto, ok := leerLinea(s)
	if !ok {
		return
	}
	// Ejecutar interpreta el monto, aplica la operación y la registra en el log.
	tx, err := k.Ejecutar(tipo, texto)
	if err != nil {
		fmt.Fprintf(out, "Error: %v\n", err)
		return
	}
	fmt.Fprintf(out, "%s de S/ %s.\n", exito, tx.Monto)
	mostrarSaldo(out, k.Saldo())
}

func verLog(out io.Writer, k *banco.Cajero) {
	h := k.Historial()
	if len(h) == 0 {
		fmt.Fprintln(out, "Aún no hay transacciones registradas.")
		return
	}
	b, _ := json.MarshalIndent(h, "", "  ")
	fmt.Fprintln(out, string(b))
}
