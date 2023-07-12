package graph

import (
	"errors"
	"testing"

	dtdv0 "ktwin/operator/api/dtd/v0"

	"github.com/stretchr/testify/assert"
)

var twinInstance01 = dtdv0.TwinInterface{
	Spec: dtdv0.TwinInterfaceSpec{
		Id: "TwinInterface01",
	},
}

var twinInstance02 = dtdv0.TwinInterface{
	Spec: dtdv0.TwinInterfaceSpec{
		Id: "TwinInterface02",
	},
}

func TestTwinInterfaceImplements_CreateGraph(t *testing.T) {
	t.Run("Should implement TwinInterfaceGraph", func(t *testing.T) {
		twinInterfaceGraph := NewTwinInterfaceGraph()
		assert.Implements(t, (*TwinInterfaceGraph)(nil), twinInterfaceGraph)
	})
}

func TestTwinInterface_CreateGraph(t *testing.T) {
	t.Run("Should create TwinInterface Graph", func(t *testing.T) {
		graph := NewTwinInterfaceGraph()
		assert.Equal(t, &twinInterfaceGraph{
			NumberOfVertex: 0,
			Vertexes:       map[string]*TwinInterfaceGraphVertex{},
		}, graph)
	})
}

func TestTwinInterface_AddVertex(t *testing.T) {

	type VertexToBeAdded struct {
		twinInterface dtdv0.TwinInterface
		expectedError error
	}

	tests := []struct {
		name            string
		expected        twinInterfaceGraph
		vertexToBeAdded []VertexToBeAdded
	}{
		{
			name: "Successful add one vertex",
			vertexToBeAdded: []VertexToBeAdded{
				{
					twinInterface: twinInstance01,
					expectedError: nil,
				},
			},
			expected: twinInterfaceGraph{
				NumberOfVertex: 1,
				Vertexes: map[string]*TwinInterfaceGraphVertex{
					"TwinInterface01": {
						TwinInterface:  twinInstance01,
						EdgeInterfaces: []*TwinInterfaceGraphVertex{},
					},
				},
			},
		},
		{
			name: "Successful add two vertexes",
			vertexToBeAdded: []VertexToBeAdded{
				{
					twinInterface: twinInstance01,
					expectedError: nil,
				},
				{
					twinInterface: twinInstance02,
					expectedError: nil,
				},
			},
			expected: twinInterfaceGraph{
				NumberOfVertex: 2,
				Vertexes: map[string]*TwinInterfaceGraphVertex{
					"TwinInterface01": {
						TwinInterface:  twinInstance01,
						EdgeInterfaces: []*TwinInterfaceGraphVertex{},
					},
					"TwinInterface02": {
						TwinInterface:  twinInstance02,
						EdgeInterfaces: []*TwinInterfaceGraphVertex{},
					},
				},
			},
		},
		{
			name: "Successful add two vertexes",
			vertexToBeAdded: []VertexToBeAdded{
				{
					twinInterface: twinInstance01,
					expectedError: nil,
				},
				{
					twinInterface: twinInstance02,
					expectedError: nil,
				},
				{
					twinInterface: twinInstance01,
					expectedError: errors.New("TwinInterface already exist in the graph"),
				},
			},
			expected: twinInterfaceGraph{
				NumberOfVertex: 2,
				Vertexes: map[string]*TwinInterfaceGraphVertex{
					"TwinInterface01": {
						TwinInterface:  twinInstance01,
						EdgeInterfaces: []*TwinInterfaceGraphVertex{},
					},
					"TwinInterface02": {
						TwinInterface:  twinInstance02,
						EdgeInterfaces: []*TwinInterfaceGraphVertex{},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			twinInterfaceGraph := twinInterfaceGraph{
				NumberOfVertex: 0,
				Vertexes:       map[string]*TwinInterfaceGraphVertex{},
			}

			for _, vertex := range tt.vertexToBeAdded {
				_, err := twinInterfaceGraph.AddVertex(vertex.twinInterface)
				assert.Equal(t, vertex.expectedError, err)
			}

			assert.Equal(t, tt.expected, twinInterfaceGraph)
		})
	}
}
