package config

import (
	"fmt"
	"os"
	"strconv"
)

func GetByKey(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Errorf("key %s doesn't exist in .env", key))
	}
	return v
}

func GetIntConfig(key string) int {
	v := GetByKey(key)
	n, err := strconv.Atoi(v)
	if err != nil {
		panic(err)
	}
	return n

}
