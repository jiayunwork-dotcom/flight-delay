package capacity

import (
	"reflect"
	"testing"

	"flight-delay/internal/parse"
)

func sample() []parse.Flight {
	return []parse.Flight{
		{FlightNo: "CA9", Origin: "PEK", Dest: "SHA", DelayMin: 5, Hour: 8},
		{FlightNo: "CA2", Origin: "PEK", Dest: "CAN", DelayMin: 70, Hour: 8.5},
		{FlightNo: "CA1", Origin: "PEK", Dest: "SZX", DelayMin: 90, Hour: 8.99},
		{FlightNo: "MU7", Origin: "SHA", Dest: "PEK", DelayMin: 120, Hour: 8},
		{FlightNo: "CZ3", Origin: "PEK", Dest: "CTU", DelayMin: 20, Hour: 9},
	}
}

// TestPeakDemand (category: other) expects only departures whose origin matches
// the airport (case-insensitively) and whose truncated hour matches to be
// counted, and expects 0 for nil input, an unknown airport, an empty code or an
// hour with no departures.
func TestPeakDemand(t *testing.T) {
	flights := sample()
	if got := PeakDemand(flights, "PEK", 8); got != 3 {
		t.Errorf("PeakDemand(PEK, 8) = %d, want 3", got)
	}
	if got := PeakDemand(flights, " pek ", 8); got != 3 {
		t.Errorf("PeakDemand(' pek ', 8) = %d, want 3", got)
	}
	if got := PeakDemand(flights, "PEK", 9); got != 1 {
		t.Errorf("PeakDemand(PEK, 9) = %d, want 1", got)
	}
	if got := PeakDemand(flights, "SHA", 8); got != 1 {
		t.Errorf("PeakDemand(SHA, 8) = %d, want 1", got)
	}
	if got := PeakDemand(flights, "CAN", 8); got != 0 {
		t.Errorf("PeakDemand(CAN, 8) = %d, want 0", got)
	}
	if got := PeakDemand(flights, "", 8); got != 0 {
		t.Errorf("PeakDemand(empty code, 8) = %d, want 0", got)
	}
	if got := PeakDemand(nil, "PEK", 8); got != 0 {
		t.Errorf("PeakDemand(nil, PEK, 8) = %d, want 0", got)
	}
}

// TestUtilization (category: other) expects demand/capacity, a value above 1.0
// when demand exceeds capacity, exactly 1.0 when they are equal, and 0 when the
// capacity is zero or negative.
func TestUtilization(t *testing.T) {
	if got := Utilization(20, 40); got != 0.5 {
		t.Errorf("Utilization(20, 40) = %v, want 0.5", got)
	}
	if got := Utilization(40, 40); got != 1.0 {
		t.Errorf("Utilization(40, 40) = %v, want 1.0", got)
	}
	if got := Utilization(50, 40); got <= 1.0 {
		t.Errorf("Utilization(50, 40) = %v, want > 1.0", got)
	}
	if got := Utilization(10, 0); got != 0 {
		t.Errorf("Utilization(10, 0) = %v, want 0", got)
	}
	if got := Utilization(10, -5); got != 0 {
		t.Errorf("Utilization(10, -5) = %v, want 0", got)
	}
	if got := Utilization(0, 40); got != 0 {
		t.Errorf("Utilization(0, 40) = %v, want 0", got)
	}

	rw := Runway{Name: "PEK-18R", CapacityPerHour: 40}
	if rw.Saturated(40) {
		t.Error("Saturated(40) with capacity 40 = true, want false")
	}
	if !rw.Saturated(41) {
		t.Error("Saturated(41) with capacity 40 = false, want true")
	}
	if (Runway{Name: "X"}).Saturated(100) {
		t.Error("Saturated with zero capacity = true, want false")
	}
}

// TestCascadeSeeds (category: slice) expects the flight numbers with DelayMin
// strictly greater than the threshold in ascending order, an empty non-nil
// slice when no flight exceeds the threshold or the input is nil, and no
// mutation of the caller's slice order.
func TestCascadeSeeds(t *testing.T) {
	flights := sample()
	before := make([]string, len(flights))
	for i, f := range flights {
		before[i] = f.FlightNo
	}

	want := []string{"CA1", "CA2", "MU7"}
	if got := CascadeSeeds(flights, 60); !reflect.DeepEqual(got, want) {
		t.Errorf("CascadeSeeds(flights, 60) = %v, want %v", got, want)
	}

	// The threshold is exclusive: a delay equal to it is not a seed.
	if got := CascadeSeeds(flights, 120); len(got) != 0 {
		t.Errorf("CascadeSeeds(flights, 120) = %v, want empty", got)
	}

	none := CascadeSeeds(flights, 1000)
	if none == nil {
		t.Fatal("CascadeSeeds(flights, 1000) = nil, want empty slice")
	}
	if len(none) != 0 {
		t.Errorf("CascadeSeeds(flights, 1000) = %v, want empty", none)
	}

	nilIn := CascadeSeeds(nil, 10)
	if nilIn == nil || len(nilIn) != 0 {
		t.Errorf("CascadeSeeds(nil, 10) = %v, want empty slice", nilIn)
	}

	all := CascadeSeeds(flights, 0)
	if len(all) != len(flights) {
		t.Errorf("len(CascadeSeeds(flights, 0)) = %d, want %d", len(all), len(flights))
	}
	for i := 1; i < len(all); i++ {
		if all[i-1] > all[i] {
			t.Fatalf("CascadeSeeds result not sorted: %v", all)
		}
	}

	for i, f := range flights {
		if f.FlightNo != before[i] {
			t.Fatalf("input slice was reordered at %d: got %s, want %s", i, f.FlightNo, before[i])
		}
	}
}
