package yaml

import (
	"fmt"
	"reflect"
	"strings"
)

type ObjectToYAMLConverter struct{}

func (*ObjectToYAMLConverter) ConvertToYaml(obj interface{}) string {
	reflectValue := reflect.ValueOf(obj)
	var yamlBuilder strings.Builder
	generateYamlWithComments(&yamlBuilder, reflectValue, 0, "")
	return yamlBuilder.String()
}

// generateYamlWithComments uses a "comment" tag in the struct definition to generate YAML with comments on corresponding lines.
// Note 1: Map support is missing!
// Note 2: There is very basic support for structs that implement the MarshalYAML() method, it should directly return a string. More complex scenarios are not implemented fully yet.
func generateYamlWithComments(yamlBuilder *strings.Builder, obj reflect.Value, indentLevel int, commentPrefix string) { //nolint: gocognit //yes, it's too big - refactoring would require introducing a fully-fledged YAML serializer, which is not the goal of this code.
	objType := obj.Type()

	indentPrefix := strings.Repeat("  ", indentLevel)
	originalCommentPrefix := commentPrefix

	for i := range objType.NumField() {
		commentPrefix = originalCommentPrefix
		field := objType.Field(i)
		value := obj.Field(i)
		yamlTag := field.Tag.Get("yaml")
		commentTag := field.Tag.Get("comment")

		// comment-out non-required empty attributes
		if value.IsZero() && !strings.Contains(commentTag, "required") {
			commentPrefix = "# "
		}

		if value.Kind() == reflect.Struct {
			// Check if there is a MarshalYAML method defined for this struct
			marshalRes := tryMarshalYAML(value)
			if marshalRes != nil {
				// MarshalYAML method returned a value, try to print it if it is a string
				if tryPrintIfString(marshalRes, yamlBuilder, commentTag, commentPrefix, indentPrefix, yamlTag) {
					continue
				}
				// MarshalYAML method returned something that is NOT a string... Process it recursively (experimental)
				generateYamlWithComments(yamlBuilder, reflect.ValueOf(marshalRes), indentLevel+1, commentPrefix)
				continue
			}

			// Serialize the struct entry in YAML output (no value here, just the name and tags, if any)
			if commentTag == "" {
				yamlBuilder.WriteString(fmt.Sprintf("%s%s%s: \n", commentPrefix, indentPrefix, yamlTag))
			} else {
				yamlBuilder.WriteString(fmt.Sprintf("%s%s%s: # %s\n", commentPrefix, indentPrefix, yamlTag, commentTag))
			}

			// Recursively serialize nested struct fields
			generateYamlWithComments(yamlBuilder, value, indentLevel+1, commentPrefix)
			continue
		}

		if value.Kind() == reflect.Slice {
			if commentTag == "" {
				yamlBuilder.WriteString(fmt.Sprintf("%s%s%s:\n", commentPrefix, indentPrefix, yamlTag))
			} else {
				yamlBuilder.WriteString(fmt.Sprintf("%s%s%s: # %s\n", commentPrefix, indentPrefix, yamlTag, commentTag))
			}

			if value.Len() == 0 {
				yamlBuilder.WriteString(fmt.Sprintf("%s%s  -\n", commentPrefix, indentPrefix))
			}
			for j := range value.Len() {
				valueStr := getValueStr(value.Index(j))
				yamlBuilder.WriteString(fmt.Sprintf("%s%s  - %s\n", "", indentPrefix, valueStr))
			}
			continue
		}

		valueStr := getValueStr(value)
		if commentTag == "" {
			yamlBuilder.WriteString(fmt.Sprintf("%s%s%s: %s\n", commentPrefix, indentPrefix, yamlTag, valueStr))
		} else {
			yamlBuilder.WriteString(fmt.Sprintf("%s%s%s: %s # %s\n", commentPrefix, indentPrefix, yamlTag, valueStr, commentTag))
		}
	}
}

func getValueStr(value reflect.Value) string {
	var valueStr string
	if value.Kind() == reflect.String {
		valueStr = fmt.Sprintf("\"%v\"", value.Interface())
	} else {
		valueStr = fmt.Sprintf("%v", value.Interface())
	}
	return valueStr
}

func checkIfMarshalYAMLExists(marshalYAMLMethod reflect.Value) bool {
	if marshalYAMLMethod.IsValid() && marshalYAMLMethod.Type().NumIn() == 0 && marshalYAMLMethod.Type().NumOut() == 2 {
		// At this point we know that the method is valid and has no input and two output parameters. Next we check the types of the output parameters.

		anyType := reflect.TypeFor[any]()
		returnsAnyFirst := marshalYAMLMethod.Type().Out(0) == anyType

		errorType := reflect.TypeFor[error]()
		returnsErrorSecond := marshalYAMLMethod.Type().Out(1) == errorType

		return returnsAnyFirst && returnsErrorSecond
	}
	return false
}

func callMarshalYAML(marshalYAMLMethod reflect.Value) (any, error) {
	const expectedNumOut = 2
	marshalYAMLResult := marshalYAMLMethod.Call(nil)
	if len(marshalYAMLResult) != expectedNumOut {
		panic(fmt.Sprintf("MarshalYAML method did not return two values as expected but %d", len(marshalYAMLResult)))
	}

	marshalRes := marshalYAMLResult[0].Interface()

	err, ok := marshalYAMLResult[1].Interface().(error)
	if !ok && err != nil {
		panic(fmt.Sprintf("MarshalYAML method second result is not an error as expected but %T", marshalYAMLResult[1].Interface()))
	}

	return marshalRes, err
}

func tryMarshalYAML(value reflect.Value) any {
	// Check if there is a MarshalYAML method defined for this struct
	marshalYAMLMethod := value.MethodByName("MarshalYAML")
	if checkIfMarshalYAMLExists(marshalYAMLMethod) {
		// The MarshalYAML method matches the expected signature: MarshalYAML() (any, error), try to use it to generate YAML data
		marshalRes, err := callMarshalYAML(marshalYAMLMethod)
		if err != nil {
			panic(fmt.Sprintf("MarshalYAML method returned an error: %v", err))
		}

		return marshalRes
	}
	// If there is no MarshalYAML method, return nil
	return nil
}

// tryPrintIfString checks if someVal is a string and if so, prints it into the yamlBuilder and returns true.
func tryPrintIfString(someVal any, yamlBuilder *strings.Builder, commentTag, commentPrefix, indentPrefix, yamlTag string) bool {
	// check if someVal is printable as a string
	yamlContent, ok := someVal.(string)
	if ok {
		if commentTag == "" {
			yamlBuilder.WriteString(fmt.Sprintf("%s%s%s: \"%s\"\n", commentPrefix, indentPrefix, yamlTag, yamlContent))
		} else {
			yamlBuilder.WriteString(fmt.Sprintf("%s%s%s: \"%s\" # %s\n", commentPrefix, indentPrefix, yamlTag, yamlContent, commentTag))
		}
		return true
	}
	return false
}
