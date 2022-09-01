package jsonutils

import (
	"encoding/json"
	"os"

	g "github.com/lassi-koykka/fin-dev-api/src/utils"
)

func JsonParse[T any](bodyData []byte) T {
	var data T
	err := json.Unmarshal(bodyData, &data)
	g.Check(err)
	return data
}

func JsonSerialize[T any](data T) []byte {
	res, err := json.Marshal(data)
	g.Check(err)
	return res
}

func JsonStringSerialize[T any](data T) string {
	return string(JsonSerialize(data))
}

func WriteJsonFile[T any](path string, data T) {
	serializedData := JsonSerialize(data)
	err := os.WriteFile(path, serializedData, 0666)
	g.Check(err)
}
