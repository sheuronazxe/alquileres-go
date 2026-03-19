package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
)

type CSVRowable interface {
	FromRow(linea []string) error
}

func LoadData[T any, PT interface {
	*T
	CSVRowable
}](path string, numColumnas int) ([]T, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Manejo de BOM UTF-8
	br := bufio.NewReader(f)
	if bom, _ := br.Peek(3); string(bom) == "\xEF\xBB\xBF" {
		br.Discard(3)
	}

	r := csv.NewReader(br)
	r.TrimLeadingSpace = true
	r.FieldsPerRecord = numColumnas

	// Saltar cabecera
	if _, err := r.Read(); err != nil {
		return nil, fmt.Errorf("error leyendo cabecera CSV (%s): %v", path, err)
	}

	rows, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error leyendo CSV (%s): %v", path, err)
	}

	var results []T
	for i, row := range rows {
		var item T
		ptr := PT(&item)
		if err := ptr.FromRow(row); err != nil {
			return nil, fmt.Errorf("error en fila %d de %s: %v", i+2, path, err)
		}
		results = append(results, item)
	}

	return results, nil
}
