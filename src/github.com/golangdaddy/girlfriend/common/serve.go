package gf

import (
		"strings"
		)

const	(
		ROBOTS_TXT = "User-agent: *\nDisallow: /api/"
		)

// main handler
func (node *Node) MainHandler(req RequestInterface, fullPath string) {

	// enforce https-only if required
	if node.Config.forceTLS {

		if !req.IsTLS() {

			HandleStatus(req, req.Respond(502, "PLEASE UPGRADE TO HTTPS"))
			return
		}
	}

	// set CORS headers
	for k, v := range node.Config.headers { req.SetHeader(k, v) }

	// return if preflight request
	if req.Method() == "OPTIONS" { return }

	switch fullPath {

		case "/_.js":

			req.SetHeader("Content-Type", "application/javascript")
			
			for handler, _ := range node.Config.activeHandlers {

				if handler.clientJS == nil { continue }

				req.Write(handler.clientJS.Bytes())

			}
			
			return

		case "/_.json":

			// render the handler documentation

			spec := node.Config.buildSpec()

			HandleStatus(req, req.Respond(spec))
			return

		case "/robots.txt":

			req.Write([]byte(ROBOTS_TXT))
			return

		default:

			rootFunc := node.Config.GetRootFunction(fullPath)

			if rootFunc != nil {

				status := node.RunModule(req); if status != nil { HandleStatus(req, status); return }
				status = node.RunModules(req); if status != nil { HandleStatus(req, status); return }

				HandleStatus(req, rootFunc(req))
				return

			}

	}

	segments := strings.Split(fullPath, "/")[1:]

	next := node

	for _, segment := range segments {

		if len(segment) == 0 { break }

		n, status := next.Next(req, segment); if status != nil { HandleStatus(req, status); return }

		if n != nil {

			for k, v := range n.requestParams { req.SetParam(k, v) }

			next = n
			continue

		}

		req.HttpError("NO ROUTE FOUND AT " + next.FullPath() + "/" + segment, 404)
		return
	}

	// resolve handler

	handler := next.handler(req)

	if handler == nil {

		req.HttpError("NO CONTROLLER FOUND AT " + next.FullPath(), 500)
		return

	}

	// read the request body and unmarshal into specified schema
	status := handler.ReadPayload(req); if status != nil { HandleStatus(req, status); return }

	handler.Handle(req)

}