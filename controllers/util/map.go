package util

// MergeLabels merges all the label maps received as argument into a single new label map.
func MergeMap(a_map ...map[string]string) map[string]string {
	res := map[string]string{}

	for _, v := range a_map {
		if v != nil {
			for k, v := range v {
				res[k] = v
			}
		}
	}
	return res
}
