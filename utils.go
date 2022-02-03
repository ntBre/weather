package main

func max(ds ...int) int {
	max := ds[0]
	if len(ds) < 2 {
		return max
	}
	for _, d := range ds[1:] {
		if d > max {
			max = d
		}
	}
	return max
}

func min(ds ...int) int {
	min := ds[0]
	if len(ds) < 2 {
		return min
	}
	for _, d := range ds[1:] {
		if d < min && d != -999 {
			min = d
		}
	}
	return min
}
