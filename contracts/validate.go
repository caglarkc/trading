package contracts

import (
	"encoding/json"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

func Validate(schema, body []byte) error {
	var doc, v any
	if e := json.Unmarshal(schema, &doc); e != nil {
		return e
	}
	if e := json.Unmarshal(body, &v); e != nil {
		return e
	}
	c := jsonschema.NewCompiler()
	const id = "https://trading.invalid/contracts/schema.json"
	if e := c.AddResource(id, doc); e != nil {
		return e
	}
	s, e := c.Compile(id)
	if e != nil {
		return e
	}
	return s.Validate(v)
}
