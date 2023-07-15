package graph

import (
	"errors"
	"fmt"
	dtdv0 "ktwin/operator/api/dtd/v0"
)

const (
	NO_INDEX = -1
)

// A Graph is composed by multiple vertexes connected by edges

type TwinInterfaceGraph interface {
	AddVertex(twinInterface dtdv0.TwinInterface) (*TwinInterfaceGraphVertex, error)
	GetVertex(twinInterfaceId string) *dtdv0.TwinInterface
	RemoveVertex(twinInterface dtdv0.TwinInterface) error
	AddEdge(sourceTwinInterface dtdv0.TwinInterface, targetTwinInterface dtdv0.TwinInterface) error
	RemoveEdge(sourceTwinInterface dtdv0.TwinInterface, targetTwinInterface dtdv0.TwinInterface) error
	PrintGraph()
}

type twinInterfaceGraph struct {
	NumberOfVertex int
	Vertexes       map[string]*TwinInterfaceGraphVertex
}

type TwinInterfaceGraphVertex struct {
	TwinInterface  dtdv0.TwinInterface
	EdgeInterfaces []*TwinInterfaceGraphVertex
}

func NewTwinInterfaceGraph() TwinInterfaceGraph {
	return &twinInterfaceGraph{
		NumberOfVertex: 0,
		Vertexes:       map[string]*TwinInterfaceGraphVertex{},
	}
}

func (g *twinInterfaceGraph) GetVertex(twinInterfaceId string) *dtdv0.TwinInterface {
	if g.Vertexes[twinInterfaceId] == nil {
		return nil
	}

	return &g.Vertexes[twinInterfaceId].TwinInterface
}

func (g *twinInterfaceGraph) AddVertex(twinInterface dtdv0.TwinInterface) (*TwinInterfaceGraphVertex, error) {
	if g.Vertexes[twinInterface.Spec.Id] != nil {
		return g.Vertexes[twinInterface.Spec.Id], errors.New("TwinInterface already exist in the graph")
	}

	g.Vertexes[twinInterface.Spec.Id] = &TwinInterfaceGraphVertex{
		TwinInterface:  twinInterface,
		EdgeInterfaces: []*TwinInterfaceGraphVertex{},
	}

	g.NumberOfVertex = g.NumberOfVertex + 1

	return g.Vertexes[twinInterface.Spec.Id], nil
}

func (g *twinInterfaceGraph) RemoveVertex(twinInterface dtdv0.TwinInterface) error {
	if g.Vertexes[twinInterface.Spec.Id] == nil {
		return errors.New("TwinInterface does not exist in the graph")
	}

	delete(g.Vertexes, twinInterface.Spec.Id)

	for _, graphVertex := range g.Vertexes {
		if graphVertex == nil {
			continue
		}

		index := g.findEdgeIndex(graphVertex.EdgeInterfaces, twinInterface)

		if index != NO_INDEX {
			graphVertex.EdgeInterfaces = g.removeIndex(graphVertex.EdgeInterfaces, index)
		}
	}

	g.NumberOfVertex = g.NumberOfVertex - 1

	return nil
}

func (g *twinInterfaceGraph) AddEdge(sourceTwinInterface dtdv0.TwinInterface, targetTwinInterface dtdv0.TwinInterface) error {

	// Add both Vertex if they not exist, if some of they exist, just ignore error
	sourceTwinInterfaceGraph, _ := g.AddVertex(sourceTwinInterface)
	targetTwinInterfaceGraph, _ := g.AddVertex(targetTwinInterface)

	sourceTwinInterfaceGraph.EdgeInterfaces = append(sourceTwinInterfaceGraph.EdgeInterfaces, targetTwinInterfaceGraph)
	//targetTwinInterfaceGraph.EdgeInterfaces = append(targetTwinInterfaceGraph.EdgeInterfaces, sourceTwinInterfaceGraph)
	g.NumberOfVertex = g.NumberOfVertex + 1

	return nil
}

func (g *twinInterfaceGraph) RemoveEdge(sourceTwinInterface dtdv0.TwinInterface, targetTwinInterface dtdv0.TwinInterface) error {

	sourceVertex := g.Vertexes[sourceTwinInterface.Spec.Id]
	targetVertex := g.Vertexes[targetTwinInterface.Spec.Id]

	if sourceVertex != nil {
		index := g.findEdgeIndex(sourceVertex.EdgeInterfaces, targetTwinInterface)
		if index != NO_INDEX {
			g.removeIndex(sourceVertex.EdgeInterfaces, index)
		}
	}

	if targetVertex != nil {
		index := g.findEdgeIndex(targetVertex.EdgeInterfaces, sourceTwinInterface)
		if index != NO_INDEX {
			g.removeIndex(targetVertex.EdgeInterfaces, index)
		}
	}

	return nil
}

func (g *twinInterfaceGraph) PrintGraph() {

	fmt.Println("\nGraph: ")
	for _, vertex := range g.Vertexes {
		fmt.Print("Vertex: " + vertex.TwinInterface.Spec.Id + " - ")
		fmt.Print("Relationships: ")

		for _, edge := range vertex.EdgeInterfaces {
			fmt.Print(edge.TwinInterface.Spec.Id)
			fmt.Print(", ")
		}

		fmt.Println("")
	}
}

func (g *twinInterfaceGraph) findEdgeIndex(edgeInterfaces []*TwinInterfaceGraphVertex, twinInterface dtdv0.TwinInterface) int {
	for index, twinInterfaceGraph := range edgeInterfaces {
		if twinInterfaceGraph != nil && twinInterfaceGraph.TwinInterface.Spec.Id == twinInterface.Spec.Id {
			return index
		}
	}

	return NO_INDEX
}

func (g *twinInterfaceGraph) removeIndex(edgeInterfaces []*TwinInterfaceGraphVertex, index int) []*TwinInterfaceGraphVertex {
	return append(edgeInterfaces[:index], edgeInterfaces[index+1:]...)
}
