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
	Name           string `yaml:"name" comment:"required, the name of the module"`
	Count          int    `yaml:"count" comment:"required, the count of items"`
	Active         bool   `yaml:"active" comment:"optional, indicates if the module is active"`
	AdditionalInfo Nested `yaml:"additionalInfo" comment:"optional, additional information about the module"`
}

type Nested struct {
	NestedName  string       `yaml:"nestedName" comment:"required, the name of the nested structure"`
	NestedCount int          `yaml:"nestedCount" comment:"required, the count of items in the nested structure"`
	NestedInner DoubleNested `yaml:"nestedInner" comment:"optional, additional nested structure"`
}

type DoubleNested struct {
	DoubleNestedName  string `yaml:"doubleNestedName" comment:"required, the name of the double nested structure"`
	DoubleNestedCount int    `yaml:"doubleNestedCount" comment:"required, the count of items in the double nested structure"`
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
	Name           string            `yaml:"name" comment:"required, the name of the module"`
	Count          int               `yaml:"count" comment:"required, the count of items"`
	Active         bool              `yaml:"active" comment:"optional, indicates if the module is active"`
	AdditionalInfo NestedWithMarshal `yaml:"additionalInfo" comment:"optional, additional information about the module"`
}

type NestedWithMarshal struct {
	NestedName  string                  `yaml:"nestedName" comment:"required, the name of the nested structure"`
	NestedCount int                     `yaml:"nestedCount" comment:"required, the count of items in the nested structure"`
	NestedInner DoubleNestedWithMarshal `yaml:"nestedInner" comment:"optional, additional nested structure"`
}

type DoubleNestedWithMarshal struct {
	DoubleNestedName  string `yaml:"doubleNestedName" comment:"required, the name of the double nested structure"`
	DoubleNestedCount int    `yaml:"doubleNestedCount" comment:"required, the count of items in the double nested structure"`
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

// Helper function to (s)trip (l)eading (n)ew(l)ines
func slnl(s string) string {
	return strings.TrimPrefix(s, "\n")
}
