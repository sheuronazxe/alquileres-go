package main

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	"codeberg.org/go-pdf/fpdf"
	"alquileres/fonts"
)

const (
	ivaPct       = 21.0
	retencionPct = 19.0

	recibosPorPagina = 3
	anchoPagina      = 210.0
	alturaRecibo     = 297.0 / float64(recibosPorPagina)
)

// ── PDF base ────────────────────────────────────────────────────────────

func newPDF() (*fpdf.Fpdf, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(false, 0)

	type fontDef struct{ family, style, file string }
	for _, f := range []fontDef{
		{"Roboto", "", "Roboto-Regular.ttf"},
		{"Roboto", "B", "Roboto-Bold.ttf"},
		{"Roboto-Light", "", "Roboto-Light.ttf"},
	} {
		bytes, err := fonts.FS.ReadFile(f.file)
		if err != nil {
			return nil, fmt.Errorf("error cargando fuente %s: %w", f.file, err)
		}
		pdf.AddUTF8FontFromBytes(f.family, f.style, bytes)
	}
	return pdf, nil
}

// ── Generar PDF de recibos ──────────────────────────────────────────────

func generarPDF(recibos []ReciboRegistrado, fecha time.Time, filename string) error {
	pdf, err := newPDF()
	if err != nil {
		return err
	}
	pdf.SetCellMargin(0)
	pdf.SetLeftMargin(15)

	// Ordenar por inmueble
	slices.SortFunc(recibos, func(a, b ReciboRegistrado) int {
		return strings.Compare(a.Inmueble, b.Inmueble)
	})

	imprimirPaginas(pdf, recibos, func(r ReciboRegistrado, yOffset float64) {
		dibujarRecibo(pdf, r, fecha, yOffset)
	})

	return pdf.OutputFileAndClose(filename)
}

// imprimirPaginas distribuye recibos en páginas con cut & stack ordering.
func imprimirPaginas(pdf *fpdf.Fpdf, datos []ReciboRegistrado, draw func(ReciboRegistrado, float64)) {
	totalPaginas := (len(datos) + recibosPorPagina - 1) / recibosPorPagina
	if totalPaginas == 0 {
		return
	}

	// Cut & stack: distribuir para que al cortar y apilar queden en orden
	paginas := make([][]ReciboRegistrado, totalPaginas)
	for i, r := range datos {
		pagina := i % totalPaginas
		paginas[pagina] = append(paginas[pagina], r)
	}

	for _, registros := range paginas {
		pdf.AddPage()
		for pos, r := range registros {
			yOffset := float64(pos) * alturaRecibo
			draw(r, yOffset)
		}
	}
}

// ── Dibujar recibo individual ───────────────────────────────────────────

func dibujarRecibo(pdf *fpdf.Fpdf, r ReciboRegistrado, fecha time.Time, y float64) {
	dibujarMarcasCorte(pdf, y)
	dibujarTitulo(pdf, r, fecha, y)
	dibujarColumnaIzquierda(pdf, r)
	dibujarColumnaDerecha(pdf, r, y)
}

func dibujarMarcasCorte(pdf *fpdf.Fpdf, y float64) {
	pdf.SetLineWidth(0.3)
	pdf.Line(0, y+alturaRecibo, 10, y+alturaRecibo)
	pdf.Line(anchoPagina-10, y+alturaRecibo, anchoPagina, y+alturaRecibo)
}

func dibujarTitulo(pdf *fpdf.Fpdf, r ReciboRegistrado, fecha time.Time, y float64) {
	titulo := fmt.Sprintf("Recibo de alquiler, %s", formatMesAnyo(fecha))

	pdf.SetY(y + 10)
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Roboto-Light", "", 18)
	pdf.CellFormat(180, 7, titulo, "", 1, "TC", false, 0, "")

	pdf.SetFont("Roboto", "", 10)
	pdf.SetTextColor(100, 100, 100)
	pdf.SetLineWidth(0.8)
	pdf.CellFormat(90, 7, fmt.Sprintf("Referencia: %s", r.Referencia), "B", 0, "L", false, 0, "")
	pdf.CellFormat(90, 7, fmt.Sprintf("Fecha: 1 de %s de %d", meses[fecha.Month()], fecha.Year()), "B", 1, "R", false, 0, "")
}

func dibujarColumnaIzquierda(pdf *fpdf.Fpdf, r ReciboRegistrado) {
	dibujarBloquePersona(pdf, "ARRENDADOR", r.Arrendador, "DNI/NIE:", r.ArrendadorDNI)
	dibujarBloquePersona(pdf, "INQUILINO", r.Nombre, "DNI/NIF:", r.DNI)

	pdf.Ln(3)
	cabecera(pdf, "INMUEBLE", 105)

	pdf.SetFont("Roboto", "", 11)
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(90, 8, r.Inmueble, "", 1, "L", false, 0, "")
}

