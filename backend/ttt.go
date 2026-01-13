package main

import "fmt"

func main() {
	words := []string{"fusion", "layout"}
	fmt.Println(countPairs(words))
}

func countPairs(words []string) (ans int64) {
	for i := 0; i < len(words); i++ {
		for j := i + 1; j < len(words); j++ {
			if checkStr(words[i], words[j]) {
				ans++
			}
		}
	}
	return
}

func checkStr(a, b string) bool {
	t := check(a[0], b[0])
	for i := 1; i < len(a); i++ {
		if check(a[i], b[i]) != t {
			return false
		}
	}
	return true
}

func check(a, b byte) byte {
	return (a + 'z' - 'a' - b) % 26
}
