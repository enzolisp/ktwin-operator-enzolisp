package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	v0 "ktwin/operator/api/dtd/v0"
	dtdl "ktwin/operator/cmd/cli/dtdl"
	"ktwin/operator/cmd/cli/graph"
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

	dtdlGraph := graph.NewTwinInterfaceGraph()

	dtdlGraph = processAllFilesInFolder(inputFolderPath, outputFolderPath, resourceBuilder, dtdlGraph)

	dtdlGraph.PrintGraph()
}

// Process all files in the specified folder
func processAllFilesInFolder(inputFolderPath string, outputFolderPath string, resourceBuilder pkg.ResourceBuilder, dtdlGraph graph.TwinInterfaceGraph) graph.TwinInterfaceGraph {
	files, err := ioutil.ReadDir(inputFolderPath)

	if err != nil {
		log.Fatal(err)
	}

	serializer := k8sJson.NewYAMLSerializer(k8sJson.DefaultMetaFactory, nil, nil)

	pkg.PrepareOutputFolder(outputFolderPath)

	for _, file := range files {
		if !file.IsDir() {
			if !strings.Contains(file.Name(), ".json") {
				continue
			}

			fmt.Println("Processing file " + file.Name())
			dtdlGraph = processFile(inputFolderPath, file, outputFolderPath, resourceBuilder, serializer, dtdlGraph)
		} else {
			fmt.Println("Processing directory " + file.Name())
			// The file is a directory, get into the the directory and process the files recursively
			nestedInputFolderPath := inputFolderPath + "/" + file.Name()
			nestedOutputFolderPath := outputFolderPath + "/" + file.Name()
			dtdlGraph = processAllFilesInFolder(nestedInputFolderPath, nestedOutputFolderPath, resourceBuilder, dtdlGraph)
		}
	}

	return dtdlGraph
}

func processFile(inputFolderPath string, file fs.FileInfo, outputFolderPath string, resourceBuilder pkg.ResourceBuilder, serializer *k8sJson.Serializer, dtdlGraph graph.TwinInterfaceGraph) graph.TwinInterfaceGraph {
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

		ti := resourceBuilder.CreateTwinInterface(twinInterface)
		tinstance := resourceBuilder.CreateTwinInstance(ti)

		dtdlGraph = processGraph(dtdlGraph, ti)

		yamlBuffer := new(bytes.Buffer)
		serializer.Encode(&ti, yamlBuffer)
		yamlBuffer.Write([]byte("---\n"))
		serializer.Encode(&tinstance, yamlBuffer)

		if err != nil {
			log.Fatal(err)
		}

		err = pkg.WriteToFile(outputFilename, yamlBuffer.Bytes())
		if err != nil {
			log.Fatal(err)
		}
	}

	return dtdlGraph
}

func processGraph(dtdlGraph graph.TwinInterfaceGraph, twinInterface v0.TwinInterface) graph.TwinInterfaceGraph {
	dtdlGraph.AddVertex(twinInterface)

	for _, relationship := range twinInterface.Spec.Relationships {
		tInterface := v0.TwinInterface{
			Spec: v0.TwinInterfaceSpec{
				Id: relationship.Target,
			},
		}
		dtdlGraph.AddEdge(twinInterface, tInterface)
	}

	return dtdlGraph
}
