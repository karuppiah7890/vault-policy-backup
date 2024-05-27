package main

import "os"

func writeToFile(fileContent []byte, fileName string) error {
	err := os.WriteFile(fileName, fileContent, 0644)
	if err != nil {
		return err
	}
	return nil
}
