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
