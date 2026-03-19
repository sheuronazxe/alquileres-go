package main

import (
	"fmt"
	"strconv"
	"time"
)

const configPath = "datos/configuracion.json"

func main() {
	fmt.Println("╔══════════════════════════════════════╗")
	fmt.Println("║   Gestión de Recibos de Alquiler     ║")
	fmt.Println("╠══════════════════════════════════════╣")
	fmt.Println("║  1. Emitir nueva tanda de recibos    ║")
	fmt.Println("║  2. Reimprimir recibos anteriores    ║")
	fmt.Println("║  0. Salir                            ║")
	fmt.Println("╚══════════════════════════════════════╝")

	opcion := leerLinea("\nOpción: ")

	switch opcion {
	case "1":
		emitirTanda()
	case "2":
		reimprimir()
	case "0":
		return
	default:
		fatal("Opción no válida: %s", opcion)
	}
}

// ── 1. Emitir nueva tanda ───────────────────────────────────────────────

func emitirTanda() {
	cfg, err := cargarConfig(configPath)
	if err != nil {
		fatal("Error leyendo configuración: %v", err)
	}

	arrendador := Arrendador{Nombre: cfg.Arrendador, DNI: cfg.DNI}

	inquilinos, err := LoadData[Inquilino]("datos/inquilinos.csv", 9)
	if err != nil {
		fatal("Error leyendo inquilinos.csv: %v", err)
	}

	fecha, err := obtenerFecha()
	if err != nil {
		fatal("Error al obtener la fecha: %v", err)
	}

	fechaStr := formatMesAnyo(fecha)
	mostrarTablaInquilinos(inquilinos, fechaStr)

	seleccionados, err := filtrarRegistros(inquilinos)
	if err != nil {
		fatal("Error: %v", err)
	}

	nombre := leerLinea(fmt.Sprintf("Nombre de la tanda [RENTAS %s]: ",
		fmt.Sprintf("%s %d", meses[fecha.Month()], fecha.Year())))
	if nombre == "" {
		nombre = fmt.Sprintf("RENTAS %s %d", meses[fecha.Month()], fecha.Year())
	}

	// Crear snapshots inmutables con referencias correlativas
	recibos := crearSnapshotRecibos(seleccionados, arrendador, fecha, cfg)

	filename := fmt.Sprintf("alquileres_%s.pdf", fecha.Format("2006-01"))

	fmt.Printf("\nGenerando PDF con %d recibo(s)...\n", len(recibos))

	if err := generarPDF(recibos, fecha, filename); err != nil {
		fatal("Error generando PDF: %v", err)
	}

	// Guardar contador actualizado
	if err := guardarConfig(configPath, cfg); err != nil {
		fatal("Error guardando configuración: %v", err)
	}

	// Registrar tanda en el historial del año
	if err := registrarTanda(nombre, fecha.Format("2006-01"), fecha.Year(), recibos); err != nil {
		fatal("Error guardando registro: %v", err)
	}

	fmt.Println("\n✓ PDF generado: " + filename)
	fmt.Println("✓ Tanda registrada: " + nombre)
	openBrowser(filename)
}

// ── 2. Reimprimir recibos ───────────────────────────────────────────────

func reimprimir() {
	años, err := añosDisponibles()
	if err != nil {
		fatal("Error buscando registros: %v", err)
	}
	if len(años) == 0 {
		fmt.Println("No hay registros disponibles.")
		return
	}

	// Mostrar años disponibles
	fmt.Println()
	for i, año := range años {
		fmt.Printf("\033[32m%2d\033[0m  %d\n", i, año)
	}

	input := leerLinea("\nSeleccione año: ")
	idx, err := strconv.Atoi(input)
	if err != nil || idx < 0 || idx >= len(años) {
		fatal("Selección no válida")
	}

	// Cargar tandas del año seleccionado
	tandas, err := cargarRegistroAño(años[idx])
	if err != nil {
		fatal("Error cargando registro: %v", err)
	}
	if len(tandas) == 0 {
		fmt.Println("No hay tandas en este año.")
		return
	}

	mostrarHistorial(tandas)

	input = leerLinea("\nSeleccione tanda: ")
	idx, err = strconv.Atoi(input)
	if err != nil || idx < 0 || idx >= len(tandas) {
		fatal("Selección no válida")
	}

	tanda := tandas[idx]

	fecha, err := time.Parse("2006-01", tanda.Fecha)
	if err != nil {
		fatal("Error parseando fecha de tanda: %v", err)
	}

	// Mostrar recibos para selección individual
	mostrarRecibosTanda(tanda.Recibos)

	recibos, err := filtrarRegistros(tanda.Recibos)
	if err != nil {
		fatal("Error: %v", err)
	}

	filename := fmt.Sprintf("reimpresion_%s.pdf", time.Now().Format("2006-01-02"))

	fmt.Printf("\nReimprimiendo %d recibo(s)...\n", len(recibos))

	if err := generarPDF(recibos, fecha, filename); err != nil {
		fatal("Error generando PDF: %v", err)
	}

	fmt.Println("\n✓ PDF generado: " + filename)
	openBrowser(filename)
}
