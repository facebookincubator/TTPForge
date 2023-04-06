package strings

// StringSlicesEqual checks if two string slices are equal.
//
// Borrowed from https://github.com/l50/goutils/blob/main/stringutils.go#L51-L73
//
// It returns true if the slices have the same length and same values, false otherwise.
//
// Parameters:
// a: the first string slice to be compared.
// b: the second string slice to be compared.
//
// Returns:
// bool: true if the slices are equal, false otherwise.
func StringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
