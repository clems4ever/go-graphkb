package utils

import (
	"reflect"

	"github.com/sirupsen/logrus"
)

// IsStringInSlice check if string is in string slice
func IsStringInSlice(s string, slice []string) bool {
	for _, str := range slice {
		if str == s {
			return true
		}
	}
	return false
}

// AreStringSlicesEqual check equality between two slices
func AreStringSlicesEqual(s1 []string, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}

	for i := range s1 {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}

// AreStringSliceElementsEqual check the same elements are in both slices but they can be in different order
func AreStringSliceElementsEqual(s1 []string, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}

	for i := range s1 {
		el1 := s1[i]
		found := false
		for j := range s2 {
			if el1 == s2[j] {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// ChunkSlice chunk a slice in smaller chunks of size chunkSize
func ChunkSlice(s interface{}, chunkSize int) interface{} {
	if reflect.ValueOf(s).Kind() != reflect.Slice {
		logrus.Fatal("ChunkSlice only work on slices")
	}

	sValue := reflect.ValueOf(s)
	sLen := sValue.Len()

	chunks := make([][]interface{}, 0)
	var chunk []interface{}

	for i := 0; i < sLen; i++ {
		if i%chunkSize == 0 {
			chunk = make([]interface{}, 0)
		}

		value := sValue.Index(i)
		chunk = append(chunk, value.Interface())

		if ((i+1)%chunkSize == 0 || sLen-1 == i) && len(chunk) > 0 {
			chunks = append(chunks, chunk)
		}
	}
	return chunks
}
