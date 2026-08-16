// Command flight-delay reports flight delay statistics and airport
// runway capacity pressure from a flight record CSV file.
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"flight-delay/internal/capacity"
	"flight-delay/internal/parse"
	"flight-delay/internal/stats"
)

const topRouteLimit = 5

func usage(w *os.File) {
	fmt.Fprintf(w, "usage: flight-delay -flights <path> [-airport <code> -hour <int> -capacity <float> -maxdelay <float>]\n\n")
	fmt.Fprintf(w, "  -flights   path to a flight CSV file (required)\n")
	fmt.Fprintf(w, "  -airport   airport code for runway demand analysis\n")
	fmt.Fprintf(w, "  -hour      departure hour used with -airport (default 8)\n")
	fmt.Fprintf(w, "  -capacity  runway movements per hour (default 40)\n")
	fmt.Fprintf(w, "  -maxdelay  cascade-delay seed threshold in minutes (default 60)\n\n")
	fmt.Fprintf(w, "example: flight-delay -flights example/flights.csv -airport PEK -hour 8 -capacity 40\n")
}

func main() {
	fs := flag.NewFlagSet("flight-delay", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() { usage(os.Stderr) }

	flights := fs.String("flights", "", "path to a flight CSV file")
	airport := fs.String("airport", "", "airport code for runway demand analysis")
	hour := fs.Int("hour", 8, "departure hour used with -airport")
	capPerHour := fs.Float64("capacity", 40, "runway movements per hour")
	maxDelay := fs.Float64("maxdelay", 60, "cascade-delay seed threshold in minutes")

	if err := fs.Parse(os.Args[1:]); err != nil {
		usage(os.Stderr)
		os.Exit(2)
	}
	if *flights == "" {
		fmt.Fprintf(os.Stderr, "error: -flights is required\n\n")
		usage(os.Stderr)
		os.Exit(2)
	}

	if err := run(*flights, *airport, *hour, *capPerHour, *maxDelay, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(path, airport string, hour int, capPerHour, maxDelay float64, out *os.File) error {
	records, err := parse.ParseFlights(path)
	if err != nil {
		return err
	}
	if len(records) == 0 {
		return fmt.Errorf("%s contains no flight records", path)
	}

	dist := stats.DelayDist(records)
	fmt.Fprintf(out, "flights: %d\n", dist.N)
	fmt.Fprintf(out, "delay distribution: mean=%.2f median=%.2f min=%.2f max=%.2f\n",
		dist.Mean, dist.Median, dist.Min, dist.Max)

	routes := stats.ByRoute(records)
	fmt.Fprintf(out, "top routes by mean delay:\n")
	for _, key := range stats.TopRoutes(routes, topRouteLimit) {
		r := routes[key]
		fmt.Fprintf(out, "  %-10s mean=%7.2f count=%2d ontime=%.2f\n", key, r.Mean, r.Count, r.OnTimeRate)
	}

	hours := stats.ByHour(records)
	fmt.Fprintf(out, "by hour:\n")
	for _, h := range sortedIntKeys(hours) {
		s := hours[h]
		fmt.Fprintf(out, "  %02d:00 mean=%7.2f count=%2d\n", h, s.Mean, s.Count)
	}

	causes := stats.ByCause(records)
	fmt.Fprintf(out, "by cause:\n")
	for _, c := range sortedStringKeys(causes) {
		fmt.Fprintf(out, "  %-8s mean=%7.2f\n", c, causes[c])
	}

	fmt.Fprintf(out, "OTP(D15): %.4f\n", stats.OTP(records))
	fmt.Fprintf(out, "avg delay: %.2f\n", stats.AvgDelay(records))

	seeds := capacity.CascadeSeeds(records, maxDelay)
	fmt.Fprintf(out, "cascade seeds (>%.0f min): %d %v\n", maxDelay, len(seeds), seeds)

	if airport != "" {
		demand := capacity.PeakDemand(records, airport, hour)
		util := capacity.Utilization(demand, capPerHour)
		fmt.Fprintf(out, "airport %s hour %02d: peak demand=%d capacity=%.2f utilization=%.4f\n",
			airport, hour, demand, capPerHour, util)
		rw := capacity.Runway{Name: airport, CapacityPerHour: capPerHour}
		fmt.Fprintf(out, "saturated: %t\n", rw.Saturated(demand))
	}
	return nil
}

func sortedIntKeys(m map[int]stats.HourStat) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

func sortedStringKeys(m map[string]float64) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
