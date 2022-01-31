package server

func contains(list []string, val string) bool {
	for _, el := range list {
		if el == val {
			return true
		}
	}
	return false
}
