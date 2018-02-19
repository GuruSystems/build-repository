package gf

import 	(
		"sync"
		"html"
		"bytes"
		"strings"
		//
		"github.com/microcosm-cc/bluemonday"
		)

type Config struct {
	Host string
	RootRegistry Registry
	HandlerRegistry Registry
	ModuleRegistry ModuleRegistry
	SubdomainTrees map[string]*Node
	headers Headers
	lDelim, rDelim string
	activeHandlers map[*Handler]struct{}
	countries map[string]*Country
	reverseCountries map[string]*Country
	languages map[string]*Language
	sanitizer *bluemonday.Policy
	cacheFiles bool
	forceTLS bool
	sync.RWMutex
}

func (config *Config) GenerateClientJS() []byte {

	a := [][]byte{}

	for handler, _ := range config.activeHandlers {

		if handler.clientJS == nil { continue }

		a = append(a, handler.clientJS.Bytes())
	}

	return bytes.Join(a, nil)
}

func (config *Config) NoCache() {

	config.Lock()
		config.cacheFiles = false
	config.Unlock()
}

func (config *Config) SubTree(subdomain string) *Node {

	config.RLock()
		tree := config.SubdomainTrees[subdomain]
	config.RUnlock()

	if tree == nil {

		tree = rootNode()

		newConfig := *config

		tree.Config = &newConfig

		config.Lock()
			config.SubdomainTrees[subdomain] = tree
		config.Unlock()
	}

	return tree
}

func (config *Config) Sanitize(s string) string {

	return config.sanitizer.Sanitize(html.UnescapeString(s))
}

// block all non-https requests
func (config *Config) ForceTLS() {

	config.Lock()
		config.forceTLS = true
	config.Unlock()
}


func (config *Config) SetDelims(l, r string) {

	config.Lock()
		config.lDelim = l
		config.rDelim = r
	config.Unlock()
}

// Sets the root registry to the specified map
func (config *Config) SetRootRegistry(reg Registry) {

	config.Lock()
		config.RootRegistry = reg
	config.Unlock()
}

// Sets the handler registry to the specified map
func (config *Config) SetHandlerRegistry(reg Registry) {

	config.Lock()
		config.HandlerRegistry = reg
	config.Unlock()
}

// Adds the handler registry's values to the config
func (config *Config) AddHandlerRegistry(reg Registry) {

	config.Lock()
		for k, v := range reg { config.HandlerRegistry[k] = v }
	config.Unlock()
}

// Sets the module registry to the specified map
func (config *Config) SetModuleRegistry(reg ModuleRegistry) {

	config.Lock()
		config.ModuleRegistry = reg
	config.Unlock()
}

// Adds the module registry's values to the config
func (config *Config) AddModuleRegistry(reg ModuleRegistry) {

	config.Lock()
		for k, v := range reg { config.ModuleRegistry[k] = v }
	config.Unlock()
}

// Sets the http preflight headers to the specified map
func (config *Config) SetHeaders(h Headers) {

	config.Lock()
		config.headers = h
	config.Unlock()
}

// Returns the root function if present in the registry
func (config *Config) GetRootFunction(functionKey string) HandlerFunction {

	if config.RootRegistry == nil { return nil }

	config.RLock()
		function := config.RootRegistry[functionKey]
	config.RUnlock()

	return function
}

// Returns the handler function if present in the registry
func (config *Config) GetHandlerFunction(functionKey string) HandlerFunction {

	config.RLock()
		function := config.HandlerRegistry[functionKey]
	config.RUnlock()

	return function
}

// Returns the handler function if present in the registry
func (config *Config) GetModuleFunction(functionKey string) ModuleFunction {

	config.RLock()
		function := config.ModuleRegistry[functionKey]
	config.RUnlock()

	return function
}

// Accesses the countries map to return country struct if exists
func (config *Config) GetCountry(countryCode string) *Country {

	config.RLock()
		c := config.countries[countryCode]
	config.RUnlock()

	return c
}

// Accesses the countries map to return country struct if exists
func (config *Config) ReverseGetCountry(countryName string) *Country {

	countryName = strings.ToLower(countryName)

	config.RLock()
		c := config.reverseCountries[countryName]
	config.RUnlock()

	return c
}