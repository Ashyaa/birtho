package util

func Max(arr []int) int {
	res := 0
	for _, n := range arr {
		if n > res {
			res = n
		}
	}
	return res
}

func ToHashMap[T comparable](arr []T) map[T]bool {
	res := make(map[T]bool)
	for _, item := range arr {
		res[item] = true
	}
	return res
}

func ToSlice[T any](dict map[string]T) []T {
	res := []T{}
	for _, item := range dict {
		res = append(res, item)
	}
	return res
}
