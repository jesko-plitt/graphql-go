package graphql_test

import (
	"encoding/json"
	"testing"

	"github.com/jesko-plitt/graphql-go"
	"github.com/jesko-plitt/graphql-go/testutil"
)

func TestIntrospection_SortingConsistency(t *testing.T) {
	// Create a schema with multiple types to test sorting
	testType := graphql.NewObject(graphql.ObjectConfig{
		Name: "TestType",
		Fields: graphql.Fields{
			"zebra": &graphql.Field{
				Type: graphql.String,
			},
			"apple": &graphql.Field{
				Type: graphql.String,
			},
			"banana": &graphql.Field{
				Type: graphql.String,
			},
		},
	})

	testEnum := graphql.NewEnum(graphql.EnumConfig{
		Name: "TestEnum",
		Values: graphql.EnumValueConfigMap{
			"ZEBRA":  &graphql.EnumValueConfig{Value: 0},
			"APPLE":  &graphql.EnumValueConfig{Value: 1},
			"BANANA": &graphql.EnumValueConfig{Value: 2},
		},
	})

	testInputObject := graphql.NewInputObject(graphql.InputObjectConfig{
		Name: "TestInputObject",
		Fields: graphql.InputObjectConfigFieldMap{
			"zebra":  &graphql.InputObjectFieldConfig{Type: graphql.String},
			"apple":  &graphql.InputObjectFieldConfig{Type: graphql.String},
			"banana": &graphql.InputObjectFieldConfig{Type: graphql.String},
		},
	})

	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: testType,
		Types: []graphql.Type{testEnum, testInputObject},
	})
	if err != nil {
		t.Fatalf("Error creating Schema: %v", err.Error())
	}

	// Run the introspection query multiple times to ensure consistency
	var results []string
	for range 5 {
		result := g(graphql.Params{
			Schema:        schema,
			RequestString: testutil.IntrospectionQuery,
		})

		if result.HasErrors() {
			t.Fatalf("Introspection query failed: %v", result.Errors)
		}

		// Convert result to JSON string for comparison
		jsonBytes, err := json.Marshal(result.Data)
		if err != nil {
			t.Fatalf("Failed to marshal result: %v", err)
		}
		results = append(results, string(jsonBytes))
	}

	// All results should be identical (same JSON string)
	for i := 1; i < len(results); i++ {
		if results[0] != results[i] {
			t.Errorf("Introspection results are not consistent. Result 0 != Result %d", i)
			t.Errorf("Result 0: %s", results[0])
			t.Errorf("Result %d: %s", i, results[i])
		}
	}
}

func TestIntrospection_FieldSorting(t *testing.T) {
	testType := graphql.NewObject(graphql.ObjectConfig{
		Name: "TestType",
		Fields: graphql.Fields{
			"zebra":  &graphql.Field{Type: graphql.String},
			"apple":  &graphql.Field{Type: graphql.String},
			"banana": &graphql.Field{Type: graphql.String},
		},
	})

	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: testType,
	})
	if err != nil {
		t.Fatalf("Error creating Schema: %v", err.Error())
	}

	query := `
		{
			__type(name: "TestType") {
				fields {
					name
				}
			}
		}
	`

	result := g(graphql.Params{
		Schema:        schema,
		RequestString: query,
	})

	if result.HasErrors() {
		t.Fatalf("Query failed: %v", result.Errors)
	}

	data := result.Data.(map[string]any)
	typeData := data["__type"].(map[string]any)
	fields := typeData["fields"].([]any)

	// Fields should be sorted alphabetically
	expectedNames := []string{"apple", "banana", "zebra"}
	if len(fields) != len(expectedNames) {
		t.Fatalf("Expected %d fields, got %d", len(expectedNames), len(fields))
	}

	for i, field := range fields {
		fieldData := field.(map[string]any)
		name := fieldData["name"].(string)
		if name != expectedNames[i] {
			t.Errorf("Field %d: expected %s, got %s", i, expectedNames[i], name)
		}
	}
}

