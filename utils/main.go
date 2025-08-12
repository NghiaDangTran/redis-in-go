package utils

import (
	"fmt"
	"strconv"
	"time"
)

const base36Width = 13

func TimeHash() string {
	n := time.Now().UTC().UnixNano()
	s := strconv.FormatInt(n, 36)
	return fmt.Sprintf("%0*s", base36Width, s)
}
