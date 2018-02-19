package gf

import	(
		"sync"
		"path"
		"mime"
		"bytes"
		"reflect"
		"strings"
		"net/http"
		"io/ioutil"
		)

func (node *Node) addHandler(method string, h *Handler) {

	h.method = method
	h.Config = node.Config
	h.node = node

	node.Lock()
		node.methods[method] = h
	node.Unlock()

	node.Config.Lock()
		node.Config.activeHandlers[h] = struct{}{}
	node.Config.Unlock()

	h.GenerateClientJS()
}

type Handler struct {
	Config *Config
	node *Node
	method string
	functionKey string
	function HandlerFunction
	isFile bool
	filePath string
	fileType string
	fileCache []byte
	responseSchema interface{}
	payloadSchema interface{}
	clientJS *bytes.Buffer
	sync.RWMutex
}

func (handler *Handler) DetectContentType(req RequestInterface, filePath string) *ResponseStatus {

	if handler.fileCache == nil || !handler.node.Config.cacheFiles {

		// handle potential trailing slash on folder path declaration
		filePath := strings.Replace(filePath, "//", "/", -1)

		b, err := ioutil.ReadFile(filePath); if err != nil { return req.Respond(404, err.Error()) }

		handler.fileCache = b

		handler.fileType = mime.TypeByExtension(path.Ext(filePath))
		if handler.fileType == "" {

			handler.fileType = http.DetectContentType(b)

		}
	}

	return nil
}

func (handler *Handler) Name() string {

	var name string

	parts := strings.Split(handler.node.FullPath(), "/")

	for _, part := range parts {

		if len(part) == 0 { continue }

		if string(part[0]) == ":" { continue }

		name += strings.Title(part)

	}

	return name
}

func (handler *Handler) ApiUrl() string {

	var name string

	parts := strings.Split(handler.node.FullPath(), "/")

	for _, part := range parts {

		if len(part) == 0 { continue }

		if string(part[0]) == ":" {

			part = "'+" + part[1:] + "+'"

		}

		name += "/" + part

	}

	return "'" + name + "'"
}

// Applies model which describes request payload
func (handler *Handler) Payload(schema ...interface{}) *Handler {

	if len(schema) > 0 {

		handler.payloadSchema = schema[0]

	}

	return handler
}

// Applys model which describes response schema
func (handler *Handler) Response(schema ...interface{}) *Handler {

	handler.responseSchema = schema[0]

	return handler
}

// Validates any payload present in the request body, according to the payloadSchema
func (handler *Handler) ReadPayload(req RequestInterface) *ResponseStatus {

	// handle payload

	switch v := handler.payloadSchema.(type) {

		case nil:

			// do nothing

		case []byte:

			// do nothing

		case map[string]interface{}:

			status := req.ReadBodyObject(); if status != nil { return status }

		case []interface{}:

			status := req.ReadBodyArray(); if status != nil { return status }

		case Array, *Array:

			status := req.ReadBodyArray(); if status != nil { return status }

			var params Array

			if reflect.ValueOf(v).Kind() == reflect.Ptr {

				params = *v.(*Array)
			} else {

				params = v.(Array)
			}

			switch len(params) {

				case 1:

					return req.Respond(400, "INVALID TYPE FOR ARRAY PAYLOAD SCHEMA, EXPECTS 0 OR 2 ARGS (*ValidationConfig, paramKey)")

				case 2:

					validation, ok := params[0].(*ValidationConfig); if !ok { return req.Respond(500, "INVALID ARRAY PAYLOAD SCHEMA VALIDATION CONFIG") }

					paramKey, ok := params[1].(string); if !ok { return req.Respond(500, "INVALID ARRAY PAYLOAD SCHEMA PARAM KEY") }

					ok, array := validation.bodyFunction(req, req.BodyArray())
					if !ok {

						req.DebugJSON(req.BodyArray())
						return req.Respond(400, "INVALID TYPE FOR ARRAY PAYLOAD ITEM, EXPECTED: "+validation.Type())
					}

					req.SetParam(paramKey, array)

			}

		case Payload, *Payload:

			status := req.ReadBodyObject(); if status != nil { return status }
			
			var params Payload

			if reflect.ValueOf(v).Kind() == reflect.Ptr {

				params = *v.(*Payload)

			} else {

				params = v.(Payload)

			}

			for key, validation := range params {

				value := req.Body(key)

				ok, x := validation.bodyFunction(req, value)

				if !ok {
					req.DebugJSON(req.BodyObject())

					if value == nil {
						return req.Respond(400, "INVALID TYPE nil FOR PAYLOAD PARAMETER: "+key+", EXPECTED: "+validation.Type())						
					}

					return req.Respond(400, "INVALID TYPE " + reflect.TypeOf(value).String() + " FOR PAYLOAD PARAMETER: "+key+", EXPECTED: "+validation.Type())
				}

				req.SetParam("_" + key, x)

			}

		default:

			return req.Respond(500, "INVALID PAYLOAD SCHEMA CONFIG TYPE: "+reflect.TypeOf(v).String())

	}

	return nil
}

func (handler *Handler) UseFunction(f interface{}) {
  
  switch v := f.(type) {
    
    case string:
    
	    handler.functionKey = v
    
    case HandlerFunction:

      handler.functionKey = "?"
      handler.function = v
      
    case func(RequestInterface) *ResponseStatus:

      handler.functionKey = "?"
      handler.function = v

    default:
    
      panic("INVALID ARGUMENT TYPE FOR HANDLER METHOD FUNCTION DECLARATION")

  }
}

// Executes the modules and any hander-function, template or folder
func (handler *Handler) Handle(req RequestInterface) {

	// execute modules
	status := handler.node.RunModules(req); if status != nil { HandleStatus(req, status); return }

	// execute handler

	if handler.isFile {

		status := handler.DetectContentType(req, handler.filePath); if status != nil { HandleStatus(req, status); return }

		req.SetHeader("Content-Type", handler.fileType)

		HandleStatus(req, req.Respond(handler.fileCache))
		return
	}

	if len(handler.functionKey) == 0 { return }

  if handler.function == nil {

	  handler.function = handler.Config.GetHandlerFunction(handler.functionKey)

  }

	if handler.function == nil { panic("FAILED TO GET FUNCTION WITH KEY: "+handler.functionKey+". DID YOU ADD THE CORRECT REGISTRY & INCLUDE A REFERENCE TO "+handler.functionKey) }

	HandleStatus(req, handler.function(req))
	return
}