func TestIntrospection_EnumValueSorting(t *testing.T) {
	testEnum := graphql.NewEnum(graphql.EnumConfig{
		Name: "TestEnum",
		Values: graphql.EnumValueConfigMap{
			"ZEBRA":  &graphql.EnumValueConfig{Value: 0},
			"APPLE":  &graphql.EnumValueConfig{Value: 1},
			"BANANA": &graphql.EnumValueConfig{Value: 2},
		},
	})

	testType := graphql.NewObject(graphql.ObjectConfig{
		Name: "TestType",
		Fields: graphql.Fields{
			"testEnum": &graphql.Field{Type: testEnum},
		},
	})

	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: testType,
	})
	if err != nil {
		t.Fatalf("Error creating Schema: %v", err.Error())
	}

	query := `
		{
			__type(name: "TestEnum") {
				enumValues {
					name
				}
			}
		}
	`

	result := g(graphql.Params{
		Schema:        schema,
		RequestString: query,
	})

	if result.HasErrors() {
		t.Fatalf("Query failed: %v", result.Errors)
	}

	data := result.Data.(map[string]any)
	typeData := data["__type"].(map[string]any)
	enumValues := typeData["enumValues"].([]any)

	// Enum values should be sorted alphabetically
	expectedNames := []string{"APPLE", "BANANA", "ZEBRA"}
	if len(enumValues) != len(expectedNames) {
		t.Fatalf("Expected %d enum values, got %d", len(expectedNames), len(enumValues))
	}

	for i, enumValue := range enumValues {
		enumValueData := enumValue.(map[string]any)
		name := enumValueData["name"].(string)
		if name != expectedNames[i] {
			t.Errorf("Enum value %d: expected %s, got %s", i, expectedNames[i], name)
		}
	}
}

func TestIntrospection_InputFieldSorting(t *testing.T) {
	testInputObject := graphql.NewInputObject(graphql.InputObjectConfig{
		Name: "TestInputObject",
		Fields: graphql.InputObjectConfigFieldMap{
			"zebra":  &graphql.InputObjectFieldConfig{Type: graphql.String},
			"apple":  &graphql.InputObjectFieldConfig{Type: graphql.String},
			"banana": &graphql.InputObjectFieldConfig{Type: graphql.String},
		},
	})

	testType := graphql.NewObject(graphql.ObjectConfig{
		Name: "TestType",
		Fields: graphql.Fields{
			"testInput": &graphql.Field{
				Type: graphql.String,
				Args: graphql.FieldConfigArgument{
					&graphql.ArgumentConfig{
						Name: "input",
						Type: testInputObject,
					},
				},
			},
		},
	})

	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: testType,
	})
	if err != nil {
		t.Fatalf("Error creating Schema: %v", err.Error())
	}

	query := `
		{
			__type(name: "TestInputObject") {
				inputFields {
					name
				}
			}
		}
	`

	result := g(graphql.Params{
		Schema:        schema,
		RequestString: query,
	})

	if result.HasErrors() {
		t.Fatalf("Query failed: %v", result.Errors)
	}

	data := result.Data.(map[string]any)
	typeData := data["__type"].(map[string]any)
	inputFields := typeData["inputFields"].([]any)

	// Input fields should be sorted alphabetically
	expectedNames := []string{"apple", "banana", "zebra"}
	if len(inputFields) != len(expectedNames) {
		t.Fatalf("Expected %d input fields, got %d", len(expectedNames), len(inputFields))
	}

	for i, inputField := range inputFields {
		inputFieldData := inputField.(map[string]any)
		name := inputFieldData["name"].(string)
		if name != expectedNames[i] {
			t.Errorf("Input field %d: expected %s, got %s", i, expectedNames[i], name)
		}
	}
}
