package banco

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
)

// Resultados posibles de una transacción en el log.
const (
	ResultadoExito = "exito" // operación aplicada a la cuenta
	ResultadoFallo = "fallo" // rechazada por una regla de negocio (RN-01, RN-02, monto <= 0)
	ResultadoError = "error" // entrada ilegible o falla técnica
)

// Tipos de operación.
const (
	TipoRetiro = "retiro"
	TipoAbono  = "abono"
)

var (
	// ErrRegistro: no se pudo escribir el log; la operación se revierte (integridad).
	ErrRegistro = errors.New("no se pudo registrar la transacción: operación revertida")
	// ErrTipoInvalido: el tipo de operación no es retiro ni abono.
	ErrTipoInvalido = errors.New("tipo de operación inválido")
)

// Transaccion es un registro del log. Se serializa como una línea JSON.
type Transaccion struct {
	ID           int    `json:"id"`
	Fecha        string `json:"fecha"`
	Titular      string `json:"titular"`
	Tipo         string `json:"tipo"`
	Monto        string `json:"monto"`
	Resultado    string `json:"resultado"`
	Detalle      string `json:"detalle,omitempty"`
	SaldoAntes   string `json:"saldo_antes"`
	SaldoDespues string `json:"saldo_despues"`
}

// Registro es el destino del log (archivo, memoria, etc.).
type Registro interface {
	Guardar(Transaccion) error
}

// RegistroJSON escribe una transacción por línea (JSON Lines) en un io.Writer.
type RegistroJSON struct {
	mu sync.Mutex
	w  io.Writer
}

func NuevoRegistroJSON(w io.Writer) *RegistroJSON { return &RegistroJSON{w: w} }

func (r *RegistroJSON) Guardar(t Transaccion) error {
	b, err := json.Marshal(t)
	if err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	_, err = r.w.Write(append(b, '\n'))
	return err
}

// Cajero aplica operaciones sobre una Cuenta y deja constancia de cada una.
// Integridad: si la transacción no puede registrarse, el saldo se restaura.
// Es seguro para uso concurrente (la App Web atiende varias peticiones a la vez).
type Cajero struct {
	mu        sync.Mutex
	cuenta    *Cuenta
	registro  Registro
	reloj     func() time.Time
	historial []Transaccion
}

// NuevoCajero crea un cajero. reloj puede ser nil (usa time.Now).
func NuevoCajero(c *Cuenta, r Registro, reloj func() time.Time) *Cajero {
	if reloj == nil {
		reloj = time.Now
	}
	return &Cajero{cuenta: c, registro: r, reloj: reloj}
}

// Cuenta devuelve la cuenta asociada (acceso sin bloqueo: usar solo en código de un solo hilo).
func (k *Cajero) Cuenta() *Cuenta { return k.cuenta }

// Saldo devuelve el saldo actual en centavos, de forma segura entre hilos.
func (k *Cajero) Saldo() int64 {
	k.mu.Lock()
	defer k.mu.Unlock()
	return k.cuenta.saldo
}

// Historial devuelve una copia de las transacciones registradas.
func (k *Cajero) Historial() []Transaccion {
	k.mu.Lock()
	defer k.mu.Unlock()
	return append([]Transaccion(nil), k.historial...)
}

// Retirar ejecuta el retiro y lo registra.
func (k *Cajero) Retirar(monto int64) (Transaccion, error) {
	k.mu.Lock()
	defer k.mu.Unlock()
	return k.operar(TipoRetiro, monto, k.cuenta.Retirar)
}

// Abonar ejecuta el abono y lo registra.
func (k *Cajero) Abonar(monto int64) (Transaccion, error) {
	k.mu.Lock()
	defer k.mu.Unlock()
	return k.operar(TipoAbono, monto, k.cuenta.Abonar)
}

// Ejecutar es el punto de entrada común de las interfaces (CLI y Web): recibe el
// tipo y el monto como texto, lo interpreta, aplica la operación y la registra.
// Texto ilegible => se registra como "error" y el saldo no cambia.
func (k *Cajero) Ejecutar(tipo, textoMonto string) (Transaccion, error) {
	k.mu.Lock()
	defer k.mu.Unlock()

	var op func(int64) error
	switch tipo {
	case TipoRetiro:
		op = k.cuenta.Retirar
	case TipoAbono:
		op = k.cuenta.Abonar
	default:
		return Transaccion{}, ErrTipoInvalido
	}

	monto, err := ParseMonto(textoMonto)
	if err != nil {
		if errLog := k.registrarError(tipo, textoMonto, err); errLog != nil {
			return Transaccion{}, errLog
		}
		return k.historial[len(k.historial)-1], err
	}
	return k.operar(tipo, monto, op)
}

// RegistrarError deja constancia de una entrada ilegible (p. ej. "abc") sin tocar el saldo.
func (k *Cajero) RegistrarError(tipo, textoIngresado string, causa error) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	return k.registrarError(tipo, textoIngresado, causa)
}

func (k *Cajero) registrarError(tipo, textoIngresado string, causa error) error {
	s := FormatoMonto(k.cuenta.saldo)
	t := k.nueva(tipo, textoIngresado, ResultadoError, causa.Error(), s, s)
	if err := k.guardar(t); err != nil {
		return fmt.Errorf("%w: %v", ErrRegistro, err)
	}
	return nil
}

func (k *Cajero) operar(tipo string, monto int64, op func(int64) error) (Transaccion, error) {
	antes := k.cuenta.saldo
	opErr := op(monto)

	resultado, detalle := ResultadoExito, ""
	if opErr != nil {
		resultado, detalle = ResultadoFallo, opErr.Error()
	}
	t := k.nueva(tipo, FormatoMonto(monto), resultado, detalle,
		FormatoMonto(antes), FormatoMonto(k.cuenta.saldo))

	if err := k.guardar(t); err != nil {
		k.cuenta.saldo = antes // rollback: sin registro no hay transacción
		return Transaccion{}, fmt.Errorf("%w: %v", ErrRegistro, err)
	}
	return t, opErr
}

func (k *Cajero) nueva(tipo, monto, resultado, detalle, antes, despues string) Transaccion {
	return Transaccion{
		ID:           len(k.historial) + 1,
		Fecha:        k.reloj().Format(time.RFC3339),
		Titular:      k.cuenta.titular,
		Tipo:         tipo,
		Monto:        monto,
		Resultado:    resultado,
		Detalle:      detalle,
		SaldoAntes:   antes,
		SaldoDespues: despues,
	}
}

func (k *Cajero) guardar(t Transaccion) error {
	if err := k.registro.Guardar(t); err != nil {
		return err
	}
	k.historial = append(k.historial, t)
	return nil
}
