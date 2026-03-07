package model

type API struct {
	Endpoints map[string]Endpoint
}

type Endpoint struct {
	Method string
	Path   string

	Params         map[ParamKey]Param
	Response200    SchemaObject
	HasResponse200 bool
	RequestBody    SchemaObject
	HasRequestBody bool
}

type ParamKey struct {
	In   string
	Name string
}

type Param struct {
	In       string
	Name     string
	Required bool
	Type     string
	Enum     []string
}

type SchemaObject struct {
	Fields   map[string]Field
	Required map[string]bool
}

type Field struct {
	Type string
	Enum []string
}
