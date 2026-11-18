package datos

import "testing"

func TestBuscarCuenta(t *testing.T) {
	c, err := BuscarCuenta("Alice")
	if err != nil || c.Saldo() != 23456 {
		t.Fatalf("Alice: %v saldo=%v", err, c)
	}
	if _, err := BuscarCuenta("Nadie"); err == nil {
		t.Error("debía fallar con un titular inexistente")
	}
	ts, err := Titulares()
	if err != nil || len(ts) != 2 {
		t.Errorf("titulares = %v, %v", ts, err)
	}
}
