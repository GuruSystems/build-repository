package gf

import 	(
		)

type Module struct {
	config *Config
	functionKey string
	function ModuleFunction
	arg interface{}
}

func (mod *Module) Run(req RequestInterface) *ResponseStatus {

	if mod.function == nil {

		mod.function = mod.config.GetModuleFunction(mod.functionKey)
		
		if mod.function == nil { return req.Respond(500, "MODULE NOT FOUND WITH KEY: "+mod.functionKey) }

	}

	return mod.function(req, mod.arg)
} 

