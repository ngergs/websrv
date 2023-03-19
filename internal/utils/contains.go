package utils

import "strings"

// Contains checks if the list slice contains the val.
func Contains[T comparable](list []T, val T) bool {
	for _, el := range list {
		if el == val {
			return true
		}
	}
	return false
}

// ContainsAfterSplit checks if the list string slice contains the val string after splitting each element along the splitter.
func ContainsAfterSplit(list []string, splitter string, val string) bool {
	for _, el := range list {
		splitted := strings.Split(el, splitter)
		if Contains(splitted, val) {
			return true
		}
	}
	return false
}
