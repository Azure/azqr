package to

import (
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"
)

func String(i interface{}) string {
	if i == nil {
		return ""
	}

	switch v := i.(type) {
	case string:
		return v
	case int:
		return fmt.Sprintf("%d", v)
	case bool:
		return fmt.Sprintf("%t", v)
	default:
		jsonStr, err := json.Marshal(i)
		if err != nil {
			log.Fatal().Err(err).Msg("Unsupported type found in ARG query result")
		}
		return string(jsonStr)
	}
}
