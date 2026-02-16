package contentprovider

type ObjectToYAMLConverter interface {
	ConvertToYaml(obj any) string
}
