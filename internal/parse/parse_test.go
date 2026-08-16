package parse

import (
	"os"
	"path/filepath"
	"testing"
)

const header = "flightno,origin,dest,scheddep,actualdep,delaymin,cause,season,hour\n"

func writeCSV(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "flights.csv")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return path
}

// TestParseFlights (category: error) expects a valid file to be parsed into
// normalized records, and expects an error for a missing file, an empty file,
// a bad header, a short row and a non-numeric delaymin or hour.
func TestParseFlights(t *testing.T) {
	good := writeCSV(t, header+
		"CA1501, pek , sha ,800,805,5,none,spring,8\n"+
		"CA1502,PEK,CAN,810,850,40.5,WEATHER,SPRING,8\n")

	flights, err := ParseFlights(good)
	if err != nil {
		t.Fatalf("ParseFlights(valid) error = %v, want nil", err)
	}
	if len(flights) != 2 {
		t.Fatalf("len(flights) = %d, want 2", len(flights))
	}
	first := flights[0]
	if first.FlightNo != "CA1501" || first.Origin != "PEK" || first.Dest != "SHA" {
		t.Errorf("first record = %+v, want FlightNo CA1501 Origin PEK Dest SHA", first)
	}
	if first.Cause != "NONE" || first.Season != "SPRING" {
		t.Errorf("first record cause/season = %q/%q, want NONE/SPRING", first.Cause, first.Season)
	}
	if first.SchedDep != 800 || first.ActualDep != 805 {
		t.Errorf("first record sched/actual = %d/%d, want 800/805", first.SchedDep, first.ActualDep)
	}
	if first.DelayMin != 5 || first.Hour != 8 {
		t.Errorf("first record delay/hour = %v/%v, want 5/8", first.DelayMin, first.Hour)
	}
	if got := first.Route(); got != "PEK-SHA" {
		t.Errorf("Route() = %q, want %q", got, "PEK-SHA")
	}
	if flights[1].DelayMin != 40.5 {
		t.Errorf("second record delay = %v, want 40.5", flights[1].DelayMin)
	}

	bad := map[string]string{
		"empty file":        "",
		"header too short":  "flightno,origin,dest\nCA1,PEK,SHA\n",
		"row too short":     header + "CA1501,PEK,SHA,800,805,5\n",
		"non-numeric delay": header + "CA1501,PEK,SHA,800,805,late,NONE,SPRING,8\n",
		"non-numeric hour":  header + "CA1501,PEK,SHA,800,805,5,NONE,SPRING,morning\n",
		"non-numeric sched": header + "CA1501,PEK,SHA,08:00,805,5,NONE,SPRING,8\n",
		"empty flightno":    header + ",PEK,SHA,800,805,5,NONE,SPRING,8\n",
		"empty origin":      header + "CA1501,,SHA,800,805,5,NONE,SPRING,8\n",
	}
	for name, body := range bad {
		path := writeCSV(t, body)
		got, err := ParseFlights(path)
		if err == nil {
			t.Errorf("ParseFlights(%s) error = nil, want error", name)
		}
		if got != nil {
			t.Errorf("ParseFlights(%s) records = %v, want nil", name, got)
		}
	}

	missing := filepath.Join(t.TempDir(), "does-not-exist.csv")
	if _, err := ParseFlights(missing); err == nil {
		t.Error("ParseFlights(missing file) error = nil, want error")
	}
}

// TestParseFlightsDeferredRemoval (category: defer) expects a file removed by a
// deferred cleanup to be readable before the cleanup and to fail afterwards,
// confirming ParseFlights reports missing files as errors instead of panicking.
func TestParseFlightsDeferredRemoval(t *testing.T) {
	path := writeCSV(t, header+"CA1501,PEK,SHA,800,805,5,NONE,SPRING,8\n")

	removed := false
	func() {
		defer func() {
			if err := os.Remove(path); err != nil {
				t.Fatalf("remove fixture: %v", err)
			}
			removed = true
		}()
		flights, err := ParseFlights(path)
		if err != nil {
			t.Fatalf("ParseFlights before removal error = %v, want nil", err)
		}
		if len(flights) != 1 {
			t.Fatalf("len(flights) = %d, want 1", len(flights))
		}
	}()

	if !removed {
		t.Fatal("deferred cleanup did not run")
	}
	if _, err := ParseFlights(path); err == nil {
		t.Error("ParseFlights after removal error = nil, want error")
	}
}
