// Cajero: Retiro / Abono / Log JSON, como CLI interactivo o como App Web.
//
//	go run ./cmd/cajero                    CLI interactivo (cuenta de Alice)
//	go run ./cmd/cajero -modo web          App Web en http://localhost:8080
//	go run ./cmd/cajero -titular Hamza     otra cuenta
//
// Datos iniciales: internal/datos/cuentas.json (incrustado con go:embed).
// Log de transacciones: transacciones.json (una línea JSON por transacción).
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"prjwithdrawal/internal/banco"
	"prjwithdrawal/internal/cli"
	"prjwithdrawal/internal/datos"
	"prjwithdrawal/internal/web"
)

func main() {
	modo := flag.String("modo", "cli", "interfaz: cli | web")
	addr := flag.String("addr", "localhost:8080", "dirección del servidor web")
	titular := flag.String("titular", "Alice", "titular de la cuenta")
	archivoLog := flag.String("log", "transacciones.json", "archivo del log de transacciones")
	flag.Parse()

	cuenta, err := datos.BuscarCuenta(*titular)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	f, err := os.OpenFile(*archivoLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error al abrir el log:", err)
		os.Exit(1)
	}
	defer f.Close()

	cajero := banco.NuevoCajero(cuenta, banco.NuevoRegistroJSON(f), nil)

	switch *modo {
	case "cli":
		cli.Run(os.Stdin, os.Stdout, cajero)
	case "web":
		log.Printf("App Web en http://%s (Ctrl+C para salir)", *addr)
		log.Fatal(http.ListenAndServe(*addr, web.Handler(cajero)))
	default:
		fmt.Fprintf(os.Stderr, "modo %q no válido (use cli o web)\n", *modo)
		os.Exit(2)
	}
}
