# flight-delay

`flight-delay` is a pure standard-library Go CLI for flight delay prediction
support and airport capacity analysis (题261 航班延误预测与机场运力分析).

It imports flight records from CSV and reports:

- the delay distribution (mean / median / min / max / count) over `delaymin`
- per-route breakdown (`Origin-Dest`) with mean delay, count and D15 on-time rate
- per-hour and per-cause breakdowns
- OTP(D15) — the fraction of flights delayed by no more than 15 minutes — and the average delay
- runway peak demand and utilization for a chosen airport and hour
- cascade-delay seeds: flights whose delay exceeds a threshold and can propagate downstream

## Layout

```
main.go                    CLI entry point
internal/parse             CSV loading and validation (Flight, ParseFlights)
internal/stats             delay aggregation (DelayDist, ByRoute, ByHour, ByCause, OTP, AvgDelay, TopRoutes)
internal/capacity          runway analysis (Runway, PeakDemand, Utilization, CascadeSeeds)
example/flights.csv        sample data set, 14 rows
```

## Input format

The CSV file must start with this exact header:

```
flightno,origin,dest,scheddep,actualdep,delaymin,cause,season,hour
```

`scheddep` and `actualdep` are integers (HHMM clock values), `delaymin` and
`hour` are numbers, `cause` is a label such as `WEATHER`, `CREW`, `ATC` or
`NONE`. A missing file, a wrong column count or a non-numeric numeric field is
reported as an error; the tool never panics on bad input.

## Usage

```
flight-delay -flights <path> [-airport <code> -hour <int> -capacity <float> -maxdelay <float>]
```

| Flag | Meaning | Default |
| --- | --- | --- |
| `-flights` | path to the flight CSV file (required) | — |
| `-airport` | airport code for runway demand analysis | empty (skipped) |
| `-hour` | departure hour used together with `-airport` | `8` |
| `-capacity` | runway movements per hour | `40` |
| `-maxdelay` | cascade-delay seed threshold in minutes | `60` |

### Examples

Full report including runway pressure at PEK during hour 08:

```
go run . -flights example/flights.csv -airport PEK -hour 8 -capacity 40
```

Statistics only, with a stricter cascade threshold:

```
go run . -flights example/flights.csv -maxdelay 45
```

## Exit codes

| Code | Condition |
| --- | --- |
| `0` | report printed successfully |
| `1` | input file missing, malformed or containing no records |
| `2` | `-flights` not supplied, or an unknown flag (usage goes to stderr) |

## Development

```
export GOTOOLCHAIN=local CGO_ENABLED=0
go vet ./...
go build ./...
go test ./...
```

## Docker

```
docker build -t flight-delay .
docker run --rm flight-delay /app/bin -flights example/flights.csv -airport PEK -hour 8
```

## License

MIT — see `LICENSE`.
