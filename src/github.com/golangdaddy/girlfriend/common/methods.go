package gf

// Allows POST requests to the node's handler
func (node *Node) HEAD(functionKeys ...interface{}) *Handler {

	handler := &Handler{}
	
	if len(functionKeys) > 0 { handler.UseFunction(functionKeys[0]) }

	node.addHandler("HEAD", handler)

	return handler
}

// Allows GET requests to the node's handler
func (node *Node) GET(functionKeys ...interface{}) *Handler {

	handler := &Handler{}
	
	if len(functionKeys) > 0 { handler.UseFunction(functionKeys[0]) }

	node.addHandler("GET", handler)

	return handler
}

// Allows POST requests to the node's handler
func (node *Node) POST(functionKeys ...interface{}) *Handler {

	handler := &Handler{}
	
	if len(functionKeys) > 0 { handler.UseFunction(functionKeys[0]) }

	node.addHandler("POST", handler)

	return handler
}

// Allows PUT requests to the node's handler
func (node *Node) PUT(functionKeys ...interface{}) *Handler {

	handler := &Handler{}
	
	if len(functionKeys) > 0 { handler.UseFunction(functionKeys[0]) }

	node.addHandler("PUT", handler)

	return handler
}

// Allows POST requests to the node's handler
func (node *Node) DELETE(functionKeys ...interface{}) *Handler {

	handler := &Handler{}
	
	if len(functionKeys) > 0 { handler.UseFunction(functionKeys[0]) }

	node.addHandler("DELETE", handler)

	return handler
}

// Allows POST requests to the node's handler
func (node *Node) PATCH(functionKeys ...interface{}) *Handler {

	handler := &Handler{}
	
	if len(functionKeys) > 0 { handler.UseFunction(functionKeys[0]) }

	node.addHandler("PATCH", handler)

	return handler
}