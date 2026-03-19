package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// ── Entrada ─────────────────────────────────────────────────────────────

var inputReader = bufio.NewScanner(os.Stdin)

func leerLinea(prompt string) string {
	fmt.Print(prompt)
	inputReader.Scan()
	return strings.TrimSpace(inputReader.Text())
}

// ── Fechas ──────────────────────────────────────────────────────────────

var meses = map[time.Month]string{
	time.January: "enero", time.February: "febrero", time.March: "marzo",
	time.April: "abril", time.May: "mayo", time.June: "junio",
	time.July: "julio", time.August: "agosto", time.September: "septiembre",
	time.October: "octubre", time.November: "noviembre", time.December: "diciembre",
}

// obtenerFecha devuelve el mes/año. Si el día > 3, propone el mes siguiente.
func obtenerFecha() (time.Time, error) {
	t := time.Now()
	if t.Day() > 3 {
		t = t.AddDate(0, 1, 0)
	}

	def := fmt.Sprintf("%s %d", meses[t.Month()], t.Year())
	input := leerLinea(fmt.Sprintf("Introduce fecha [%s]: ", def))
	if input == "" {
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location()), nil
	}

	partes := strings.Fields(input)
	if len(partes) != 2 {
		return time.Time{}, fmt.Errorf("formato incorrecto. Usa: 'Mes Año'")
	}

	nombreMes := strings.ToLower(partes[0])
	mes := time.Month(0)
	for m, n := range meses {
		if n == nombreMes {
			mes = m
			break
		}
	}
	if mes == 0 {
		return time.Time{}, fmt.Errorf("mes inválido: %s", nombreMes)
	}

	año, err := strconv.Atoi(partes[1])
	if err != nil {
		return time.Time{}, fmt.Errorf("el año debe ser numérico")
	}

	return time.Date(año, mes, 1, 0, 0, 0, 0, t.Location()), nil
}

func formatMesAnyo(fecha time.Time) string {
	return fmt.Sprintf("%s %d", meses[fecha.Month()], fecha.Year())
}

// ── Moneda ──────────────────────────────────────────────────────────────

var formatoMoneda = strings.NewReplacer("€", "", ".", "", ",", ".")

func parseMoneda(s string) (float64, error) {
	return strconv.ParseFloat(strings.TrimSpace(formatoMoneda.Replace(s)), 64)
}

func formatRenta(val float64) string {
	val = math.Round(val*100) / 100

	signo := ""
	if val < 0 {
		signo = "-"
		val = -val
	}

	s := fmt.Sprintf("%.2f", val)
	s = strings.ReplaceAll(s, ".", ",")
	parts := strings.Split(s, ",")
	intPart, decPart := parts[0], parts[1]

	if len(intPart) > 3 {
		var result []byte
		for i, c := range intPart {
			if i > 0 && (len(intPart)-i)%3 == 0 {
				result = append(result, '.')
			}
			result = append(result, byte(c))
		}
		intPart = string(result)
	}

	return signo + intPart + "," + decPart + " €"
}

// ── Sistema ─────────────────────────────────────────────────────────────

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		fmt.Printf("No se puede abrir en esta plataforma: %s\n", runtime.GOOS)
		return
	}
	_ = cmd.Start()
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

// ── Selección interactiva ───────────────────────────────────────────────

func mostrarTablaInquilinos(datos []Inquilino, fechaStr string) {
	fmt.Println()
	for i, inq := range datos {
		fmt.Printf("\033[32m%2d\033[0m %-30.30s  \033[33m%-25.25s\033[0m  %10s\n",
			i, inq.Nombre, inq.Inmueble, formatRenta(inq.Renta))
	}
	fmt.Printf("\n  Fecha: \033[36m%s\033[0m\n", fechaStr)
}

func filtrarRegistros[T any](datos []T) ([]T, error) {
	input := leerLinea("¿Qué registros desea imprimir? (ej: 1,3,10) [todos]: ")

	if input == "" {
		return datos, nil
	}

	var filtrados []T
	for s := range strings.SplitSeq(input, ",") {
		if i, err := strconv.Atoi(strings.TrimSpace(s)); err == nil && i >= 0 && i < len(datos) {
			filtrados = append(filtrados, datos[i])
		}
	}
	if len(filtrados) == 0 {
		return nil, fmt.Errorf("selección no válida")
	}
	return filtrados, nil
}
