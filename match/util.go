package match

func minLen(ms []Matcher) (min int) {
	for i, m := range ms {
		n := m.MinLen()
		if i == 0 || n < min {
			min = n
		}
	}
	return min
}

func maxLen(ms []Matcher) (max int) {
	for i, m := range ms {
		n := m.MinLen()
		if i == 0 || n > max {
			max = n
		}
	}
	return max
}
