package graph

import (
	"errors"
	"testing"

	dtdv0 "github.com/Open-Digital-Twin/ktwin-operator/api/dtd/v0"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"
)

var twinInstance01 = dtdv0.TwinInstance{
	ObjectMeta: v1.ObjectMeta{
		Name: "TwinInstance01",
	},
	Spec: dtdv0.TwinInstanceSpec{
		TwinInstanceRelationships: []dtdv0.TwinInstanceRelationship{
			{
				Name:      "NameRelationship01",
				Interface: "InterfaceRelationship01",
				Instance:  "InstanceRelationship02",
			},
		},
	},
}

var twinInstance02 = dtdv0.TwinInstance{
	ObjectMeta: v1.ObjectMeta{
		Name: "TwinInstance02",
	},
	Spec: dtdv0.TwinInstanceSpec{
		TwinInstanceRelationships: []dtdv0.TwinInstanceRelationship{
			{
				Name:      "NameRelationship02",
				Interface: "InterfaceRelationship02",
				Instance:  "InstanceRelationship02",
			},
		},
	},
}

func TestTwinInstanceImplements_CreateGraph(t *testing.T) {
	t.Run("Should implement TwinInstanceGraph", func(t *testing.T) {
		twinInstanceGraph := NewTwinInstanceGraph()
		assert.Implements(t, (*TwinInstanceGraph)(nil), twinInstanceGraph)
	})
}

func TestTwinInstance_CreateGraph(t *testing.T) {
	t.Run("Should create TwinInstance Graph", func(t *testing.T) {
		graph := NewTwinInstanceGraph()
		assert.Equal(t, &twinInstanceGraph{
			NumberOfVertex: 0,
			Vertexes:       map[string]*TwinInstanceGraphVertex{},
		}, graph)
	})
}

func TestTwinInstance_AddVertex(t *testing.T) {

	type VertexToBeAdded struct {
		twinInstance  dtdv0.TwinInstance
		expectedError error
	}

	tests := []struct {
		name            string
		expected        twinInstanceGraph
		vertexToBeAdded []VertexToBeAdded
	}{
		{
			name: "Successful add one vertex",
			vertexToBeAdded: []VertexToBeAdded{
				{
					twinInstance:  twinInstance01,
					expectedError: nil,
				},
			},
			expected: twinInstanceGraph{
				NumberOfVertex: 1,
				Vertexes: map[string]*TwinInstanceGraphVertex{
					"TwinInstance01": {
						TwinInstance:  twinInstance01,
						EdgeInstances: []*TwinInstanceGraphVertex{},
					},
				},
			},
		},
		{
			name: "Successful add two vertexes",
			vertexToBeAdded: []VertexToBeAdded{
				{
					twinInstance:  twinInstance01,
					expectedError: nil,
				},
				{
					twinInstance:  twinInstance02,
					expectedError: nil,
				},
			},
			expected: twinInstanceGraph{
				NumberOfVertex: 2,
				Vertexes: map[string]*TwinInstanceGraphVertex{
					"TwinInstance01": {
						TwinInstance:  twinInstance01,
						EdgeInstances: []*TwinInstanceGraphVertex{},
					},
					"TwinInstance02": {
						TwinInstance:  twinInstance02,
						EdgeInstances: []*TwinInstanceGraphVertex{},
					},
				},
			},
		},
		{
			name: "Successful add two vertexes",
			vertexToBeAdded: []VertexToBeAdded{
				{
					twinInstance:  twinInstance01,
					expectedError: nil,
				},
				{
					twinInstance:  twinInstance02,
					expectedError: nil,
				},
				{
					twinInstance:  twinInstance01,
					expectedError: errors.New("TwinInstance already exist in the graph"),
				},
			},
			expected: twinInstanceGraph{
				NumberOfVertex: 2,
				Vertexes: map[string]*TwinInstanceGraphVertex{
					"TwinInstance01": {
						TwinInstance:  twinInstance01,
						EdgeInstances: []*TwinInstanceGraphVertex{},
					},
					"TwinInstance02": {
						TwinInstance:  twinInstance02,
						EdgeInstances: []*TwinInstanceGraphVertex{},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			twinInstanceGraph := twinInstanceGraph{
				NumberOfVertex: 0,
				Vertexes:       map[string]*TwinInstanceGraphVertex{},
			}

			for _, vertex := range tt.vertexToBeAdded {
				_, err := twinInstanceGraph.AddVertex(vertex.twinInstance)
				assert.Equal(t, vertex.expectedError, err)
			}

			assert.Equal(t, tt.expected, twinInstanceGraph)
		})
	}
}

func TestTwinInstance_MarshalJSON(t *testing.T) {

	type VertexToBeAdded struct {
		twinInstance  dtdv0.TwinInstance
		expectedError error
	}

	tests := []struct {
		name                     string
		initialTwinInstanceGraph twinInstanceGraph
		expectedResult           string
	}{
		{
			name: "Successful add one vertex",
			initialTwinInstanceGraph: twinInstanceGraph{
				NumberOfVertex: 1,
				Vertexes: map[string]*TwinInstanceGraphVertex{
					"TwinInstance01": {
						TwinInstance:  twinInstance01,
						EdgeInstances: []*TwinInstanceGraphVertex{},
					},
				},
			},
			expectedResult: "[{\"metadata\":{\"name\":\"TwinInstance01\",\"creationTimestamp\":null},\"spec\":{\"twinInstanceRelationships\":[{\"name\":\"NameRelationship01\",\"interface\":\"InterfaceRelationship01\",\"instance\":\"InstanceRelationship02\"}]},\"status\":{}}]",
		},
		{
			name: "Successful add two vertexes",
			initialTwinInstanceGraph: twinInstanceGraph{
				NumberOfVertex: 2,
				Vertexes: map[string]*TwinInstanceGraphVertex{
					"TwinInstance01": {
						TwinInstance:  twinInstance01,
						EdgeInstances: []*TwinInstanceGraphVertex{},
					},
					"TwinInstance02": {
						TwinInstance:  twinInstance02,
						EdgeInstances: []*TwinInstanceGraphVertex{},
					},
				},
			},
			expectedResult: "[{\"metadata\":{\"name\":\"TwinInstance01\",\"creationTimestamp\":null},\"spec\":{\"twinInstanceRelationships\":[{\"name\":\"NameRelationship01\",\"interface\":\"InterfaceRelationship01\",\"instance\":\"InstanceRelationship02\"}]},\"status\":{}},{\"metadata\":{\"name\":\"TwinInstance02\",\"creationTimestamp\":null},\"spec\":{\"twinInstanceRelationships\":[{\"name\":\"NameRelationship02\",\"interface\":\"InterfaceRelationship02\",\"instance\":\"InstanceRelationship02\"}]},\"status\":{}}]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.initialTwinInstanceGraph.MarshalJson()

			if err != nil {
				assert.Fail(t, "Error is not supposed to be different of nul")
			}

			assert.Equal(t, tt.expectedResult, result)
		})
	}
}
