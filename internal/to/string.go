package to

import (
	"encoding/json"
	"strconv"

	"github.com/rs/zerolog/log"
)

// String converts an interface{} to string efficiently.
func String(i interface{}) string {
	if i == nil {
		return ""
	}

	switch v := i.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case bool:
		return strconv.FormatBool(v)
	default:
		jsonStr, err := json.Marshal(i)
		if err != nil {
			log.Fatal().Err(err).Msg("Unsupported type found in ARG query result")
		}
		return string(jsonStr)
	}
}

// Float converts an interface{} to float64. Returns 0 for nil or unsupported types.
func Float(i interface{}) float64 {
	if i == nil {
		return 0
	}
	switch v := i.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0
		}
		return f
	}
	return 0
}

// Int converts an interface{} to int. Returns 0 for nil or unsupported types.
func Int(i interface{}) int {
	if i == nil {
		return 0
	}
	switch v := i.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float64:
		return int(v)
	case float32:
		return int(v)
	}
	return 0
}
