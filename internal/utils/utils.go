package utils

func MergeAndDeduplicateSlices(slices ...[]string) []string {
	itemSet := make(map[string]struct{})
	for _, slice := range slices {
		for _, item := range slice {
			if item != "" {
				itemSet[item] = struct{}{}
			}
		}
	}
	result := make([]string, 0, len(itemSet))
	for item := range itemSet {
		result = append(result, item)
	}
	return result
}
