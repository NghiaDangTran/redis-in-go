package utils

import (
	"fmt"
	"hash/fnv"
	"time"
)

func TimeHash() string {
	now := time.Now().UnixNano()
	h := fnv.New32a()
	h.Write([]byte(fmt.Sprint(now)))
	hashValue := h.Sum32()
	shortHash := fmt.Sprintf("%x", hashValue)[:7]
	return shortHash
}
