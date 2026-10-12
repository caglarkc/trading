package contracts

import "embed"

//go:embed decision-v1.schema.json
var Files embed.FS

func DecisionSchema() []byte {
	b, e := Files.ReadFile("decision-v1.schema.json")
	if e != nil {
		panic(e)
	}
	return b
}
