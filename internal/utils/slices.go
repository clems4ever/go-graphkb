package utils

import "reflect"

import "log"

func IsStringInSlice(s string, slice []string) bool {
	for _, str := range slice {
		if str == s {
			return true
		}
	}
	return false
}

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

func ChunkSlice(s interface{}, chunkSize int) interface{} {
	if reflect.ValueOf(s).Kind() != reflect.Slice {
		log.Fatal("ChunkSlice only work on slices")
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
