package utils

import "reflect"

func Types2Float64(i reflect.Value) float64 {
	switch i.Type().String() {
	case "int":
		return float64(i.Int())
	case "uint64":
		return float64(i.Uint())
	case "float64":
		return i.Float()
	case "bool":
		if i.Bool() {
			return 1.0
		}
	}
	return 0.0
}
