// Package gf is the common folder for implementations of the Girlfriend http router
package gf

import 	(
		"fmt"
		//
		"github.com/microcosm-cc/bluemonday"
		)

// for internal debugging use
func debug(s string) { fmt.Println(s) }

var globalNode *Node

func rootNode() *Node {

	return &Node{
		routes:			      map[string]*Node{},
		methods:		      map[string]*Handler{},
		requestParams:    map[string]interface{}{},
		modules:		      []*Module{},
		validations:	    []*ValidationConfig{},
	}

}

func init() {

	globalNode = rootNode()

	globalNode.Config = &Config{
		cacheFiles:			true,
		SubdomainTrees:		map[string]*Node{},
		activeHandlers:		map[*Handler]struct{}{},
		countries:			Countries(),
		reverseCountries:	ReverseCountries(),
		languages:			Languages(),
		sanitizer:			bluemonday.StrictPolicy(),
		lDelim:				"{{",
		rDelim:				"}}",
	}

}

func Root() *Node {

	return globalNode
}
