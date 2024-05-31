package main

import "os"

func writeToFile(fileContent []byte, filePath string) error {
	err := os.WriteFile(filePath, fileContent, 0644)
	if err != nil {
		return err
	}
	return nil
}
