package myLib

import (
	"fmt"
	"os"
)

const (
	Npos = -1
)

func CheckError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error ", err.Error())
		os.Exit(1)
	}
}


// reverse an array of type byte
func Reverse(s []byte) []byte {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}


// checks if the elem is is the array
// if yes returns the index of the element else returns myLib.Npos
func ContainsInt(arr []int, elem int) int {
	for i,e := range arr {
		if e == elem {
			return i
		}
	}
	return Npos
}

// remove an element by index from an int array
func RemoveInt(arr []int, index int) []int {
	return append(arr[:index], arr[index+1:]...)
}