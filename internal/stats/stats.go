// Package stats aggregates delay statistics over flight records.
package stats

import (
	"math"
	"sort"

	"flight-delay/internal/parse"
)

// OnTimeThreshold is the D15 on-time cut-off in minutes.
const OnTimeThreshold = 15.0

// Hist summarises a delay distribution.
type Hist struct {
	Mean   float64
	Median float64
	Max    float64
	Min    float64
	N      int
}

// RouteStat summarises one origin-destination pair.
type RouteStat struct {
	Mean       float64
	Count      int
	OnTimeRate float64
}

// HourStat summarises one departure hour bucket.
type HourStat struct {
	Mean  float64
	Count int
}

// DelayDist returns the distribution of DelayMin values.
// A nil or empty slice yields Hist{N: 0} with all fields zero.
func DelayDist(flights []parse.Flight) Hist {
	if len(flights) == 0 {
		return Hist{N: 0}
	}
	values := make([]float64, 0, len(flights))
	sum := 0.0
	for _, f := range flights {
		values = append(values, f.DelayMin)
		sum += f.DelayMin
	}
	sort.Float64s(values)
	n := len(values)
	median := values[n/2]
	if n%2 == 0 {
		median = (values[n/2-1] + values[n/2]) / 2
	}
	return Hist{
		Mean:   sum / float64(n),
		Median: median,
		Max:    values[n-1],
		Min:    values[0],
		N:      n,
	}
}

// ByRoute groups flights by "Origin-Dest" and reports the mean delay,
// the flight count and the fraction of flights with DelayMin <= 15.
// A nil or empty slice yields an empty (non-nil) map.
func ByRoute(flights []parse.Flight) map[string]RouteStat {
	sums := make(map[string]float64)
	counts := make(map[string]int)
	onTime := make(map[string]int)
	for _, f := range flights {
		key := f.Route()
		sums[key] += f.DelayMin
		counts[key]++
		if f.DelayMin <= OnTimeThreshold {
			onTime[key]++
		}
	}
	out := make(map[string]RouteStat, len(counts))
	for key, c := range counts {
		out[key] = RouteStat{
			Mean:       sums[key] / float64(c),
			Count:      c,
			OnTimeRate: float64(onTime[key]) / float64(c),
		}
	}
	return out
}

// ByHour groups flights by their integer departure hour and reports the
// mean delay and flight count of every bucket.
func ByHour(flights []parse.Flight) map[int]HourStat {
	sums := make(map[int]float64)
	counts := make(map[int]int)
	for _, f := range flights {
		h := int(math.Floor(f.Hour))
		sums[h] += f.DelayMin
		counts[h]++
	}
	out := make(map[int]HourStat, len(counts))
	for h, c := range counts {
		out[h] = HourStat{Mean: sums[h] / float64(c), Count: c}
	}
	return out
}

// ByCause returns the mean DelayMin per delay cause.
func ByCause(flights []parse.Flight) map[string]float64 {
	sums := make(map[string]float64)
	counts := make(map[string]int)
	for _, f := range flights {
		sums[f.Cause] += f.DelayMin
		counts[f.Cause]++
	}
	out := make(map[string]float64, len(counts))
	for cause, c := range counts {
		out[cause] = sums[cause] / float64(c)
	}
	return out
}

// OTP returns the D15 on-time performance: the fraction of flights with
// DelayMin <= 15. A nil or empty slice yields 0.
func OTP(flights []parse.Flight) float64 {
	if len(flights) == 0 {
		return 0
	}
	onTime := 0
	for _, f := range flights {
		if f.DelayMin <= OnTimeThreshold {
			onTime++
		}
	}
	return float64(onTime) / float64(len(flights))
}

// AvgDelay returns the mean DelayMin. A nil or empty slice yields 0.
func AvgDelay(flights []parse.Flight) float64 {
	if len(flights) == 0 {
		return 0
	}
	sum := 0.0
	for _, f := range flights {
		sum += f.DelayMin
	}
	return sum / float64(len(flights))
}

// TopRoutes returns route keys ordered by descending mean delay, ties
// broken by ascending key, limited to at most n entries.
func TopRoutes(routes map[string]RouteStat, n int) []string {
	keys := make([]string, 0, len(routes))
	for k := range routes {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		a, b := routes[keys[i]], routes[keys[j]]
		if a.Mean != b.Mean {
			return a.Mean > b.Mean
		}
		return keys[i] < keys[j]
	})
	if n >= 0 && n < len(keys) {
		keys = keys[:n]
	}
	return keys
}
