package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// ── Configuración ───────────────────────────────────────────────────────

type Configuracion struct {
	Arrendador string `json:"arrendador"`
	DNI        string `json:"DNI"`
	Mes        string `json:"mes"`
	ID         int    `json:"id"`
}

func cargarConfig(path string) (*Configuracion, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Configuracion
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func guardarConfig(path string, cfg *Configuracion) error {
	data, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0644)
}

// siguienteReferencia devuelve "YYYY-MM/NNNN" y avanza el contador.
func (cfg *Configuracion) siguienteReferencia(fecha time.Time) string {
	mesActual := fecha.Format("2006-01")
	if cfg.Mes != mesActual {
		cfg.Mes = mesActual
		cfg.ID = 1
	}
	ref := fmt.Sprintf("%s/%04d", mesActual, cfg.ID)
	cfg.ID++
	return ref
}

// ── Arrendador ──────────────────────────────────────────────────────────

type Arrendador struct {
	Nombre string
	DNI    string
}

// ── Inquilino (leído desde CSV) ─────────────────────────────────────────

type Inquilino struct {
	Nombre      string
	DNI         string
	Inmueble    string
	Renta       float64
	Empresa     bool
	Comentarios string
}

func (inq *Inquilino) FromRow(linea []string) error {
	inq.Nombre = linea[0]
	inq.DNI = linea[1]
	inq.Inmueble = linea[3]

	renta, err := parseMoneda(linea[4])
	if err != nil {
		return fmt.Errorf("renta '%s': %v", linea[4], err)
	}
	inq.Renta = renta

	esEmpresa, err := strconv.ParseBool(strings.TrimSpace(linea[7]))
	if err != nil {
		return fmt.Errorf("empresa '%s': %v", linea[7], err)
	}
	inq.Empresa = esEmpresa
	inq.Comentarios = linea[8]

	return nil
}

// ── Recibo registrado (snapshot inmutable) ───────────────────────────────

type ReciboRegistrado struct {
	Referencia  string  `json:"referencia"`
	Arrendador  string  `json:"arrendador"`
	ArrendadorDNI string `json:"arrendador_dni"`
	Nombre      string  `json:"nombre"`
	DNI         string  `json:"dni"`
	Inmueble    string  `json:"inmueble"`
	Renta       float64 `json:"renta"`
	Empresa     bool    `json:"empresa"`
	Comentarios string  `json:"comentarios,omitempty"`
}

// ── Tanda (grupo de recibos emitidos juntos) ────────────────────────────

type Tanda struct {
	Nombre  string             `json:"nombre"`
	Fecha   string             `json:"fecha"`
	Recibos []ReciboRegistrado `json:"recibos"`
}
