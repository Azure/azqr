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
