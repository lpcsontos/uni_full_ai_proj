package unitconv

import (
    "errors"
    "strings"
)

// Canonical units used by the system
const (
    UnitGram      = "g"
    UnitKilogram  = "kg"
    UnitMilliliter = "ml"
    UnitLiter     = "l"
    UnitTsp       = "tsp"
    UnitTbsp      = "tbsp"
)

var ErrIncompatibleUnits = errors.New("incompatible units")
var ErrUnknownUnit = errors.New("unknown unit")

// NormalizeUnit returns a canonical unit string for common synonyms.
func NormalizeUnit(u string) string {
    u = strings.TrimSpace(strings.ToLower(u))
    switch u {
    case "g", "gram", "grams":
        return UnitGram
    case "kg", "kilogram", "kilograms":
        return UnitKilogram
    case "ml", "milliliter", "milliliters":
        return UnitMilliliter
    case "l", "liter", "litre", "liters", "litres":
        return UnitLiter
    case "tsp", "teaspoon", "teaspoons":
        return UnitTsp
    case "tbsp", "tablespoon", "tablespoons":
        return UnitTbsp
    default:
        return u // return as-is; caller can decide unknown
    }
}

// Convert converts quantity from unit 'from' to unit 'to'.
// Returns converted quantity or an error if units are incompatible or unknown.
func Convert(q float64, from, to string) (float64, error) {
    f := NormalizeUnit(from)
    t := NormalizeUnit(to)

    // same unit
    if f == t {
        return q, nil
    }

    // grams <-> kilograms
    if (f == UnitGram || f == UnitKilogram) && (t == UnitGram || t == UnitKilogram) {
        // convert everything to grams then to target
        var grams float64
        if f == UnitGram {
            grams = q
        } else {
            grams = q * 1000.0
        }
        if t == UnitGram {
            return grams, nil
        }
        return grams / 1000.0, nil
    }

    // milliliters <-> liters
    if (f == UnitMilliliter || f == UnitLiter) && (t == UnitMilliliter || t == UnitLiter) {
        var ml float64
        if f == UnitMilliliter {
            ml = q
        } else {
            ml = q * 1000.0
        }
        if t == UnitMilliliter {
            return ml, nil
        }
        return ml / 1000.0, nil
    }

    // teaspoons <-> tablespoons
    if (f == UnitTsp || f == UnitTbsp) && (t == UnitTsp || t == UnitTbsp) {
        var tsp float64
        if f == UnitTsp {
            tsp = q
        } else {
            tsp = q * 3.0 // 1 tbsp = 3 tsp
        }
        if t == UnitTsp {
            return tsp, nil
        }
        return tsp / 3.0, nil
    }

    // If both are identical unknown strings, allow no-op
    if f == from && t == to && f == t {
        return q, nil
    }

    return 0, ErrIncompatibleUnits
}

// Merge sums two quantities with possibly different units and returns the result in preferredUnit.
// If preferredUnit is empty, the function returns the sum in the canonical unit of the first quantity.
func Merge(q1 float64, u1 string, q2 float64, u2 string, preferredUnit string) (float64, string, error) {
    u1c := NormalizeUnit(u1)
    u2c := NormalizeUnit(u2)

    // choose target unit
    target := strings.TrimSpace(preferredUnit)
    if target == "" {
        target = u1c
    } else {
        target = NormalizeUnit(target)
    }

    // convert both to target
    c1, err := Convert(q1, u1c, target)
    if err != nil {
        // try converting q2 to u1's unit instead if target failed
        if target != u1c {
            // fallback: try target = u1c
            target = u1c
            c1 = q1
            c2, err2 := Convert(q2, u2c, target)
            if err2 != nil {
                return 0, "", ErrIncompatibleUnits
            }
            return c1 + c2, target, nil
        }
        return 0, "", err
    }
    c2, err := Convert(q2, u2c, target)
    if err != nil {
        return 0, "", err
    }
    return c1 + c2, target, nil
}
