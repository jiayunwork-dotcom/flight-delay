package stats

import (
	"reflect"
	"testing"

	"flight-delay/internal/parse"
)

func sample() []parse.Flight {
	return []parse.Flight{
		{FlightNo: "CA1", Origin: "PEK", Dest: "SHA", Cause: "NONE", DelayMin: 0, Hour: 8},
		{FlightNo: "CA2", Origin: "PEK", Dest: "SHA", Cause: "WEATHER", DelayMin: 30, Hour: 8},
		{FlightNo: "CA3", Origin: "PEK", Dest: "CAN", Cause: "ATC", DelayMin: 10, Hour: 9},
		{FlightNo: "CA4", Origin: "SHA", Dest: "PEK", Cause: "WEATHER", DelayMin: 60, Hour: 9.75},
	}
}

// TestDelayDist (category: nil) expects a nil or empty slice to produce a
// zero-valued Hist, and expects mean, median, min and max over DelayMin for
// both odd and even sample sizes.
func TestDelayDist(t *testing.T) {
	if got := DelayDist(nil); got != (Hist{}) {
		t.Errorf("DelayDist(nil) = %+v, want zero Hist", got)
	}
	if got := DelayDist([]parse.Flight{}); got.N != 0 {
		t.Errorf("DelayDist(empty).N = %d, want 0", got.N)
	}

	// Even count: 0,10,30,60 -> mean 25, median (10+30)/2 = 20.
	want := Hist{Mean: 25, Median: 20, Max: 60, Min: 0, N: 4}
	if got := DelayDist(sample()); got != want {
		t.Errorf("DelayDist(sample) = %+v, want %+v", got, want)
	}

	// Odd count: 0,10,30 -> mean 40/3, median 10.
	odd := sample()[:3]
	got := DelayDist(odd)
	if got.N != 3 || got.Median != 10 || got.Min != 0 || got.Max != 30 {
		t.Errorf("DelayDist(odd) = %+v, want N=3 Median=10 Min=0 Max=30", got)
	}
	if diff := got.Mean - 40.0/3.0; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("DelayDist(odd).Mean = %v, want %v", got.Mean, 40.0/3.0)
	}
}

// TestByRoute (category: slice) expects flights grouped by "Origin-Dest" with
// per-route mean delay, count and D15 on-time rate, an empty non-nil map for
// nil input, and a deterministic TopRoutes ordering by descending mean delay.
func TestByRoute(t *testing.T) {
	empty := ByRoute(nil)
	if empty == nil {
		t.Fatal("ByRoute(nil) = nil, want empty map")
	}
	if len(empty) != 0 {
		t.Errorf("len(ByRoute(nil)) = %d, want 0", len(empty))
	}

	routes := ByRoute(sample())
	if len(routes) != 3 {
		t.Fatalf("len(routes) = %d, want 3", len(routes))
	}
	pekSha := routes["PEK-SHA"]
	if pekSha.Count != 2 || pekSha.Mean != 15 || pekSha.OnTimeRate != 0.5 {
		t.Errorf("routes[PEK-SHA] = %+v, want Mean=15 Count=2 OnTimeRate=0.5", pekSha)
	}
	pekCan := routes["PEK-CAN"]
	if pekCan.Count != 1 || pekCan.Mean != 10 || pekCan.OnTimeRate != 1 {
		t.Errorf("routes[PEK-CAN] = %+v, want Mean=10 Count=1 OnTimeRate=1", pekCan)
	}
	shaPek := routes["SHA-PEK"]
	if shaPek.OnTimeRate != 0 {
		t.Errorf("routes[SHA-PEK].OnTimeRate = %v, want 0", shaPek.OnTimeRate)
	}
	if _, ok := routes["CAN-PEK"]; ok {
		t.Error("routes contains CAN-PEK, want absent")
	}

	wantTop := []string{"SHA-PEK", "PEK-SHA"}
	if got := TopRoutes(routes, 2); !reflect.DeepEqual(got, wantTop) {
		t.Errorf("TopRoutes(routes, 2) = %v, want %v", got, wantTop)
	}
	if got := TopRoutes(routes, 10); len(got) != 3 {
		t.Errorf("len(TopRoutes(routes, 10)) = %d, want 3", len(got))
	}
	if got := TopRoutes(nil, 3); len(got) != 0 {
		t.Errorf("TopRoutes(nil, 3) = %v, want empty", got)
	}
}

// TestByHourAndByCause (category: other) expects hour buckets to use the
// truncated integer hour and cause buckets to report the mean delay per cause.
func TestByHourAndByCause(t *testing.T) {
	hours := ByHour(sample())
	if len(hours) != 2 {
		t.Fatalf("len(hours) = %d, want 2", len(hours))
	}
	if got := hours[8]; got.Count != 2 || got.Mean != 15 {
		t.Errorf("hours[8] = %+v, want Mean=15 Count=2", got)
	}
	// Hour 9.75 truncates into bucket 9 together with hour 9.
	if got := hours[9]; got.Count != 2 || got.Mean != 35 {
		t.Errorf("hours[9] = %+v, want Mean=35 Count=2", got)
	}
	if len(ByHour(nil)) != 0 {
		t.Error("ByHour(nil) is not empty")
	}

	causes := ByCause(sample())
	if got := causes["WEATHER"]; got != 45 {
		t.Errorf("causes[WEATHER] = %v, want 45", got)
	}
	if got := causes["ATC"]; got != 10 {
		t.Errorf("causes[ATC] = %v, want 10", got)
	}
	if got := causes["NONE"]; got != 0 {
		t.Errorf("causes[NONE] = %v, want 0", got)
	}
	if _, ok := causes["CREW"]; ok {
		t.Error("causes contains CREW, want absent")
	}
}

// TestOTP (category: other) expects the D15 fraction: 1.0 when every delay is
// at most 15 minutes, 0.0 when every delay exceeds 15, 0 for empty input, and
// the exact fraction for a mixed sample. AvgDelay follows the same emptiness rule.
func TestOTP(t *testing.T) {
	if got := OTP(nil); got != 0 {
		t.Errorf("OTP(nil) = %v, want 0", got)
	}
	if got := AvgDelay(nil); got != 0 {
		t.Errorf("AvgDelay(nil) = %v, want 0", got)
	}

	allOnTime := []parse.Flight{{DelayMin: 0}, {DelayMin: 15}, {DelayMin: 3}}
	if got := OTP(allOnTime); got != 1.0 {
		t.Errorf("OTP(all <= 15) = %v, want 1.0", got)
	}

	allLate := []parse.Flight{{DelayMin: 15.5}, {DelayMin: 90}}
	if got := OTP(allLate); got != 0.0 {
		t.Errorf("OTP(all > 15) = %v, want 0.0", got)
	}

	// Sample: 0 and 10 are on time, 30 and 60 are not -> 2/4.
	if got := OTP(sample()); got != 0.5 {
		t.Errorf("OTP(sample) = %v, want 0.5", got)
	}
	if got := AvgDelay(sample()); got != 25 {
		t.Errorf("AvgDelay(sample) = %v, want 25", got)
	}
}
