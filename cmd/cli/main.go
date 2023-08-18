package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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

type ProcessedFile struct {
	InputFilePath   string
	outputFilePath  string
	TwinInterfaceId string
}

func main() {
	allArgs := os.Args
	args := allArgs[1:]

	if len(args) < 2 {
		log.Fatal("Inform DTDL input and output folders path")
	}

	inputFolderPath := args[0]
	outputFolderPath := args[1]

	dtdlGraph := graph.NewTwinInterfaceGraph()
	processedFiles := []ProcessedFile{}

	// Load all DTDL interfaces files
	fmt.Println("Processing folder " + inputFolderPath)

	dtdlGraph, processedFiles = processAllFilesInFolder(inputFolderPath, outputFolderPath, dtdlGraph, processedFiles)

	// Print Graph
	dtdlGraph.PrintGraph()

	// Generate Output files with TwinInterfaces and TwinInstances examples
	generateOutputFiles(processedFiles, dtdlGraph)

}

// Process all files in the specified folder
func processAllFilesInFolder(inputFolderPath string, outputFolderPath string, dtdlGraph graph.TwinInterfaceGraph, processedFiles []ProcessedFile) (graph.TwinInterfaceGraph, []ProcessedFile) {
	files, err := ioutil.ReadDir(inputFolderPath)

	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		inputFilePath := filepath.Join(inputFolderPath, file.Name())

		if !file.IsDir() {
			if !pkg.IsJsonFile(inputFilePath) {
				continue
			}

			fmt.Println("Processing file " + file.Name())
			twinInterface := loadDTDLFileIntoGraph(inputFilePath)

			outputFileName := strings.Split(file.Name(), ".")[0]
			outputFilePath := filepath.Join(outputFolderPath, outputFileName+".yaml")

			processedFiles = append(processedFiles, ProcessedFile{
				InputFilePath:   inputFilePath,
				outputFilePath:  outputFilePath,
				TwinInterfaceId: twinInterface.Spec.Id,
			})

			dtdlGraph = updateGraph(dtdlGraph, twinInterface)

		} else {
			fmt.Println("Processing directory " + file.Name())

			// The file is a directory, get into the the directory and process the files recursively
			nestedInputFolderPath := inputFolderPath + "/" + file.Name()
			nestedOutputFolderPath := outputFolderPath + "/" + file.Name()
			dtdlGraph, processedFiles = processAllFilesInFolder(nestedInputFolderPath, nestedOutputFolderPath, dtdlGraph, processedFiles)
		}
	}

	return dtdlGraph, processedFiles
}

func loadDTDLFileIntoGraph(inputFilePath string) v0.TwinInterface {
	fileContent, err := os.ReadFile(inputFilePath)
	if err != nil {
		log.Fatal(err)
	}

	twinInterface := dtdl.Interface{}
	err = json.Unmarshal(fileContent, &twinInterface)

	if err != nil {
		log.Fatal(err)
	}

	twinInterfaceResource := pkg.NewResourceBuilder().CreateTwinInterface(twinInterface)
	return twinInterfaceResource
}

// Process Inherit Properties and Telemetries
func processInheritContents(dtdlGraph graph.TwinInterfaceGraph) graph.TwinInterfaceGraph {
	return dtdlGraph
}

func generateOutputFiles(processedFiles []ProcessedFile, dtdlGraph graph.TwinInterfaceGraph) {

	fmt.Println("Generating output files...")

	for _, processedFile := range processedFiles {

		twinInterface := dtdlGraph.GetVertex(processedFile.TwinInterfaceId)

		if twinInterface == nil {
			fmt.Printf("Twin Interface {%s} not found\n", processedFile.TwinInterfaceId)
			continue
		}

		twinInstance := pkg.NewResourceBuilder().CreateTwinInstance(*twinInterface)
		writeOutputFile(processedFile.outputFilePath, *twinInterface, twinInstance)
	}
}

func writeOutputFile(outputFilePath string, twinInterface v0.TwinInterface, twinInstance v0.TwinInstance) {
	outputFolderPath := filepath.Dir(outputFilePath)
	subFoldersPath := strings.Split(outputFolderPath, "/")

	fmt.Printf("Writing file " + outputFilePath + "\n")

	var outputSubFolderPath string
	for _, subFolderPath := range subFoldersPath {
		if subFolderPath != "" {
			outputSubFolderPath += "/"
			outputSubFolderPath += subFolderPath
			pkg.PrepareOutputFolder(outputSubFolderPath)
		}
	}

	serializer := k8sJson.NewYAMLSerializer(k8sJson.DefaultMetaFactory, nil, nil)
	yamlBuffer := new(bytes.Buffer)
	serializer.Encode(&twinInterface, yamlBuffer)
	yamlBuffer.Write([]byte("---\n"))
	serializer.Encode(&twinInstance, yamlBuffer)
	err := pkg.WriteToFile(outputFilePath, yamlBuffer.Bytes())
	if err != nil {
		log.Fatal(err)
	}
}

func updateGraph(dtdlGraph graph.TwinInterfaceGraph, twinInterface v0.TwinInterface) graph.TwinInterfaceGraph {
	dtdlGraph.AddVertex(twinInterface)

	for _, relationship := range twinInterface.Spec.Relationships {
		tInterface := v0.TwinInterface{
			Spec: v0.TwinInterfaceSpec{
				Id: relationship.Interface,
			},
		}
		dtdlGraph.AddEdge(twinInterface, tInterface)
	}

	return dtdlGraph
}
