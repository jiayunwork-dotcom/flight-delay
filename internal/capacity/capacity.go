// Package capacity evaluates runway demand, utilization and cascade risk.
package capacity

import (
	"math"
	"sort"
	"strings"

	"flight-delay/internal/parse"
)

// Runway describes a runway and its hourly movement capacity.
type Runway struct {
	Name            string
	CapacityPerHour float64
}

// PeakDemand counts departures from airport in the given clock hour.
// Matching is case-insensitive on the airport code; the hour is the
// truncated integer part of Flight.Hour.
func PeakDemand(flights []parse.Flight, airport string, hour int) int {
	code := strings.ToUpper(strings.TrimSpace(airport))
	if code == "" {
		return 0
	}
	n := 0
	for _, f := range flights {
		if f.Origin == code && int(math.Floor(f.Hour)) == hour {
			n++
		}
	}
	return n
}

// Utilization returns demand divided by capacity. A capacity <= 0 yields 0.
// Demand above capacity yields a value greater than 1.
func Utilization(demand int, capacity float64) float64 {
	if capacity <= 0 {
		return 0
	}
	return float64(demand) / capacity
}

// CascadeSeeds returns the flight numbers whose DelayMin is strictly
// greater than maxDelay, sorted in ascending order. When no flight
// exceeds the threshold the result is an empty slice.
func CascadeSeeds(flights []parse.Flight, maxDelay float64) []string {
	seeds := make([]string, 0)
	for _, f := range flights {
		if f.DelayMin > maxDelay {
			seeds = append(seeds, f.FlightNo)
		}
	}
	sort.Strings(seeds)
	return seeds
}

// Saturated reports whether the demand exceeds the runway capacity.
func (r Runway) Saturated(demand int) bool {
	return Utilization(demand, r.CapacityPerHour) > 1
}
