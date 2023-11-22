package pkg

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func IsJsonFile(filePath string) bool {
	return filepath.Ext(filePath) == ".json"
}

func AddSuffixToFileName(filePath string, prefix string, suffix string) string {
	directory, fileNameWithExtension := filepath.Split(filePath)
	fileExtension := filepath.Ext(fileNameWithExtension)
	fileName := strings.Split(fileNameWithExtension, ".")

	finalFileName := directory + prefix + fileName[0] + suffix + fileExtension

	return finalFileName
}

func PrepareOutputFolder(dirname string) error {
	fileInfo, err := os.Stat(dirname)

	if err == nil && fileInfo.IsDir() {
		return nil
	}

	if err != nil && !os.IsNotExist(err) {
		log.Fatal("Output path has a file, it is impossible to proceed")
	}

	if err != nil && os.IsNotExist(err) {
		err = os.Mkdir(dirname, os.ModePerm)
	}

	if os.IsExist(err) {
		info, err := os.Stat(dirname)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			log.Fatal("File is not a directory")
		} else {
			log.Default().Print("Folder already exists")
		}
	}

	return err
}

func WriteToFile(fileName string, data []byte) error {

	err := os.WriteFile(fileName, data, 0664)

	if err != nil {
		fmt.Println("Error while opening file: " + fileName)
		return err
	}

	return nil
}
