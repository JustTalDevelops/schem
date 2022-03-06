package schem

import (
	_ "embed"
	"encoding/json"
	"github.com/tidwall/gjson"
)

// bedrockState contains information about a Bedrock edition state.
type bedrockState struct {
	// Name is the Bedrock edition name.
	Name string `json:"bedrock_identifier"`
	// Properties is the Bedrock edition properties.
	Properties map[string]interface{} `json:"bedrock_states"`
}

var (
	//go:embed blocks.json
	mappingsData []byte
	// editionConversion maps between a Java encoded state and a Bedrock state.
	editionConversion = make(map[string]bedrockState)
)

func init() {
	parsedData := gjson.ParseBytes(mappingsData)
	parsedData.ForEach(func(key, value gjson.Result) bool {
		var state bedrockState
		err := json.Unmarshal([]byte(value.String()), &state)
		if err != nil {
			panic(err)
		}

		// Fix the values.
		for k, v := range state.Properties {
			if v, ok := v.(float64); ok {
				state.Properties[k] = int32(v)
			}
		}

		editionConversion[key.String()] = state
		return true
	})
}
