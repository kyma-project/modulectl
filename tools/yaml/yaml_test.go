package yaml_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/modulectl/tools/yaml"
)

type TestStruct struct {
	Name           string `comment:"required, the name of the module"                  yaml:"name"`
	Count          int    `comment:"required, the count of items"                      yaml:"count"`
	Active         bool   `comment:"optional, indicates if the module is active"       yaml:"active"`
	AdditionalInfo Nested `comment:"optional, additional information about the module" yaml:"additionalInfo"`
}

type Nested struct {
	NestedName  string       `comment:"required, the name of the nested structure"           yaml:"nestedName"`
	NestedCount int          `comment:"required, the count of items in the nested structure" yaml:"nestedCount"`
	NestedInner DoubleNested `comment:"optional, additional nested structure"                yaml:"nestedInner"`
}

type DoubleNested struct {
	DoubleNestedName  string `comment:"required, the name of the double nested structure"           yaml:"doubleNestedName"`
	DoubleNestedCount int    `comment:"required, the count of items in the double nested structure" yaml:"doubleNestedCount"`
}

func TestSerializeWithoutMarshalYAML(t *testing.T) {
	ts := TestStruct{
		Name:   "TestName",
		Count:  5,
		Active: true,
		AdditionalInfo: Nested{
			NestedName:  "TestNestedName",
			NestedCount: 30,
			NestedInner: DoubleNested{
				DoubleNestedName:  "TestDoubleNestedName",
				DoubleNestedCount: 200,
			},
		},
	}
	// Convert the struct to YAML
	conv := yaml.ObjectToYAMLConverter{}
	ymlData := conv.ConvertToYaml(ts)

	expectedYAML := slnl(`
name: "TestName" # required, the name of the module
count: 5 # required, the count of items
active: true # optional, indicates if the module is active
additionalInfo: # optional, additional information about the module
  nestedName: "TestNestedName" # required, the name of the nested structure
  nestedCount: 30 # required, the count of items in the nested structure
  nestedInner: # optional, additional nested structure
    doubleNestedName: "TestDoubleNestedName" # required, the name of the double nested structure
    doubleNestedCount: 200 # required, the count of items in the double nested structure
`)
	require.NotEmpty(t, ymlData, "YAML output should not be empty")
	assert.YAMLEq(t, expectedYAML, ymlData, "YAML output should match the expected format")
}

type TestStructWithMarshalYAML struct {
	Name           string            `comment:"required, the name of the module"                  yaml:"name"`
	Count          int               `comment:"required, the count of items"                      yaml:"count"`
	Active         bool              `comment:"optional, indicates if the module is active"       yaml:"active"`
	AdditionalInfo NestedWithMarshal `comment:"optional, additional information about the module" yaml:"additionalInfo"`
}

type NestedWithMarshal struct {
	NestedName  string                  `comment:"required, the name of the nested structure"           yaml:"nestedName"`
	NestedCount int                     `comment:"required, the count of items in the nested structure" yaml:"nestedCount"`
	NestedInner DoubleNestedWithMarshal `comment:"optional, additional nested structure"                yaml:"nestedInner"`
}

type DoubleNestedWithMarshal struct {
	DoubleNestedName  string `comment:"required, the name of the double nested structure"           yaml:"doubleNestedName"`
	DoubleNestedCount int    `comment:"required, the count of items in the double nested structure" yaml:"doubleNestedCount"`
}

func (d DoubleNestedWithMarshal) MarshalYAML() (any, error) {
	return fmt.Sprintf("%s-%d", d.DoubleNestedName, d.DoubleNestedCount), nil
}

func TestSerializeWithMarshalYAML(t *testing.T) {
	ts := TestStructWithMarshalYAML{
		Name:   "TestName",
		Count:  5,
		Active: true,
		AdditionalInfo: NestedWithMarshal{
			NestedName:  "TestNestedName",
			NestedCount: 30,
			NestedInner: DoubleNestedWithMarshal{
				DoubleNestedName:  "TestDoubleNestedName",
				DoubleNestedCount: 200,
			},
		},
	}

	// Convert the struct to YAML
	conv := yaml.ObjectToYAMLConverter{}
	ymlData := conv.ConvertToYaml(ts)

	expectedYAML := slnl(`
name: "TestName" # required, the name of the module
count: 5 # required, the count of items
active: true # optional, indicates if the module is active
additionalInfo: # optional, additional information about the module
  nestedName: "TestNestedName" # required, the name of the nested structure
  nestedCount: 30 # required, the count of items in the nested structure
  nestedInner: "TestDoubleNestedName-200" # optional, additional nested structure
`)

	require.NotEmpty(t, ymlData, "YAML output should not be empty")
	assert.YAMLEq(t, expectedYAML, ymlData, "YAML output should match the expected format")
}

// Helper function to (s)trip (l)eading (n)ew(l)ines.
func slnl(s string) string {
	return strings.TrimPrefix(s, "\n")
}
