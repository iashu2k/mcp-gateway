package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
)

var ErrInvalidToolArguments = errors.New(
	"tool arguments do not satisfy input schema",
)

type SchemaValidator interface {
	Validate(
		schemaJSON json.RawMessage,
		argumentsJSON json.RawMessage,
	) error
}

type JSONSchemaValidator struct{}

func NewJSONSchemaValidator() *JSONSchemaValidator {
	return &JSONSchemaValidator{}
}

func (v *JSONSchemaValidator) Validate(
	schemaJSON json.RawMessage,
	argumentsJSON json.RawMessage,
) error {
	// v6: AddResource requires a decoded JSON value, not an io.Reader.
	// UnmarshalJSON preserves number precision via json.Number.
	schemaDocument, err := jsonschema.UnmarshalJSON(
		bytes.NewReader(schemaJSON),
	)
	if err != nil {
		return fmt.Errorf("decode stored input schema: %w", err)
	}

	arguments, err := jsonschema.UnmarshalJSON(
		bytes.NewReader(argumentsJSON),
	)
	if err != nil {
		return fmt.Errorf(
			"%w: arguments must be valid JSON",
			ErrInvalidToolArguments,
		)
	}

	if _, ok := arguments.(map[string]any); !ok {
		return fmt.Errorf(
			"%w: arguments must be a JSON object",
			ErrInvalidToolArguments,
		)
	}

	compiler := jsonschema.NewCompiler()

	const schemaURL = "https://mcp-gateway.local/schemas/tool-input.json"

	if err := compiler.AddResource(schemaURL, schemaDocument); err != nil {
		return fmt.Errorf("load tool input schema: %w", err)
	}

	schema, err := compiler.Compile(schemaURL)
	if err != nil {
		return fmt.Errorf("compile tool input schema: %w", err)
	}

	if err := schema.Validate(arguments); err != nil {
		return fmt.Errorf(
			"%w: %v",
			ErrInvalidToolArguments,
			err,
		)
	}

	return nil
}
