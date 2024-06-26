package utils

import (
	"fmt"
	"os"
)

func CreateDir(dirName string) bool {
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		err := os.MkdirAll(dirName, 0755)
		if err != nil {
			fmt.Println("Error creating directory:", err)
			panic(err)
		}
		// fmt.Println("Directory created successfully:", dirName)
		return true
	}
	return false
}
