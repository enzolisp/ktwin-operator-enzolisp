package graph

import (
	"errors"
	"fmt"

	dtdv0 "github.com/Open-Digital-Twin/ktwin-operator/api/dtd/v0"
)

const (
	NO_INDEX = -1
)

// A Graph is composed by multiple vertexes connected by edges

type TwinInstanceGraph interface {
	AddVertex(twinInstance dtdv0.TwinInstance) (*TwinInstanceGraphVertex, error)
	GetVertex(twinInstanceId string) *dtdv0.TwinInstance
	RemoveVertex(twinInstance dtdv0.TwinInstance) error
	AddEdge(sourceTwinInstance dtdv0.TwinInstance, targetTwinInstance dtdv0.TwinInstance) error
	RemoveEdge(sourceTwinInstance dtdv0.TwinInstance, targetTwinInstance dtdv0.TwinInstance) error
	PrintGraph()
}

type twinInstanceGraph struct {
	NumberOfVertex int
	Vertexes       map[string]*TwinInstanceGraphVertex
}

type TwinInstanceGraphVertex struct {
	TwinInstance  dtdv0.TwinInstance
	EdgeInstances []*TwinInstanceGraphVertex

	// Used for when the vertex was not processed yet, but it is listed in a relationship
	// At the end, none of the vertexes must be temporary
	HasTemporaryInstance bool
}

func NewTwinInstanceGraph() TwinInstanceGraph {
	return &twinInstanceGraph{
		NumberOfVertex: 0,
		Vertexes:       map[string]*TwinInstanceGraphVertex{},
	}
}

func (g *twinInstanceGraph) GetVertex(twinInstanceId string) *dtdv0.TwinInstance {
	if g.Vertexes[twinInstanceId] == nil {
		return nil
	}

	return &g.Vertexes[twinInstanceId].TwinInstance
}

func (g *twinInstanceGraph) AddVertex(twinInstance dtdv0.TwinInstance) (*TwinInstanceGraphVertex, error) {
	vertex := g.Vertexes[twinInstance.Name]
	if vertex != nil {
		if vertex.HasTemporaryInstance {
			g.Vertexes[twinInstance.Name].TwinInstance = twinInstance
			g.Vertexes[twinInstance.Name].HasTemporaryInstance = false
		}
		return g.Vertexes[twinInstance.Name], errors.New("TwinInstance already exist in the graph")
	}

	g.Vertexes[twinInstance.Name] = &TwinInstanceGraphVertex{
		TwinInstance:  twinInstance,
		EdgeInstances: []*TwinInstanceGraphVertex{},
	}

	g.NumberOfVertex = g.NumberOfVertex + 1

	return g.Vertexes[twinInstance.Name], nil
}

func (g *twinInstanceGraph) addTemporaryVertex(twinInstance dtdv0.TwinInstance) (*TwinInstanceGraphVertex, error) {
	if g.Vertexes[twinInstance.Name] != nil {
		return g.Vertexes[twinInstance.Name], errors.New("TwinInstance already exist in the graph")
	}

	g.Vertexes[twinInstance.Name] = &TwinInstanceGraphVertex{
		TwinInstance:         twinInstance,
		EdgeInstances:        []*TwinInstanceGraphVertex{},
		HasTemporaryInstance: true,
	}

	g.NumberOfVertex = g.NumberOfVertex + 1

	return g.Vertexes[twinInstance.Name], nil
}

func (g *twinInstanceGraph) RemoveVertex(twinInstance dtdv0.TwinInstance) error {
	if g.Vertexes[twinInstance.Name] == nil {
		return errors.New("TwinInstance does not exist in the graph")
	}

	delete(g.Vertexes, twinInstance.Name)

	for _, graphVertex := range g.Vertexes {
		if graphVertex == nil {
			continue
		}

		index := g.findEdgeIndex(graphVertex.EdgeInstances, twinInstance)

		if index != NO_INDEX {
			graphVertex.EdgeInstances = g.removeIndex(graphVertex.EdgeInstances, index)
		}
	}

	g.NumberOfVertex = g.NumberOfVertex - 1

	return nil
}

func (g *twinInstanceGraph) AddEdge(sourceTwinInstance dtdv0.TwinInstance, targetTwinInstance dtdv0.TwinInstance) error {

	// Add both Vertex if they not exist, if some of they exist, just ignore error
	sourceTwinInstanceGraph, _ := g.addTemporaryVertex(sourceTwinInstance)
	targetTwinInstanceGraph, _ := g.addTemporaryVertex(targetTwinInstance)

	sourceTwinInstanceGraph.EdgeInstances = append(sourceTwinInstanceGraph.EdgeInstances, targetTwinInstanceGraph)
	//targetTwinInstanceGraph.EdgeInterfaces = append(targetTwinInstanceGraph.EdgeInterfaces, sourceTwinInstanceGraph)
	g.NumberOfVertex = g.NumberOfVertex + 1

	return nil
}

func (g *twinInstanceGraph) RemoveEdge(sourceTwinInstance dtdv0.TwinInstance, targetTwinInstance dtdv0.TwinInstance) error {

	sourceVertex := g.Vertexes[sourceTwinInstance.Name]
	targetVertex := g.Vertexes[targetTwinInstance.Name]

	if sourceVertex != nil {
		index := g.findEdgeIndex(sourceVertex.EdgeInstances, targetTwinInstance)
		if index != NO_INDEX {
			g.removeIndex(sourceVertex.EdgeInstances, index)
		}
	}

	if targetVertex != nil {
		index := g.findEdgeIndex(targetVertex.EdgeInstances, sourceTwinInstance)
		if index != NO_INDEX {
			g.removeIndex(targetVertex.EdgeInstances, index)
		}
	}

	return nil
}

func (g *twinInstanceGraph) PrintGraph() {

	fmt.Println("\nGraph: ")
	for _, vertex := range g.Vertexes {
		fmt.Print("Vertex: " + vertex.TwinInstance.Name + " - ")
		fmt.Print("Relationships: ")

		for _, edge := range vertex.EdgeInstances {
			fmt.Print(edge.TwinInstance.Name)
			fmt.Print(", ")
		}

		fmt.Println("")
	}
}

func (g *twinInstanceGraph) findEdgeIndex(edgeInstances []*TwinInstanceGraphVertex, twinInstance dtdv0.TwinInstance) int {
	for index, twinInstanceGraph := range edgeInstances {
		if twinInstanceGraph != nil && twinInstanceGraph.TwinInstance.Name == twinInstance.Name {
			return index
		}
	}

	return NO_INDEX
}

func (g *twinInstanceGraph) removeIndex(edgeInstances []*TwinInstanceGraphVertex, index int) []*TwinInstanceGraphVertex {
	return append(edgeInstances[:index], edgeInstances[index+1:]...)
}
