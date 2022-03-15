package util

func Contains(list []string, s string) bool {
	for _, elt := range list {
		if elt == s {
			return true
		}
	}
	return false
}

func AppendUnique(list []string, s string) []string {
	if Contains(list, s) {
		return list
	}
	return append(list, s)
}

// Returns the first found index of s in list, else -1
func Index(list []string, s string) int {
	for i, elt := range list {
		if elt == s {
			return i
		}
	}
	return -1
}

func Remove(list []string, s string) []string {
	idx := Index(list, s)
	if idx < 0 {
		return list
	}
	tmp := make([]string, len(list))
	copy(tmp, list)
	return append(tmp[:idx], tmp[idx+1:]...)
}
