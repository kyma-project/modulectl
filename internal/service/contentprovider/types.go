package contentprovider

type ObjectToYAMLConverter interface {
	ConvertToYaml(obj interface{}) string
}
