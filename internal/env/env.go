package env

import "os"

// Get the value of an environmental variable.
func GetString(key, fallback string) string {
	// If the key exists in the environment, return its value.
	if val := os.Getenv(key); val != "" {
		return val
	}

	// No such key found in the envronment variables, return the fallback string.
	return fallback
}
