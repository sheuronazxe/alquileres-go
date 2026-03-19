package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// rutaRegistro devuelve "datos/registro_YYYY.json" para el año dado.
func rutaRegistro(año int) string {
	return fmt.Sprintf("datos/registro_%d.json", año)
}

func cargarRegistroAño(año int) ([]Tanda, error) {
	data, err := os.ReadFile(rutaRegistro(año))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var tandas []Tanda
	if err := json.Unmarshal(data, &tandas); err != nil {
		return nil, err
	}
	return tandas, nil
}

func guardarRegistroAño(año int, tandas []Tanda) error {
	data, err := json.MarshalIndent(tandas, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(rutaRegistro(año), append(data, '\n'), 0644)
}

// registrarTanda añade la tanda al fichero de registro del año correspondiente.
func registrarTanda(nombre string, fechaMes string, año int, recibos []ReciboRegistrado) error {
	tandas, err := cargarRegistroAño(año)
	if err != nil {
		return fmt.Errorf("error cargando registro %d: %w", año, err)
	}

	tanda := Tanda{
		Nombre:  nombre,
		Fecha:   fechaMes,
		Recibos: recibos,
	}

	tandas = append(tandas, tanda)
	return guardarRegistroAño(año, tandas)
}

// añosDisponibles busca ficheros datos/registro_YYYY.json y devuelve los años ordenados.
func añosDisponibles() ([]int, error) {
	matches, err := filepath.Glob("datos/registro_*.json")
	if err != nil {
		return nil, err
	}

	var años []int
	for _, m := range matches {
		base := filepath.Base(m)                            // registro_2026.json
		base = strings.TrimPrefix(base, "registro_")       // 2026.json
		base = strings.TrimSuffix(base, ".json")           // 2026
		if año, err := strconv.Atoi(base); err == nil {
			años = append(años, año)
		}
	}
	sort.Sort(sort.Reverse(sort.IntSlice(años)))
	return años, nil
}

// crearSnapshotRecibos genera los ReciboRegistrado (snapshots inmutables) a partir de Inquilinos.
func crearSnapshotRecibos(inquilinos []Inquilino, arrendador Arrendador, fecha time.Time, cfg *Configuracion) []ReciboRegistrado {
	recibos := make([]ReciboRegistrado, len(inquilinos))
	for i, inq := range inquilinos {
		recibos[i] = ReciboRegistrado{
			Referencia:    cfg.siguienteReferencia(fecha),
			Arrendador:    arrendador.Nombre,
			ArrendadorDNI: arrendador.DNI,
			Nombre:        inq.Nombre,
			DNI:           inq.DNI,
			Inmueble:      inq.Inmueble,
			Renta:         inq.Renta,
			Empresa:       inq.Empresa,
			Comentarios:   inq.Comentarios,
		}
	}
	return recibos
}

// mostrarHistorial imprime la lista de tandas con formato para selección.
func mostrarHistorial(tandas []Tanda) {
	fmt.Println()
	for i, t := range tandas {
		fmt.Printf("\033[32m%2d\033[0m  %-40s  %s  (%d recibos)\n",
			i, t.Nombre, t.Fecha, len(t.Recibos))
	}
}

// mostrarRecibosTanda imprime la lista de recibos de una tanda para selección.
func mostrarRecibosTanda(recibos []ReciboRegistrado) {
	fmt.Println()
	for i, r := range recibos {
		fmt.Printf("\033[32m%2d\033[0m  %-10s  %-30.30s  \033[33m%-25.25s\033[0m  %10s\n",
			i, r.Referencia, r.Nombre, r.Inmueble, formatRenta(r.Renta))
	}
}
