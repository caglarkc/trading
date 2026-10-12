package contracts

import "embed"

//go:embed *.json
var Files embed.FS

func DecisionSchema() []byte {
	b, e := Files.ReadFile("decision-v1.schema.json")
	if e != nil {
		panic(e)
	}
	return b
}

func ResearchSchema() []byte {
	b, e := Files.ReadFile("research-v1.schema.json")
	if e != nil {
		panic(e)
	}
	return b
}
func ReceiptSchema() []byte {
	b, e := Files.ReadFile("receipt-v1.schema.json")
	if e != nil {
		panic(e)
	}
	return b
}
func DynamicSchema() []byte {
	b, e := Files.ReadFile("dynamic-v1.schema.json")
	if e != nil {
		panic(e)
	}
	return b
}
