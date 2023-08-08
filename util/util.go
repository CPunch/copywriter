package util

import (
	"fmt"
	"os"
	"time"
)

func GetEnv(key string, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	return value
}

func GetTimeString() string {
	dt := time.Now()
	return fmt.Sprintf(dt.Format("2006-01-02 15:04:05"))
}
