package language

// Copied from https://github.com/sc0Vu/didyoumean/blob/main/didyoumean.go
// Would import package directly but this creates a bit of a supply chain issue due to low traffic.
//
// Slight modifications to terminology and such but corpus is unaffected.

/*
	MIT License

	Copyright (c) 2020 Peter Lai

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE.
*/

import (
	"strings"
	"unicode/utf8"
)

var (
	// ThresholdRate is the rate that allows the edit distance less than, eg 0.4
	// means the edit distance less than 40%
	ThresholdRate float64
	// CaseInsensitive compare the edit distance in case insensitive mode
	CaseInsensitive bool
)

// minimum returns minimum value
func minimum(values ...int) (min int) {
	min = values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
	}
	return
}

// findEditDistance returns edit distance between a and b
// it's use Levenshtein distance, see: https://en.wikipedia.org/wiki/Levenshtein_distance for
// more information
func findEditDistance(a, b string) (distance int) {
	if !utf8.ValidString(a) && !utf8.ValidString(b) {
		return
	}
	ra := []rune(a)
	rb := []rune(b)
	lenA := len(ra)
	lenB := len(rb)
	totalLen := lenB + 1
	// only use two rows
	v0 := make([]int, totalLen)
	v1 := make([]int, totalLen)
	// initialize first row
	for i := 0; i < totalLen; i++ {
		v0[i] = i
	}
	// loop through the text
	for i := 0; i < lenA; i++ {
		v1[0] = i + 1
		for j := 0; j < lenB; j++ {
			dc := v0[j+1] + 1
			ic := v1[j] + 1
			sc := v0[j]
			if a[i] != b[j] {
				sc++
			}
			min := minimum(dc, ic, sc)
			v1[j+1] = min
		}
		// copy v1 to v0
		copy(v0, v1)
	}
	distance = v0[lenB]
	return
}

// DidYouMean returns first match of didyoumean
func DidYouMean(key string, list []string) (result string) {
	if len(key) == 0 {
		return
	}
	if CaseInsensitive {
		key = strings.ToLower(key)
	}
	var winner int
	if ThresholdRate > 0 {
		winner = int(ThresholdRate * float64(len(key)))
	}
	for _, str := range list {
		distance := findEditDistance(key, str)
		if winner <= 0 || distance <= winner {
			// winner = distance
			result = str
			return
		}
	}
	return
}

// Match returns all match of didyoumean
func Match(key string, list []string) (results []string) {
	if len(key) == 0 {
		return
	}
	if CaseInsensitive {
		key = strings.ToLower(key)
	}
	var winner int
	if ThresholdRate > 0 {
		winner = int(ThresholdRate * float64(len(key)))
	}
	for _, result := range list {
		distance := findEditDistance(key, result)
		if winner <= 0 || distance <= winner {
			winner = distance
			results = append(results, result)
		}
	}
	return
}
