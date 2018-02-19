package gf

type Router interface{
  Serve(int)
}

type Array []interface{}
type Object map[string]interface{}

type Headers map[string]string

type HandlerFunction func (req RequestInterface) *ResponseStatus

type Registry map[string]HandlerFunction

type ModuleFunction func (RequestInterface, interface{}) *ResponseStatus

type ModuleRegistry map[string]ModuleFunction

type Payload map[string]*ValidationConfig

func (payload Payload) WithFields(fields Payload) Payload {
  
  for k, v := range payload { fields[k] = v }

  return fields
}