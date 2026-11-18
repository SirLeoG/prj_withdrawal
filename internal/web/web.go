// Package web implementa la App Web: una página HTML y una API JSON sobre el mismo Cajero
// que usa el CLI (la lógica de negocio no se duplica).
package web

import (
	"embed"
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"strconv"

	"prjwithdrawal/internal/banco"
)

//go:embed templates/index.html
var plantillas embed.FS

type servidor struct {
	cajero *banco.Cajero
	tpl    *template.Template
}

// Handler devuelve el http.Handler de la aplicación.
func Handler(k *banco.Cajero) http.Handler {
	s := &servidor{cajero: k, tpl: template.Must(template.ParseFS(plantillas, "templates/index.html"))}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.inicio)
	mux.HandleFunc("POST /operar", s.operar)
	mux.HandleFunc("GET /api/saldo", s.apiSaldo)
	mux.HandleFunc("GET /api/transacciones", s.apiTransacciones)
	mux.HandleFunc("POST /api/retiro", s.apiOperar(banco.TipoRetiro))
	mux.HandleFunc("POST /api/abono", s.apiOperar(banco.TipoAbono))
	return mux
}

type mensaje struct{ Resultado, Titulo, Detalle string }

type vista struct {
	Titular   string
	Saldo     string
	Mensaje   *mensaje
	Historial []banco.Transaccion // más reciente primero
}

func (s *servidor) inicio(w http.ResponseWriter, r *http.Request) {
	h := s.cajero.Historial()
	rev := make([]banco.Transaccion, len(h))
	for i, t := range h {
		rev[len(h)-1-i] = t
	}
	v := vista{
		Titular:   s.cajero.Cuenta().Titular(),
		Saldo:     banco.FormatoMonto(s.cajero.Saldo()),
		Historial: rev,
	}
	// El mensaje se arma en el servidor a partir del id de la transacción (no del texto de la URL).
	if id, err := strconv.Atoi(r.URL.Query().Get("id")); err == nil && id >= 1 && id <= len(h) {
		v.Mensaje = describir(h[id-1])
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tpl.Execute(w, v); err != nil {
		http.Error(w, "error al generar la página", http.StatusInternalServerError)
	}
}

func describir(t banco.Transaccion) *mensaje {
	m := &mensaje{Resultado: t.Resultado}
	switch t.Resultado {
	case banco.ResultadoExito:
		m.Titulo = "Operación exitosa."
		m.Detalle = t.Tipo + " de S/ " + t.Monto + "."
	case banco.ResultadoFallo:
		m.Titulo = "Operación rechazada."
		m.Detalle = t.Detalle
	default:
		m.Titulo = "Entrada no válida."
		m.Detalle = t.Detalle
	}
	return m
}

func (s *servidor) operar(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "formulario inválido", http.StatusBadRequest)
		return
	}
	tx, err := s.cajero.Ejecutar(r.PostFormValue("tipo"), r.PostFormValue("monto"))
	if errors.Is(err, banco.ErrRegistro) || errors.Is(err, banco.ErrTipoInvalido) {
		http.Error(w, err.Error(), estado(tx, err))
		return
	}
	http.Redirect(w, r, "/?id="+strconv.Itoa(tx.ID), http.StatusSeeOther)
}

// estado traduce el resultado de una operación a un código HTTP.
func estado(tx banco.Transaccion, err error) int {
	switch {
	case err == nil:
		return http.StatusOK
	case errors.Is(err, banco.ErrRegistro):
		return http.StatusInternalServerError
	case errors.Is(err, banco.ErrTipoInvalido):
		return http.StatusBadRequest
	case tx.Resultado == banco.ResultadoError:
		return http.StatusBadRequest // entrada ilegible
	default:
		return http.StatusUnprocessableEntity // regla de negocio
	}
}

func escribirJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func (s *servidor) apiSaldo(w http.ResponseWriter, _ *http.Request) {
	escribirJSON(w, http.StatusOK, map[string]string{
		"titular": s.cajero.Cuenta().Titular(),
		"saldo":   banco.FormatoMonto(s.cajero.Saldo()),
	})
}

func (s *servidor) apiTransacciones(w http.ResponseWriter, _ *http.Request) {
	h := s.cajero.Historial()
	if h == nil {
		h = []banco.Transaccion{}
	}
	escribirJSON(w, http.StatusOK, h)
}

func (s *servidor) apiOperar(tipo string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Monto string `json:"monto"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<10)).Decode(&body); err != nil {
			escribirJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON inválido: se esperaba {\"monto\":\"200.00\"}"})
			return
		}
		tx, err := s.cajero.Ejecutar(tipo, body.Monto)
		code := estado(tx, err)
		if err != nil && tx.ID == 0 {
			escribirJSON(w, code, map[string]string{"error": err.Error()})
			return
		}
		escribirJSON(w, code, tx)
	}
}
