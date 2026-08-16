// Package parse loads flight records from CSV files.
package parse

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// Flight is a single scheduled flight leg with its observed departure delay.
type Flight struct {
	FlightNo  string
	Origin    string
	Dest      string
	Cause     string
	Season    string
	SchedDep  int
	ActualDep int
	DelayMin  float64
	Hour      float64
}

// Route returns the canonical route key "Origin-Dest".
func (f Flight) Route() string {
	return f.Origin + "-" + f.Dest
}

const columnCount = 9

// ParseFlights reads a flight CSV file and returns its records.
//
// The file must start with the header
// flightno,origin,dest,scheddep,actualdep,delaymin,cause,season,hour
// and every following row must have exactly 9 columns with numeric
// scheddep, actualdep, delaymin and hour values. Any missing file,
// missing column or non-numeric field yields an error and no records.
func ParseFlights(path string) ([]Flight, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("parse: open %s: %w", path, err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	r.TrimLeadingSpace = true

	header, err := r.Read()
	if err == io.EOF {
		return nil, fmt.Errorf("parse: %s: empty file, header required", path)
	}
	if err != nil {
		return nil, fmt.Errorf("parse: %s: read header: %w", path, err)
	}
	if len(header) != columnCount {
		return nil, fmt.Errorf("parse: %s: header has %d columns, want %d", path, len(header), columnCount)
	}

	var out []Flight
	line := 1
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("parse: %s: read record: %w", path, err)
		}
		line++
		if len(rec) != columnCount {
			return nil, fmt.Errorf("parse: %s:%d: has %d columns, want %d", path, line, len(rec), columnCount)
		}

		flightNo := strings.TrimSpace(rec[0])
		if flightNo == "" {
			return nil, fmt.Errorf("parse: %s:%d: empty flightno", path, line)
		}
		origin := strings.ToUpper(strings.TrimSpace(rec[1]))
		dest := strings.ToUpper(strings.TrimSpace(rec[2]))
		if origin == "" || dest == "" {
			return nil, fmt.Errorf("parse: %s:%d: empty origin or dest", path, line)
		}

		schedDep, err := parseInt(rec[3])
		if err != nil {
			return nil, fmt.Errorf("parse: %s:%d: scheddep: %w", path, line, err)
		}
		actualDep, err := parseInt(rec[4])
		if err != nil {
			return nil, fmt.Errorf("parse: %s:%d: actualdep: %w", path, line, err)
		}
		delayMin, err := parseFloat(rec[5])
		if err != nil {
			return nil, fmt.Errorf("parse: %s:%d: delaymin: %w", path, line, err)
		}
		hour, err := parseFloat(rec[8])
		if err != nil {
			return nil, fmt.Errorf("parse: %s:%d: hour: %w", path, line, err)
		}

		out = append(out, Flight{
			FlightNo:  flightNo,
			Origin:    origin,
			Dest:      dest,
			Cause:     strings.ToUpper(strings.TrimSpace(rec[6])),
			Season:    strings.ToUpper(strings.TrimSpace(rec[7])),
			SchedDep:  schedDep,
			ActualDep: actualDep,
			DelayMin:  delayMin,
			Hour:      hour,
		})
	}
	return out, nil
}

func parseInt(s string) (int, error) {
	t := strings.TrimSpace(s)
	if t == "" {
		return 0, fmt.Errorf("empty value")
	}
	v, err := strconv.Atoi(t)
	if err != nil {
		return 0, fmt.Errorf("%q is not an integer", t)
	}
	return v, nil
}

func parseFloat(s string) (float64, error) {
	t := strings.TrimSpace(s)
	if t == "" {
		return 0, fmt.Errorf("empty value")
	}
	v, err := strconv.ParseFloat(t, 64)
	if err != nil {
		return 0, fmt.Errorf("%q is not a number", t)
	}
	return v, nil
}
