package gf

import	(
		"sort"
		"strings"
		)

type ValidationConfigSpec struct {
	Type string
	Keys []string
}

func (vc *ValidationConfig) Spec() *ValidationConfigSpec {

	spec := &ValidationConfigSpec{
		Type:		vc.Type(),
		Keys:		vc.keys,
	}

	return spec
}

type HandlerSpec struct {
	Name string
	Method string
	Function string
	Endpoint string
	PayloadSchema map[string]string
	ResponseSchema interface{}
	RouteParams []*ValidationConfigSpec
	IsFile bool
	FilePath string
}

func (handler *Handler) Spec() *HandlerSpec {

	validations := []*ValidationConfigSpec{}

	for _, vc := range handler.node.validations { validations = append(validations, vc.Spec()) }

	// build useful payloadSchema for the spec

	payloadSchema := map[string]string{}

	switch data := handler.payloadSchema.(type) {

		case Payload:

			for k, v := range data { payloadSchema[k] = v.Type() }

		case Array:

	}

	spec := &HandlerSpec{
		Name:					handler.Name(),
		Method:					handler.method,
		Function:				handler.functionKey,
		Endpoint:				handler.node.FullPath(),
		PayloadSchema:			payloadSchema,
		ResponseSchema:			handler.responseSchema,
		IsFile:					handler.isFile,
		FilePath:				handler.filePath,
		RouteParams:			validations,
	}

	return spec
}

type HandlerArray []*HandlerSpec

func (a HandlerArray) Len() int { return len(a) }

func (a HandlerArray) Swap(x, y int) { a[x], a[y] = a[y], a[x] }

func (a HandlerArray) Less(x, y int) bool {

	if strings.Compare(a[x].Endpoint, a[y].Endpoint) == 1 { return true }

	return false
}

// Builds the handler documentation object.
func (config *Config) buildSpec() []*HandlerSpec {

	a := HandlerArray{}

	config.RLock()

	for handler, _ := range config.activeHandlers { a = append(a, handler.Spec()) }

	config.RUnlock()

	sort.Sort(a)

	return a
}