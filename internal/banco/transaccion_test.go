package banco

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

// Pruebas de features/transaccion-log.feature (US-04).

type escrituraFallida struct{}

func (escrituraFallida) Write([]byte) (int, error) { return 0, errors.New("disco lleno") }

func relojFijo() time.Time { return time.Date(2026, 11, 10, 15, 0, 0, 0, time.UTC) }

func nuevoCajero(titular string, saldo int64) (*Cajero, *bytes.Buffer) {
	var buf bytes.Buffer
	return NuevoCajero(NuevaCuenta(titular, saldo), NuevoRegistroJSON(&buf), relojFijo), &buf
}

func ultima(t *testing.T, buf *bytes.Buffer) Transaccion {
	t.Helper()
	lineas := strings.Split(strings.TrimSpace(buf.String()), "\n")
	var tx Transaccion
	if err := json.Unmarshal([]byte(lineas[len(lineas)-1]), &tx); err != nil {
		t.Fatalf("la línea del log no es JSON válido: %v", err)
	}
	return tx
}

func TestEscenario_RetiroExitosoQuedaRegistradoComoExito(t *testing.T) {
	// Given Alice tiene 234.56
	k, buf := nuevoCajero("Alice", 23456)
	// When Alice intenta retirar 200.00
	if _, err := k.Retirar(20000); err != nil {
		t.Fatalf("retiro: %v", err)
	}
	// Then el log registra un "retiro" con resultado "exito"
	tx := ultima(t, buf)
	if tx.Tipo != TipoRetiro || tx.Resultado != ResultadoExito {
		t.Errorf("registro = %+v", tx)
	}
	// And saldo_antes 234.56 y saldo_despues 34.56
	if tx.SaldoAntes != "234.56" || tx.SaldoDespues != "34.56" {
		t.Errorf("saldos = %s / %s", tx.SaldoAntes, tx.SaldoDespues)
	}
}

func TestEscenario_RetiroPorEncimaDelSaldoQuedaRegistradoComoFallo(t *testing.T) {
	k, buf := nuevoCajero("Hamza", 19876)
	_, err := k.Retirar(20000)
	if err != ErrSaldoInsuficiente {
		t.Fatalf("se esperaba ErrSaldoInsuficiente, se obtuvo %v", err)
	}
	tx := ultima(t, buf)
	if tx.Resultado != ResultadoFallo || tx.Detalle == "" {
		t.Errorf("registro = %+v", tx)
	}
	if k.Cuenta().Saldo() != 19876 {
		t.Errorf("el saldo no debía cambiar")
	}
}

func TestEscenario_AbonoExitosoQuedaRegistradoComoExito(t *testing.T) {
	k, buf := nuevoCajero("Alice", 23456)
	if _, err := k.Abonar(10000); err != nil {
		t.Fatalf("abono: %v", err)
	}
	tx := ultima(t, buf)
	if tx.Tipo != TipoAbono || tx.Resultado != ResultadoExito || tx.SaldoDespues != "334.56" {
		t.Errorf("registro = %+v", tx)
	}
}

func TestEscenario_MontoNoNumericoQuedaRegistradoComoError(t *testing.T) {
	k, buf := nuevoCajero("Alice", 23456)
	_, errParse := ParseMonto("abc")
	if err := k.RegistrarError(TipoRetiro, "abc", errParse); err != nil {
		t.Fatalf("registrar error: %v", err)
	}
	tx := ultima(t, buf)
	if tx.Tipo != TipoRetiro || tx.Resultado != ResultadoError || tx.Monto != "abc" {
		t.Errorf("registro = %+v", tx)
	}
	if k.Cuenta().Saldo() != 23456 {
		t.Errorf("el saldo no debía cambiar")
	}
}

func TestEscenario_SiElLogFallaLaOperacionSeRevierte(t *testing.T) {
	// Given Alice tiene 234.56 y el log no puede escribirse
	k := NuevoCajero(NuevaCuenta("Alice", 23456), NuevoRegistroJSON(escrituraFallida{}), relojFijo)
	// When Alice intenta retirar 200.00
	_, err := k.Retirar(20000)
	// Then la operación termina con error
	if !errors.Is(err, ErrRegistro) {
		t.Fatalf("se esperaba ErrRegistro, se obtuvo %v", err)
	}
	// And el saldo sigue siendo 234.56 (atomicidad, RN-03)
	if k.Cuenta().Saldo() != 23456 {
		t.Errorf("saldo = %s, se esperaba 234.56", FormatoMonto(k.Cuenta().Saldo()))
	}
	if len(k.Historial()) != 0 {
		t.Errorf("no debe quedar historial de una transacción no registrada")
	}
}

func TestEscenario_CadaLineaDelLogEsJSONValidoConIdConsecutivo(t *testing.T) {
	k, buf := nuevoCajero("Alice", 23456)
	k.Abonar(5000)
	k.Retirar(1000)
	k.Retirar(99999999) // fallo, también se registra
	lineas := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lineas) != 3 {
		t.Fatalf("líneas = %d, se esperaban 3", len(lineas))
	}
	for i, l := range lineas {
		var tx Transaccion
		if err := json.Unmarshal([]byte(l), &tx); err != nil {
			t.Fatalf("línea %d no es JSON válido: %v", i+1, err)
		}
		if tx.ID != i+1 {
			t.Errorf("id = %d, se esperaba %d", tx.ID, i+1)
		}
		if tx.Fecha != "2026-11-10T15:00:00Z" {
			t.Errorf("fecha = %s", tx.Fecha)
		}
	}
}

// ---- Iteración 4: Ejecutar (punto de entrada común CLI/Web) y concurrencia ----

func TestEjecutar_TextoIlegibleSeRegistraComoError(t *testing.T) {
	k, buf := nuevoCajero("Alice", 23456)
	tx, err := k.Ejecutar(TipoRetiro, "abc")
	if err == nil || tx.Resultado != ResultadoError || tx.ID != 1 {
		t.Fatalf("tx=%+v err=%v", tx, err)
	}
	if ultima(t, buf).Resultado != ResultadoError || k.Saldo() != 23456 {
		t.Errorf("el error debe quedar en el log y el saldo intacto")
	}
}

func TestEjecutar_TipoInvalido(t *testing.T) {
	k, buf := nuevoCajero("Alice", 23456)
	if _, err := k.Ejecutar("transferir", "10"); err != ErrTipoInvalido {
		t.Errorf("err = %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("un tipo inválido no debe escribir en el log")
	}
}

func TestEjecutar_ConcurrenciaMantieneSaldoYLogConsistentes(t *testing.T) {
	k, buf := nuevoCajero("Alice", 0)
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); k.Ejecutar(TipoAbono, "1.00") }()
	}
	wg.Wait()
	if k.Saldo() != 5000 {
		t.Errorf("saldo = %s, se esperaba 50.00", FormatoMonto(k.Saldo()))
	}
	if n := len(strings.Split(strings.TrimSpace(buf.String()), "\n")); n != 50 {
		t.Errorf("líneas de log = %d, se esperaban 50", n)
	}
}
