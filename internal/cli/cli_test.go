package cli

import (
	"bytes"
	"strings"
	"testing"

	"prjwithdrawal/internal/banco"
)

func ejecutar(entrada string, c *banco.Cuenta) string {
	var out, log bytes.Buffer
	k := banco.NuevoCajero(c, banco.NuevoRegistroJSON(&log), nil)
	Run(strings.NewReader(entrada), &out, k)
	return out.String()
}

func TestCLI_RetiroExitoso(t *testing.T) {
	c := banco.NuevaCuenta("Alice", 23456)
	salida := ejecutar("2\n200\n0\n", c)
	if !strings.Contains(salida, "Retiro exitoso de S/ 200.00") {
		t.Errorf("salida inesperada:\n%s", salida)
	}
	if c.Saldo() != 3456 {
		t.Errorf("saldo = %d", c.Saldo())
	}
}

func TestCLI_RetiroRechazadoPorSaldo(t *testing.T) {
	c := banco.NuevaCuenta("Hamza", 19876)
	salida := ejecutar("2\n200\n0\n", c)
	if !strings.Contains(salida, "saldo insuficiente") {
		t.Errorf("salida inesperada:\n%s", salida)
	}
	if c.Saldo() != 19876 {
		t.Errorf("el saldo no debía cambiar: %d", c.Saldo())
	}
}

func TestCLI_MontoNoNumerico(t *testing.T) {
	c := banco.NuevaCuenta("Alice", 23456)
	salida := ejecutar("2\nabc\n0\n", c)
	if !strings.Contains(salida, "monto inválido") {
		t.Errorf("salida inesperada:\n%s", salida)
	}
}

func TestCLI_AbonoExitoso(t *testing.T) {
	c := banco.NuevaCuenta("Alice", 23456)
	salida := ejecutar("3\n100\n0\n", c)
	if !strings.Contains(salida, "Abono exitoso de S/ 100.00") {
		t.Errorf("salida inesperada:\n%s", salida)
	}
	if c.Saldo() != 33456 {
		t.Errorf("saldo = %d", c.Saldo())
	}
}

func TestCLI_AbonoRechazadoMontoNegativo(t *testing.T) {
	c := banco.NuevaCuenta("Hamza", 19876)
	salida := ejecutar("3\n-20\n0\n", c)
	if !strings.Contains(salida, "monto inválido") {
		t.Errorf("salida inesperada:\n%s", salida)
	}
	if c.Saldo() != 19876 {
		t.Errorf("el saldo no debía cambiar: %d", c.Saldo())
	}
}

func TestCLI_VerLogMuestraJSONConResultados(t *testing.T) {
	c := banco.NuevaCuenta("Alice", 23456)
	salida := ejecutar("2\n200\n2\n500\n2\nabc\n4\n0\n", c)
	for _, want := range []string{`"resultado": "exito"`, `"resultado": "fallo"`, `"resultado": "error"`} {
		if !strings.Contains(salida, want) {
			t.Errorf("falta %s en la salida:\n%s", want, salida)
		}
	}
}

func TestCLI_LogVacio(t *testing.T) {
	c := banco.NuevaCuenta("Alice", 23456)
	if !strings.Contains(ejecutar("4\n0\n", c), "Aún no hay transacciones") {
		t.Errorf("se esperaba mensaje de log vacío")
	}
}

func TestCLI_OpcionInvalidaYFinDeEntrada(t *testing.T) {
	c := banco.NuevaCuenta("Alice", 23456)
	salida := ejecutar("9\n", c) // sin "0": el bucle debe terminar al acabar la entrada
	if !strings.Contains(salida, "Opción no válida") {
		t.Errorf("salida inesperada:\n%s", salida)
	}
}
