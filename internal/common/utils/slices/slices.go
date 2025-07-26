package slices

func MergeAndDeduplicate(slices ...[]string) []string {
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

func SetToSlice(itemSet map[string]struct{}) []string {
	items := make([]string, 0, len(itemSet))
	for item := range itemSet {
		items = append(items, item)
	}
	return items
}
