package utils

import "strings"

func Contains(list []string, val string) bool {
	for _, el := range list {
		if el == val {
			return true
		}
	}
	return false
}

func ContainsAfterSplit(list []string, splitter string, val string) bool {
	for _, el := range list {
		splitted := strings.Split(el, splitter)
		if Contains(splitted, val) {
			return true
		}
	}
	return false
}
