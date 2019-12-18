package main

import (
	"fmt"
)

func getValidString(data []byte) string {
	var validStr = "valid"
	if valid := checkValidData(data); !valid {
		validStr = "not valid"
	}
	return validStr
}

func main() {
	data, err := getPrivData()
	if err != nil {
		fmt.Printf("Error: %v", err)
	}
	valid := getValidString(data)
	fmt.Printf("privateData is: %x (%s)\n", data, valid)
}
