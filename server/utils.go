package server

import "strings"

func contains(list []string, val string) bool {
	for _, el := range list {
		if el == val {
			return true
		}
	}
	return false
}

func containsAfterSplit(list []string, splitter string, val string) bool {
	for _, el := range list {
		splitted := strings.Split(el, splitter)
		if contains(splitted, val) {
			return true
		}
	}
	return false
}