func cabecera(pdf *fpdf.Fpdf, texto string, width float64) {
	pdf.SetCellMargin(0)
	pdf.SetLineWidth(0.1)
	pdf.SetFont("Roboto", "B", 9)
	pdf.SetTextColor(0, 0, 0)
	pdf.SetDashPattern([]float64{1, 1}, 0)
	pdf.CellFormat(width, 7, texto, "B", 2, "", false, 0, "")
	pdf.SetDashPattern([]float64{}, 0)
}

func dibujarBloquePersona(pdf *fpdf.Fpdf, titulo, nombre, etiquetaDoc, doc string) {
	pdf.Ln(3)
	cabecera(pdf, titulo, 105)

	pdf.SetCellMargin(2)
	pdf.SetFont("Roboto", "", 10)
	pdf.SetTextColor(0, 0, 0)

	pdf.CellFormat(20, 8, "Nombre:", "0", 0, "", false, 0, "")
	pdf.SetFont("Roboto", "B", 11)
	pdf.CellFormat(62.5, 8, nombre, "0", 1, "", false, 0, "")

	pdf.SetFont("Roboto", "", 10)
	pdf.CellFormat(20, 8, etiquetaDoc, "0", 0, "", false, 0, "")
	pdf.SetFont("Roboto", "B", 11)
	pdf.CellFormat(62.5, 8, doc, "0", 1, "", false, 0, "")
}

func dibujarColumnaDerecha(pdf *fpdf.Fpdf, r ReciboRegistrado, y float64) {
	anchoConcepto := 40.0
	anchoImporte := 25.0
	segundaColumna := 130.0

	pdf.SetXY(segundaColumna, y+27)
	pdf.SetLeftMargin(segundaColumna)
	defer pdf.SetLeftMargin(15)

	cabecera(pdf, "DESGLOSE", 65)
	pdf.Ln(5)

	// Encabezados
	pdf.SetFillColor(230, 230, 230)
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Roboto", "B", 9)
	pdf.CellFormat(40, 6.5, "  Concepto", "1", 0, "L", true, 0, "")
	pdf.CellFormat(25, 6.5, "Importe  ", "1", 1, "R", true, 0, "")

	if r.Empresa {
		base := r.Renta
		iva := math.Round(base*ivaPct) / 100
		retencion := math.Round(base*retencionPct) / 100
		total := base + iva - retencion

		pdf.SetTextColor(0, 0, 0)
		pdf.SetFont("Roboto", "", 9)
		pdf.CellFormat(anchoConcepto, 6.5, "  Base imponible", "LR", 0, "L", false, 0, "")
		pdf.SetFont("Roboto", "B", 9)
		pdf.CellFormat(anchoImporte, 6.5, formatRenta(base)+"  ", "LR", 1, "R", false, 0, "")

		pdf.SetFont("Roboto", "", 9)
		pdf.CellFormat(anchoConcepto, 6.5, fmt.Sprintf("  IVA (%.0f%%)", ivaPct), "LR", 0, "L", false, 0, "")
		pdf.SetFont("Roboto", "B", 9)
		pdf.CellFormat(anchoImporte, 6.5, formatRenta(iva)+"  ", "LR", 1, "R", false, 0, "")

		pdf.SetFont("Roboto", "", 9)
		pdf.CellFormat(anchoConcepto, 6.5, fmt.Sprintf("  Retención IRPF (%.0f%%)", retencionPct), "LBR", 0, "L", false, 0, "")
		pdf.SetFont("Roboto", "B", 9)
		pdf.CellFormat(anchoImporte, 6.5, "-"+formatRenta(retencion)+"  ", "LBR", 1, "R", false, 0, "")

		pdf.SetTextColor(0, 0, 0)
		pdf.SetFont("Roboto", "B", 11)
		pdf.CellFormat(anchoConcepto, 9.5, "  TOTAL", "1", 0, "L", true, 0, "")
		pdf.CellFormat(anchoImporte, 9.5, formatRenta(total)+"  ", "1", 1, "R", true, 0, "")
	} else {
		pdf.SetTextColor(0, 0, 0)
		pdf.SetFont("Roboto", "", 9)
		pdf.CellFormat(anchoConcepto, 6.5, "  Renta base mensual", "LBR", 0, "L", false, 0, "")
		pdf.SetFont("Roboto", "B", 9)
		pdf.CellFormat(anchoImporte, 6.5, formatRenta(r.Renta)+"  ", "LBR", 1, "R", false, 0, "")

		pdf.SetTextColor(0, 0, 0)
		pdf.SetFont("Roboto", "B", 11)
		pdf.CellFormat(anchoConcepto, 9.5, "  TOTAL", "1", 0, "L", true, 0, "")
		pdf.CellFormat(anchoImporte, 9.5, formatRenta(r.Renta)+"  ", "1", 1, "R", true, 0, "")
	}

	pdf.Ln(5)
	pdf.SetFont("Roboto", "", 10)
	pdf.MultiCell(65, 5, r.Comentarios, "", "L", false)
}
