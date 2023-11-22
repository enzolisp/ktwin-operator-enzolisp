package dtdl

import "github.com/Open-Digital-Twin/ktwin-operator/cmd/cli/types"

// This file stores the `Digital Twins Definition Language` Models in JSON Format
// https://github.com/Azure/opendigitaltwins-dtdl

type DTMI string
type IRI string
type LocalizedString string

type TwinDefinition struct {
	Id          string                  `json:"@id"`
	Type        string                  `json:"@type"`
	DisplayName string                  `json:"displayName"`
	Description string                  `json:"description"`
	Comment     string                  `json:"comment"`
	Contents    []TwinDefinitionContent `json:"contents"`
	Context     string                  `json:"@context"`
	Extends     []string                `json:"extends"`
}

type TwinDefinitionContent struct {
	Type      string `json:"@type"`
	Name      string `json:"name"`
	Comment   string `json:"comment"`
	Schema    Schema `json:"schema"`
	Unit      string `json:"unit"`
	Writeable bool   `json:"writable"`
}

// DTDL Types

type Interface struct {
	Context     IRI               `json:"@context"`
	Type        IRI               `json:"@type"`
	Id          DTMI              `json:"@id"`
	Comment     string            `json:"comment,omitempty"`
	Contents    []Content         `json:"contents,omitempty"`
	Description LocalizedString   `json:"description,omitempty"`
	DisplayName LocalizedString   `json:"displayName,omitempty"`
	Extends     types.StringArray `json:"extends,omitempty"` // Error: it must be Interface according to the definition
	Schemas     []Schema          `json:"schemas"`
}

type Telemetry struct {
	Type        IRI             `json:"@type"`
	Id          DTMI            `json:"@id,omitempty"`
	Comment     string          `json:"comment,omitempty"`
	Description LocalizedString `json:"description,omitempty"`
	DisplayName LocalizedString `json:"displayName,omitempty"`
	Name        string          `json:"name"`
	Schema      Schema          `json:"schema"`
}

type Property struct {
	Type        IRI             `json:"@type"`
	Id          DTMI            `json:"@id,omitempty"`
	Comment     string          `json:"comment,omitempty"`
	Description LocalizedString `json:"description,omitempty"`
	DisplayName LocalizedString `json:"displayName,omitempty"`
	Name        string          `json:"name"`
	Schema      Schema          `json:"schema"`
	Writeable   bool            `json:"writable"`
}

type Command struct {
	Type        IRI             `json:"@type"`
	Id          DTMI            `json:"@id,omitempty"`
	Comment     string          `json:"comment,omitempty"`
	CommandType string          `json:"commandType,omitempty"` // Deprecated
	Description LocalizedString `json:"description,omitempty"`
	DisplayName LocalizedString `json:"displayName,omitempty"`
	Name        string          `json:"name"`
	Request     CommandRequest  `json:"request"`
	Response    CommandResponse `json:"response"`
}

type CommandRequest struct {
	Type        IRI             `json:"@type"`
	Id          DTMI            `json:"@id,omitempty"`
	Comment     string          `json:"comment,omitempty"`
	Description LocalizedString `json:"description,omitempty"`
	DisplayName LocalizedString `json:"displayName,omitempty"`
	Name        string          `json:"name"`
	Schema      Schema          `json:"schema"`
}

type CommandResponse struct {
	Type        IRI             `json:"@type"`
	Id          DTMI            `json:"@id,omitempty"`
	Comment     string          `json:"comment,omitempty"`
	Description LocalizedString `json:"description,omitempty"`
	DisplayName LocalizedString `json:"displayName,omitempty"`
	Name        string          `json:"name"`
	Schema      Schema          `json:"schema"`
}

type Relationship struct {
	Type            IRI             `json:"@type"`
	Id              DTMI            `json:"@id,omitempty"`
	Comment         string          `json:"comment,omitempty"`
	Description     LocalizedString `json:"description,omitempty"`
	DisplayName     LocalizedString `json:"displayName,omitempty"`
	MaxMultiplicity int             `json:"maxMultiplicity,omitempty"`
	MinMultiplicity int             `json:"minMultiplicity,omitempty"`
	Name            string          `json:"name"`
	Properties      []Property      `json:"properties"`
	Target          DTMI            `json:"target"`
	Schema          Schema          `json:"schema"`
	Writeable       bool            `json:"writeable"`
}

type Component struct {
	Type        IRI             `json:"@type"`
	Id          DTMI            `json:"@id,omitempty"`
	Comment     string          `json:"comment,omitempty"`
	Description LocalizedString `json:"description,omitempty"`
	DisplayName LocalizedString `json:"displayName,omitempty"`
	Name        string          `json:"name"`
	Schema      Schema          `json:"schema"`
}
