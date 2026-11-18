package banco

import "testing"

// Cada test implementa un Escenario de features/retiro-cash.feature
// siguiendo la estructura Given / When / Then.

func TestEscenario_RetiroExitosoDentroDelSaldo(t *testing.T) {
	// Given Alice tiene 234.56 en su cuenta
	alice := NuevaCuenta("Alice", 23456)
	// When Alice intenta retirar 200.00
	err := alice.Retirar(20000)
	// Then el retiro es exitoso
	if err != nil {
		t.Fatalf("se esperaba retiro exitoso, error: %v", err)
	}
	// And el saldo de Alice es 34.56
	if alice.Saldo() != 3456 {
		t.Errorf("saldo = %s, se esperaba 34.56", FormatoMonto(alice.Saldo()))
	}
}

func TestEscenario_RetiroRechazadoPorExcederSaldo(t *testing.T) {
	// Given Hamza tiene 198.76 en su cuenta
	hamza := NuevaCuenta("Hamza", 19876)
	// When Hamza intenta retirar 200.00
	err := hamza.Retirar(20000)
	// Then el retiro es rechazado
	if err != ErrSaldoInsuficiente {
		t.Fatalf("se esperaba ErrSaldoInsuficiente, se obtuvo: %v", err)
	}
	// And el saldo de Hamza sigue siendo 198.76 (RN-03 atomicidad)
	if hamza.Saldo() != 19876 {
		t.Errorf("saldo = %s, se esperaba 198.76", FormatoMonto(hamza.Saldo()))
	}
}

func TestEscenario_RetiroRechazadoPorMontoInvalido(t *testing.T) {
	// Given Alice tiene 234.56 en su cuenta
	alice := NuevaCuenta("Alice", 23456)
	// When Alice intenta retirar -50.00
	err := alice.Retirar(-5000)
	// Then el retiro es rechazado
	if err != ErrMontoInvalido {
		t.Fatalf("se esperaba ErrMontoInvalido, se obtuvo: %v", err)
	}
	// And el saldo de Alice sigue siendo 234.56
	if alice.Saldo() != 23456 {
		t.Errorf("saldo = %s, se esperaba 234.56", FormatoMonto(alice.Saldo()))
	}
}

// ---- Iteración 2: US-01b Abono en efectivo (features/abono-cash.feature) ----

func TestEscenario_AbonoExitoso(t *testing.T) {
	// Given Alice tiene 234.56 en su cuenta
	alice := NuevaCuenta("Alice", 23456)
	// When Alice abona 100.00
	err := alice.Abonar(10000)
	// Then el abono es exitoso
	if err != nil {
		t.Fatalf("se esperaba abono exitoso, error: %v", err)
	}
	// And el saldo de Alice es 334.56
	if alice.Saldo() != 33456 {
		t.Errorf("saldo = %s, se esperaba 334.56", FormatoMonto(alice.Saldo()))
	}
}

func TestEscenario_AbonoRechazadoPorMontoCero(t *testing.T) {
	alice := NuevaCuenta("Alice", 23456)
	if err := alice.Abonar(0); err != ErrMontoInvalido {
		t.Fatalf("se esperaba ErrMontoInvalido, se obtuvo: %v", err)
	}
	if alice.Saldo() != 23456 {
		t.Errorf("saldo = %s, se esperaba 234.56", FormatoMonto(alice.Saldo()))
	}
}

func TestEscenario_AbonoRechazadoPorMontoNegativo(t *testing.T) {
	hamza := NuevaCuenta("Hamza", 19876)
	if err := hamza.Abonar(-2000); err != ErrMontoInvalido {
		t.Fatalf("se esperaba ErrMontoInvalido, se obtuvo: %v", err)
	}
	if hamza.Saldo() != 19876 {
		t.Errorf("saldo = %s, se esperaba 198.76", FormatoMonto(hamza.Saldo()))
	}
}

func TestEscenario_AbonoSeguidoDeRetiro(t *testing.T) {
	// Given Hamza tiene 198.76; When abona 50.00 y retira 200.00
	hamza := NuevaCuenta("Hamza", 19876)
	if err := hamza.Abonar(5000); err != nil {
		t.Fatalf("abono: %v", err)
	}
	// Then el retiro es exitoso (antes del abono habría sido rechazado)
	if err := hamza.Retirar(20000); err != nil {
		t.Fatalf("retiro: %v", err)
	}
	// And el saldo de Hamza es 48.76
	if hamza.Saldo() != 4876 {
		t.Errorf("saldo = %s, se esperaba 48.76", FormatoMonto(hamza.Saldo()))
	}
}

func TestRN02_RetiroDelSaldoExactoDejaCero(t *testing.T) {
	c := NuevaCuenta("Alice", 10000)
	if err := c.Retirar(10000); err != nil {
		t.Fatalf("retirar el saldo exacto debe ser válido: %v", err)
	}
	if c.Saldo() != 0 {
		t.Errorf("saldo = %d, se esperaba 0", c.Saldo())
	}
}

func TestRetiroMontoCeroEsInvalido(t *testing.T) {
	c := NuevaCuenta("Alice", 10000)
	if err := c.Retirar(0); err != ErrMontoInvalido {
		t.Errorf("se esperaba ErrMontoInvalido, se obtuvo: %v", err)
	}
}

func TestParseMonto(t *testing.T) {
	casos := []struct {
		entrada string
		want    int64
		ok      bool
	}{
		{"200", 20000, true},
		{"200.00", 20000, true},
		{"200.5", 20050, true},
		{"234.56", 23456, true},
		{"0,99", 99, true},
		{"-50.00", -5000, true},
		{".5", 50, true},
		{"", 0, false},
		{"abc", 0, false},
		{"1.234", 0, false},
		{"1.2.3", 0, false},
		{".", 0, false},
	}
	for _, c := range casos {
		got, err := ParseMonto(c.entrada)
		if c.ok && (err != nil || got != c.want) {
			t.Errorf("ParseMonto(%q) = %d, %v; se esperaba %d", c.entrada, got, err, c.want)
		}
		if !c.ok && err == nil {
			t.Errorf("ParseMonto(%q) debía fallar", c.entrada)
		}
	}
}

func TestFormatoMonto(t *testing.T) {
	if got := FormatoMonto(3456); got != "34.56" {
		t.Errorf("got %s", got)
	}
	if got := FormatoMonto(5); got != "0.05" {
		t.Errorf("got %s", got)
	}
}
