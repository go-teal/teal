package utils

import (
	"os"
	"strings"
)

func CheckModelExists(filepath, modelName, extenstion string) (bool, error) {
	_, err := os.Stat(filepath + "/" + strings.Replace(modelName, ".", "/", 1) + "." + extenstion)
	return (err == nil), err
}
