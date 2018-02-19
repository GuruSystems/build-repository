package gf

import	(
		"os"
		"fmt"
		"sync"
		"strings"
		)

type Node struct {
	Config *Config
	parent *Node
	path string
	param *Node
	requestParams map[string]interface{}
	routes map[string]*Node
	methods map[string]*Handler
	module *Module
	modules []*Module
	validation *ValidationConfig
	validations []*ValidationConfig
	sync.RWMutex
}

func (node *Node) new(path string) *Node {

	n := &Node{
		parent:			      node,
		Config:			      node.Config,
		requestParams:    map[string]interface{}{},
		routes:			  		map[string]*Node{},
		methods:			  	map[string]*Handler{},
		modules:				  node.modules,
		// inherited properties
		path: 					  path,
		validations:			node.validations,
	}

	return n
}

// returns the base param map including node params
func (node *Node) RequestParams() map[string]interface{} {

  m := map[string]interface{}{}

  node.RLock()
  
  for k, v := range node.requestParams { m[k] = v }
  
  node.RUnlock()

  return m
}

// Returns the node's path string
func (node *Node) Path() string {

	return node.path
}

// Returns the node's full path string
func (node *Node) FullPath() string {

	if node == nil { return "" }

	parent := node

	path := node.path
	
	for {
		parent = parent.parent

		if parent == nil { break }

		path = parent.path + "/" + path
	}

	if len(path) == 0 { path = "/" }

	return path
}

// Adds a new node to the tree
func (node *Node) Add(path string, pathKeys ...string) *Node {

	path = strings.TrimSpace(strings.Replace(path, "/", "", -1))

	if existing := node.routes[path]; existing != nil { return existing }

	n := node.new(path)

	node.Lock()
		
		node.routes[path] = n
		
		if len(pathKeys) > 0 {
		  for _, key := range pathKeys { n.requestParams[key] = path }
		}
	
	node.Unlock()

	return n
}

// Adds a new param-node
func (node *Node) Param(config *ValidationConfig, keys ...string) *Node {

	if len(keys) == 0 { panic("NO KEYS SUPPLIED FOR NEW PARAMETER") }

	node.RLock()
		p := node.param
	node.RUnlock()
	
	if p != nil { return p }

	n := node.new(":" + keys[0])

	config.keys = keys

	n.Lock()
		n.validation = config
		n.validations = append(n.validations, config)
	n.Unlock()

	node.Lock()
		node.param = n
	node.Unlock()

	return n
}

func (node *Node) newModule(functionKey string, arg interface{}) *Module {

	if node.Config.ModuleRegistry == nil { panic("Config has no ModuleRegistry setting!") }

	return &Module{
		config:					node.Config,
		functionKey:			functionKey,
		arg:					arg,
	}
}

// Adds a module that will be executed at the point it is added to the route
func (node *Node) Init(functionKey string, arg interface{}) *Node {

	module := node.newModule(functionKey, arg)

	node.Lock()
		node.module = module
	node.Unlock()

	return node
}


// Adds a module that will be executed upon reaching a handler
func (node *Node) Mod(functionKey string, arg interface{}) *Node {

	module := node.newModule(functionKey, arg)

	node.Lock()
		node.modules = append(node.modules, module)
	node.Unlock()

	return node
}

// execute init function added with .Init(...)
func (node *Node) RunModule(req RequestInterface) *ResponseStatus {

	node.RLock()
		module := node.module
	node.RUnlock()

	if module != nil {

		status := module.Run(req); if status != nil { return status }
	}

	return nil
}

// execute all module functions added with .Mod(...)
func (node *Node) RunModules(req RequestInterface) *ResponseStatus {
	
	for _, module := range node.modules {

		status := module.Run(req); if status != nil { return status }
	}

	return nil
}

// traversal

// finds next node according to supplied URL path segment
func (node *Node) Next(req RequestInterface, pathSegment string) (*Node, *ResponseStatus) {

	// execute any init module(s)

	node.RunModule(req)

	// check for child routes

	next := node.routes[pathSegment]

	if next != nil { return next, nil }

	// check for path param

	next = node.param; if next == nil { return nil, nil }

	if next.validation != nil {

		ok, value := next.validation.pathFunction(req, pathSegment); if !ok {

			return nil, &ResponseStatus{nil, 400, fmt.Sprintf("UNEXPECTED VALUE  %v, %v", pathSegment, next.validation.Expecting())}

		}

		// write route params into request object

		for _, key := range next.validation.keys { req.SetParam(key, value) }

	}

	return next, nil
}

// Returns the handler assciated with the HTTP request method.
func (node *Node) handler(req RequestInterface) *Handler {

	node.RLock()

		handler := node.methods[req.Method()]

	node.RUnlock()

	return handler
}

// Adds a file to be served from the specified path.
func (node *Node) File(path string) *Node {

	h := &Handler{
		isFile:					true,
		filePath:			path,
	}

	node.addHandler("GET", h)

	return node
}

// Walks through the specified folder to mirror the file structure for files containing all filters
func (node *Node) Folder(directoryPath string, filters ...string) {

	// remove trailing slash from the directory path if existing
	directoryPath = strings.TrimSuffix(directoryPath, "/")

	f, err := os.Open(directoryPath); if err != nil { panic(err) }

	names, err := f.Readdirnames(-1)
	f.Close()
	if err != nil { panic(err) }

	for _, name := range names {

		path := strings.Replace(directoryPath + "/" + name, "//", "/", -1)

		node.checkFile(name, path, filters)

	}

}

// Checks if file or folder, adding any files
func (node *Node) checkFile(name, path string, filters []string) {

	info, err := os.Lstat(path); if err != nil { panic(err) }

	if info.IsDir() {

		node.Add(name).Folder(path)
		return

	}

	for _, filter := range filters {

		if !strings.Contains(name, filter) { return }

	}

	node.Add(name).File(path)

}