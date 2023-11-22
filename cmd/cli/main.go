package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	v0 "github.com/Open-Digital-Twin/ktwin-operator/api/dtd/v0"
	dtdl "github.com/Open-Digital-Twin/ktwin-operator/cmd/cli/dtdl"
	"github.com/Open-Digital-Twin/ktwin-operator/cmd/cli/graph"
	pkg "github.com/Open-Digital-Twin/ktwin-operator/cmd/cli/pkg"

	k8sJson "k8s.io/apimachinery/pkg/runtime/serializer/json"
)

type ProcessedFile struct {
	InputFilePath   string
	outputFilePath  string
	TwinInterfaceId string
}

func main() {
	inputFolderPath := flag.String("input-folder-path", "", "the input folder path to files")
	outputFolderPath := flag.String("output-folder-path", "", "the output folder path to files")
	numberOfInstances := flag.String("number-instances", "", "the number of instances to be created")

	flag.Parse()

	fmt.Println(*numberOfInstances)

	if *inputFolderPath == "" || *outputFolderPath == "" {
		log.Fatal("Inform DTDL input and output folders path")
	}

	dtdlGraph := graph.NewTwinInterfaceGraph()
	processedFiles := []ProcessedFile{}

	// Load all DTDL interfaces files
	fmt.Println("Processing folder " + *inputFolderPath)

	dtdlGraph, processedFiles = processAllFilesInFolder(*inputFolderPath, *outputFolderPath, dtdlGraph, processedFiles)

	// Print Graph
	dtdlGraph.PrintGraph()

	// Generate Output files with TwinInterfaces and TwinInstances examples
	generateOutputFiles(processedFiles, dtdlGraph)

}

// Process all files in the specified folder
func processAllFilesInFolder(inputFolderPath string, outputFolderPath string, dtdlGraph graph.TwinInterfaceGraph, processedFiles []ProcessedFile) (graph.TwinInterfaceGraph, []ProcessedFile) {
	files, err := os.ReadDir(inputFolderPath)

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
	//var outputString string

	fmt.Println("Generating output files...")

	for _, processedFile := range processedFiles {

		twinInterface := dtdlGraph.GetVertex(processedFile.TwinInterfaceId)

		if twinInterface == nil {
			fmt.Printf("Twin Interface {%s} not found\n", processedFile.TwinInterfaceId)
			continue
		}

		parentTwinInterfaces := getParentTwinInterfaces(*twinInterface, dtdlGraph)
		twinInstance := pkg.NewResourceBuilder().CreateTwinInstance(*twinInterface, parentTwinInterfaces)
		writeOutputFile(processedFile.outputFilePath, *twinInterface, twinInstance)
		// outputString += getTwinInstanceStringFields(twinInstance, parentTwinInterfaces)
	}

	// fmt.Printf(outputString)
}

func getTwinInstanceStringFields(twinInstance v0.TwinInstance, parentTwinInterfaces []v0.TwinInterface) string {
	var outputString string

	for _, twinInterface := range parentTwinInterfaces {
		twinInterfaceName := parentTwinInterfaces[0].Name

		for _, property := range twinInterface.Spec.Properties {
			var propertyList []string
			propertyList = append(propertyList, twinInterfaceName)
			propertyList = append(propertyList, "property")
			propertyList = append(propertyList, property.Name)
			propertyList = append(propertyList, property.Description)

			outputString = outputString + strings.Join(propertyList, ";;") + "\n"
		}

		for _, telemetry := range twinInterface.Spec.Telemetries {
			var telemetryList []string
			telemetryList = append(telemetryList, twinInterfaceName)
			telemetryList = append(telemetryList, "telemetry")
			telemetryList = append(telemetryList, telemetry.Name)
			telemetryList = append(telemetryList, telemetry.Description)

			outputString = outputString + strings.Join(telemetryList, ";;") + "\n"
		}

		for _, relationship := range twinInterface.Spec.Relationships {
			var relationshipList []string
			relationshipList = append(relationshipList, twinInterfaceName)
			relationshipList = append(relationshipList, "relationship")
			relationshipList = append(relationshipList, relationship.Name)
			relationshipList = append(relationshipList, relationship.Description)

			outputString = outputString + strings.Join(relationshipList, ";;") + "\n"
		}

	}

	return outputString
}

// Return a list of TwinInterfaces that contains the TwinInterface being processed and all the parent TwinInterfaces
func getParentTwinInterfaces(twinInterface v0.TwinInterface, dtdlGraph graph.TwinInterfaceGraph) []v0.TwinInterface {
	var parentTwinInterfaces []v0.TwinInterface

	parentTwinInterfaces = append(parentTwinInterfaces, twinInterface)

	if twinInterface.Spec.ExtendsInterface != "" {
		parentInterface := dtdlGraph.GetVertex(twinInterface.Spec.ExtendsInterface)

		if parentInterface == nil {
			return parentTwinInterfaces
		}

		parentInterfaceChain := getParentTwinInterfaces(*parentInterface, dtdlGraph)

		if parentInterfaceChain != nil {
			parentTwinInterfaces = append(parentTwinInterfaces, parentInterfaceChain...)
		}
	}

	return parentTwinInterfaces
}

func writeOutputFile(outputFilePath string, twinInterface v0.TwinInterface, twinInstance v0.TwinInstance) {
	absOutputFolderPath, err := filepath.Abs(outputFilePath)
	if err != nil {
		log.Fatal(err)
	}

	subFoldersPath := strings.Split(absOutputFolderPath, "/")

	fmt.Printf("Writing output files " + outputFilePath + "\n")

	var outputSubFolderPath string
	for _, subFolderPath := range subFoldersPath {
		if subFolderPath != "" && !strings.Contains(subFolderPath, ".") {
			outputSubFolderPath += "/"
			outputSubFolderPath += subFolderPath
			pkg.PrepareOutputFolder(outputSubFolderPath)
		}
	}

	// Write Twin Interface file
	serializer := k8sJson.NewYAMLSerializer(k8sJson.DefaultMetaFactory, nil, nil)
	yamlBuffer := new(bytes.Buffer)
	serializer.Encode(&twinInterface, yamlBuffer)
	interfaceFilePath := pkg.AddSuffixToFileName(outputFilePath, "01-", "-interface")
	err = pkg.WriteToFile(interfaceFilePath, yamlBuffer.Bytes())
	if err != nil {
		log.Fatal(err)
	}

	// Write Twin Instance files
	serializer = k8sJson.NewYAMLSerializer(k8sJson.DefaultMetaFactory, nil, nil)
	yamlBuffer = new(bytes.Buffer)
	serializer.Encode(&twinInstance, yamlBuffer)
	yamlBuffer.Write([]byte("---\n"))
	instanceFilePath := pkg.AddSuffixToFileName(outputFilePath, "02-", "-instances")
	err = pkg.WriteToFile(instanceFilePath, yamlBuffer.Bytes())
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
