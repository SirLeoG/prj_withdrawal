package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"prjwithdrawal/internal/banco"
)

func nuevo(saldo int64) (http.Handler, *banco.Cajero, *bytes.Buffer) {
	var log bytes.Buffer
	k := banco.NuevoCajero(banco.NuevaCuenta("Alice", saldo), banco.NuevoRegistroJSON(&log), nil)
	return Handler(k), k, &log
}

func peticion(h http.Handler, metodo, ruta, cuerpo, ctype string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo))
	if ctype != "" {
		r.Header.Set("Content-Type", ctype)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func TestPaginaInicioMuestraSaldo(t *testing.T) {
	h, _, _ := nuevo(23456)
	w := peticion(h, "GET", "/", "", "")
	if w.Code != 200 || !strings.Contains(w.Body.String(), "S/ 234.56") {
		t.Errorf("código %d, cuerpo:\n%s", w.Code, w.Body.String())
	}
}

func TestFormularioRetiroRedirigeYActualizaSaldo(t *testing.T) {
	h, k, _ := nuevo(23456)
	f := url.Values{"tipo": {"retiro"}, "monto": {"200"}}.Encode()
	w := peticion(h, "POST", "/operar", f, "application/x-www-form-urlencoded")
	if w.Code != http.StatusSeeOther || w.Header().Get("Location") != "/?id=1" {
		t.Fatalf("código %d, location %q", w.Code, w.Header().Get("Location"))
	}
	if k.Saldo() != 3456 {
		t.Errorf("saldo = %d", k.Saldo())
	}
	p := peticion(h, "GET", w.Header().Get("Location"), "", "")
	if !strings.Contains(p.Body.String(), "Operación exitosa") || !strings.Contains(p.Body.String(), "S/ 34.56") {
		t.Errorf("la página no muestra el resultado:\n%s", p.Body.String())
	}
}

func TestAPI_CodigosPorResultado(t *testing.T) {
	h, _, _ := nuevo(19876)
	casos := []struct {
		ruta, cuerpo, resultado string
		code                    int
	}{
		{"/api/abono", `{"monto":"50.00"}`, "exito", 200},
		{"/api/retiro", `{"monto":"500.00"}`, "fallo", 422}, // RN-01
		{"/api/retiro", `{"monto":"abc"}`, "error", 400},    // entrada ilegible
		{"/api/retiro", `{"monto":"-5"}`, "fallo", 422},     // monto <= 0
	}
	for _, c := range casos {
		w := peticion(h, "POST", c.ruta, c.cuerpo, "application/json")
		var tx banco.Transaccion
		json.Unmarshal(w.Body.Bytes(), &tx)
		if w.Code != c.code || tx.Resultado != c.resultado {
			t.Errorf("%s %s → %d %q; se esperaba %d %q", c.ruta, c.cuerpo, w.Code, tx.Resultado, c.code, c.resultado)
		}
	}
}

func TestAPI_JSONInvalido(t *testing.T) {
	h, _, log := nuevo(100)
	w := peticion(h, "POST", "/api/retiro", "no es json", "application/json")
	if w.Code != 400 || log.Len() != 0 {
		t.Errorf("código %d, log %q", w.Code, log.String())
	}
}

func TestAPI_SaldoYTransacciones(t *testing.T) {
	h, _, _ := nuevo(23456)
	var s map[string]string
	json.Unmarshal(peticion(h, "GET", "/api/saldo", "", "").Body.Bytes(), &s)
	if s["saldo"] != "234.56" || s["titular"] != "Alice" {
		t.Errorf("saldo = %v", s)
	}
	if got := strings.TrimSpace(peticion(h, "GET", "/api/transacciones", "", "").Body.String()); got != "[]" {
		t.Errorf("sin transacciones debía ser [], fue %s", got)
	}
	peticion(h, "POST", "/api/abono", `{"monto":"10"}`, "application/json")
	var txs []banco.Transaccion
	json.Unmarshal(peticion(h, "GET", "/api/transacciones", "", "").Body.Bytes(), &txs)
	if len(txs) != 1 || txs[0].Tipo != "abono" {
		t.Errorf("transacciones = %+v", txs)
	}
}

func TestMetodoNoPermitido(t *testing.T) {
	h, _, _ := nuevo(100)
	if w := peticion(h, "GET", "/api/retiro", "", ""); w.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET /api/retiro → %d", w.Code)
	}
}
