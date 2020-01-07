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
