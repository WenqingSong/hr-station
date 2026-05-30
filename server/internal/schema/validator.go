package schema

import (
	"encoding/json"
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

// ValidationError describes a single Schema validation failure.
type ValidationError struct {
	// Path is the JSON pointer path where the error occurred (e.g. "#/match_score").
	Path string
	// Message is a human-readable description of the validation failure.
	Message string
}

// ValidationResult holds the outcome of a Schema validation.
type ValidationResult struct {
	// Success indicates whether the data conforms to the Schema.
	Success bool
	// SchemaName is the logical name of the Schema used for validation.
	SchemaName SchemaName
	// ErrorCode is set when validation fails (SCHEMA_VALIDATION_FAILED).
	ErrorCode string
	// ErrorMessage is a summary message when validation fails.
	ErrorMessage string
	// Errors lists the individual validation failures.
	Errors []ValidationError
}

// Validator validates data against JSON Schemas loaded by Loader.
type Validator struct{}

// NewValidator creates a new Validator.
func NewValidator() *Validator {
	return &Validator{}
}

// Validate validates data against the given Schema document.
// data must be a value that can be marshalled to JSON (e.g. map[string]any).
// It returns a ValidationResult with Success=true if the data conforms,
// or a ValidationResult with detailed errors if it does not.
func (v *Validator) Validate(doc *Document, data any) (*ValidationResult, error) {
	// Compile the schema from the document content.
	schema, err := jsonschema.CompileString(
		string(doc.Name)+"_schema.json",
		doc.Content,
	)
	if err != nil {
		return nil, fmt.Errorf("schema: failed to compile %s: %w", doc.Name, err)
	}

	// Validate the data.
	err = schema.Validate(data)
	if err == nil {
		return &ValidationResult{
			Success:    true,
			SchemaName: doc.Name,
		}, nil
	}

	// Collect validation errors.
	var errs []ValidationError
	collectErrors(err, &errs)

	errorMsg := "schema validation failed"
	if len(errs) > 0 {
		errorMsg = fmt.Sprintf("schema validation failed: %d error(s)", len(errs))
	}

	return &ValidationResult{
		Success:      false,
		SchemaName:   doc.Name,
		ErrorCode:    "SCHEMA_VALIDATION_FAILED",
		ErrorMessage: errorMsg,
		Errors:       errs,
	}, nil
}

// ValidateJSON validates JSON bytes against the given Schema document.
// It unmarshals the bytes first, then calls Validate.
func (v *Validator) ValidateJSON(doc *Document, jsonBytes []byte) (*ValidationResult, error) {
	var data any
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		return &ValidationResult{
			Success:      false,
			SchemaName:   doc.Name,
			ErrorCode:    "JSON_PARSE_FAILED",
			ErrorMessage: fmt.Sprintf("failed to parse JSON: %s", err.Error()),
		}, nil
	}

	return v.Validate(doc, data)
}

// collectErrors recursively extracts ValidationError entries from a
// jsonschema.ValidationError tree.
func collectErrors(err error, out *[]ValidationError) {
	if err == nil {
		return
	}

	ve, ok := err.(*jsonschema.ValidationError)
	if !ok {
		*out = append(*out, ValidationError{
			Path:    "/",
			Message: err.Error(),
		})
		return
	}

	// If this node has no sub-causes, it is a leaf error.
	if len(ve.Causes) == 0 {
		*out = append(*out, ValidationError{
			Path:    ve.InstanceLocation,
			Message: ve.Message,
		})
		return
	}

	// Otherwise recurse into causes.
	for _, cause := range ve.Causes {
		collectErrors(cause, out)
	}
}
