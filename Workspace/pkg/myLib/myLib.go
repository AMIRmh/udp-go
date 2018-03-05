package myLib

import (
	"fmt"
	"os"
)

func CheckError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error ", err.Error())
		os.Exit(1)
	}
}

func Reverse(s []byte) []byte {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}