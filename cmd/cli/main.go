package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	dtdl "ktwin/operator/cmd/cli/dtdl"
	pkg "ktwin/operator/cmd/cli/pkg"

	k8sJson "k8s.io/apimachinery/pkg/runtime/serializer/json"
)

func main() {
	allArgs := os.Args
	args := allArgs[1:]

	if len(args) < 2 {
		log.Fatal("Inform DTDL input and output folders path")
	}

	resourceBuilder := pkg.NewResourceBuilder()

	inputFolderPath := args[0]
	outputFolderPath := args[1]

	files, err := ioutil.ReadDir(inputFolderPath)

	if err != nil {
		log.Fatal(err)
	}

	serializer := k8sJson.NewYAMLSerializer(k8sJson.DefaultMetaFactory, nil, nil)

	pkg.PrepareOutputFolder(outputFolderPath)

	for _, file := range files {
		if !file.IsDir() {
			inputFilename := filepath.Join(inputFolderPath, file.Name())
			outputFileName := strings.Split(file.Name(), ".")[0]
			outputFilename := filepath.Join(outputFolderPath, outputFileName+".yaml")
			if pkg.IsJsonFile(inputFilename) {
				fileContent, err := os.ReadFile(inputFilename)
				if err != nil {
					log.Fatal(err)
				}

				twinInterface := dtdl.Interface{}
				err = json.Unmarshal(fileContent, &twinInterface)
				if err != nil {
					log.Fatal(err)
				}
				//twinYaml, err := yaml.Marshal(twinInterface)

				ti := resourceBuilder.CreateTwinInterface(twinInterface)
				tinstance := resourceBuilder.CreateTwinInstance(ti)

				yamlBuffer := new(bytes.Buffer)
				serializer.Encode(&ti, yamlBuffer)
				yamlBuffer.Write([]byte("---\n"))
				serializer.Encode(&tinstance, yamlBuffer)

				// fmt.Printf(yamlBuffer.String())

				if err != nil {
					log.Fatal(err)
				}

				err = pkg.WriteToFile(outputFilename, yamlBuffer.Bytes())
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}

}
