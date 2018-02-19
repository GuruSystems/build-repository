package gf

import 	(
    "fmt"
    "time"
    "html"
	"reflect"
	"strconv"
	"strings"
	"unicode"
	"net/mail"
	"net/url"
	"encoding/hex"
	//
	"github.com/ttacon/libphonenumber"
	//
	"golang.org/x/crypto/sha3"
)

const 	(
		STRING_MAX_LENGTH = 2000
		USERNAME_CHARS = "0123456789_abcdefghijklmnopqrstuvwxyz"
		)

func Hash256(input string) string {

	b := make([]byte, 64)

	sha3.ShakeSum256(b, []byte(input))

	return hex.EncodeToString(b)
}

type BodyValidationFunction func (RequestInterface, interface{}) (bool, interface{})
type PathValidationFunction func (RequestInterface, string) (bool, interface{})

type ValidationConfig struct {
	Model interface{}
	pathFunction PathValidationFunction
	bodyFunction BodyValidationFunction
	keys []string
	min float64
	max float64
}

func (vc *ValidationConfig) Key() string {

	return vc.keys[0]
}

func (vc *ValidationConfig) Keys() string {

	return strings.Join(vc.keys, "_")
}

func (vc *ValidationConfig) Type() string {

	return reflect.TypeOf(vc.Model).String()
}

func (vc *ValidationConfig) Expecting() string {

	return "expecting: " + vc.Type() + " for keys: "+strings.Join(vc.keys, ", ")
}

func NewValidationConfig(validationType interface{}, pathFunction PathValidationFunction, bodyFunction BodyValidationFunction, ranges ...float64) *ValidationConfig {

	cfg := &ValidationConfig{
		Model: validationType,
		pathFunction: pathFunction,
		bodyFunction: bodyFunction,

	}

	switch len(ranges) {
		
		case 2:

			cfg.min = ranges[0]
			cfg.max = ranges[1]

	}

	return cfg
}

// Returns a validation object that checks for a string with a length within optional range
func String(ranges ...float64) *ValidationConfig {

	var min, max float64

	switch len(ranges) {

		case 0:

			max = STRING_MAX_LENGTH

		case 1:

			min = ranges[0]
			max = ranges[0]

		case 2:

			min = ranges[0]
			max = ranges[1]

	}

	config := NewValidationConfig(
		"",
		func (req RequestInterface, param string) (bool, interface{}) {

			lp := float64(len(param))

			if lp < min || lp > max { return false, nil }

			return true, html.UnescapeString(strings.TrimSpace(globalNode.Config.Sanitize(param)))
		},
		func (req RequestInterface, param interface{}) (bool, interface{}) {

			if min == 0 && param == nil { return true, "" }

			s, ok := param.(string); if !ok { return false, nil }

			lp := float64(len(s))

			if lp < min || lp > max { return false, nil }

			return true, html.UnescapeString(strings.TrimSpace(globalNode.Config.Sanitize(s)))
		},
	)

	config.min = min
	config.max = max

	return config
}

// Returns a validation object which checks for delimiter-separated string like CSV
func SplitString(delimiter string) *ValidationConfig {

	return NewValidationConfig(
		[]string{},
		func (req RequestInterface, param string) (bool, interface{}) {

			lp := len(param)

			if lp == 0 || lp > STRING_MAX_LENGTH { return false, nil }
			list := []string{}

			for _, part := range strings.Split(globalNode.Config.Sanitize(param), delimiter) {

				if len(part) == 0 { continue }

				list = append(list, part)

			}

			return true, list
		},
		func (req RequestInterface, param interface{}) (bool, interface{}) {

			s, ok := param.(string); if !ok { return false, nil }

			lp := len(s)
			if lp == 0 || lp > STRING_MAX_LENGTH { return false, nil }
			
			list := []string{}

			for _, part := range strings.Split(globalNode.Config.Sanitize(s), delimiter) {

				if len(part) == 0 { continue }

				list = append(list, part)

			}

			return true, list
		},
	)
}

// Returns a validation object which ensures string is whitelisted
func StringSet(set ...string) *ValidationConfig {

  filter := map[string]bool{}
  
  for _, item := range set { filter[item] = true  }

	return NewValidationConfig(
		"",
		func (req RequestInterface, param string) (bool, interface{}) {

      		param = strings.TrimSpace(strings.ToLower(param))

			if !filter[param] { return false, nil }

			return true, param
		},
		func (req RequestInterface, param interface{}) (bool, interface{}) {

			s, ok := param.(string); if !ok { return false, nil }

			if !filter[s] { return false, nil }

			return true, s
		},
	)
}

// Returns a validation object which checks for valid email address
func EmailAddress() *ValidationConfig {

	return NewValidationConfig(
		"",
		func (req RequestInterface, param string) (bool, interface{}) {

      		param = strings.TrimSpace(strings.ToLower(param))

			_, err := mail.ParseAddress(param); if err != nil { return false, nil }

			return true, param
		},
		func (req RequestInterface, param interface{}) (bool, interface{}) {

			s, ok := param.(string); if !ok { return false, nil }

      		s = strings.TrimSpace(strings.ToLower(s))

			_, err := mail.ParseAddress(s); if err != nil { return false, nil }

			return true, s
		},
	)
}

// Returns a validation object which checks for valid email address
func OptionalEmailAddress() *ValidationConfig {

	return NewValidationConfig(
		"",
		func (req RequestInterface, param string) (bool, interface{}) {

      		param = strings.TrimSpace(strings.ToLower(param))

      		if len(param) == 0 { return true, "" }

			_, err := mail.ParseAddress(param); if err != nil { return false, nil }

			return true, param
		},
		func (req RequestInterface, param interface{}) (bool, interface{}) {

			s, ok := param.(string); if !ok { return false, nil }

      		s = strings.TrimSpace(strings.ToLower(s))

      		if len(s) == 0 { return true, "" }

			_, err := mail.ParseAddress(s); if err != nil { return false, nil }

			return true, s
		},
	)
}

type JSON struct {}

// Returns a validation object which checks for valid email address
func Json() *ValidationConfig {

	return NewValidationConfig(
		JSON{},
		func (req RequestInterface, param string) (bool, interface{}) {

			return true, param
		},
		func (req RequestInterface, param interface{}) (bool, interface{}) {

			s, ok := param.(string); if !ok { return false, nil }

			return true, s
		},
	)
}

// Returns a validation object which checks for valid time
func SQLTimestamp() *ValidationConfig {

	return NewValidationConfig(
		time.Now(),
		func (req RequestInterface, param string) (bool, interface{}) {

			return true, param
		},
		func (req RequestInterface, param interface{}) (bool, interface{}) {

			return true, param
		},
	)
}

// Returns a validation object which checks for valid time
func JSTime() *ValidationConfig {

	return NewValidationConfig(
		"",
		func (req RequestInterface, param string) (bool, interface{}) {

			ok, s := verifyJSTime(req, param); if !ok { return false, nil }

			return true, s
		},
		func (req RequestInterface, param interface{}) (bool, interface{}) {

			s, ok := param.(string); if !ok { return false, nil }

			ok, s = verifyJSTime(req, s); if !ok { return false, nil }

			return true, s
		},
	)
}

func verifyJSTime(req RequestInterface, input string) (bool, string) {

    _, err := time.Parse(time.RFC3339Nano, input); if err != nil { req.Error(err); return false, "" }

    return true, input
}

type URL struct{}

// Returns a validation object which checks for valid url
func Url() *ValidationConfig {

	return NewValidationConfig(
		"",
		func (req RequestInterface, param string) (bool, interface{}) {

      		param = strings.TrimSpace(param)

			_, err := url.ParseRequestURI(param); if err != nil { return false, "" }

			return true, param
		},
		func (req RequestInterface, param interface{}) (bool, interface{}) {

			s, ok := param.(string); if !ok { return false, nil }

      		s = strings.TrimSpace(strings.ToLower(s))

			_, err := url.ParseRequestURI(s); if err != nil { return false, "" }

			return true, s
		},
	)
}

// Returns a validation object which checks for valid username
func Username(max int) *ValidationConfig {

	return NewValidationConfig(
		"",
		func (req RequestInterface, s string) (bool, interface{}) {

      		s = strings.TrimSpace(strings.ToLower(s))

			for _, char := range s {

				if !strings.Contains(USERNAME_CHARS, string(char)) { return false, nil }

			}

			return (len(s) >= 3 && len(s) <= max), s
		},
		func (req RequestInterface, param interface{}) (bool, interface{}) {

			s, ok := param.(string); if !ok { return false, nil }

      		s = strings.TrimSpace(strings.ToLower(s))

			for _, char := range s {

				if !strings.Contains(USERNAME_CHARS, string(char)) { return false, nil }

			}

			return (len(s) >= 3 && len(s) <= max), s
		},
	)
}

// Returns a validation object which checks for password
func Password(hard bool) *ValidationConfig {

	return NewValidationConfig(
		"",
		func (req RequestInterface, param string) (bool, interface{}) {

			if hard {
				if !verifyHardPassword(param) { return false, nil }
			} else {
				if !verifyWeakPassword(param) { return false, nil }
			}

			return true, Hash256(param)
		},
		func (req RequestInterface, param interface{}) (bool, interface{}) {

			s, ok := param.(string); if !ok { return false, nil }

			if hard {
				if !verifyHardPassword(s) { return false, nil }
			} else {
				if !verifyWeakPassword(s) { return false, nil }
			}

			return true, Hash256(s)
		},
	)
}

func verifyHardPassword(input string) bool {

    var letters int
    var number bool
    var special bool
    var upper bool
    
    for _, s := range input {

      switch {
        
        case unicode.IsNumber(s):
            number = true
        
        case unicode.IsUpper(s):
            upper = true
            letters++
        
        case unicode.IsPunct(s) || unicode.IsSymbol(s):
            special = true
        
        case unicode.IsLetter(s) || s == ' ':
            letters++

      }

    }

    return (letters >= 7) && number && special && upper
}

func verifyWeakPassword(input string) bool {

    var letters int
    var numbers int
    
    for _, s := range input {

      switch {
        
        case unicode.IsNumber(s):
            numbers++
        
        case unicode.IsLetter(s) || s == ' ':
            letters++

      }

    }

    return (letters + numbers) >= 8
}


// Returns a validation object that checks for an int which parses correctly
func Int() *ValidationConfig {

	return NewValidationConfig(
		0,
		func (req RequestInterface, param string) (bool, interface{}) {

			if len(param) == 0 { return false, nil }

			val, err := strconv.Atoi(param)

			return err == nil, val
		},
		func (req RequestInterface, param interface{}) (bool, interface{}) {

			i, ok := param.(float64)

			return ok, int(i)
		},
	)
}

// Returns a validation object that checks for an int which parses correctly and is positive
func PositiveInt() *ValidationConfig {

	return NewValidationConfig(
		0,
		func (req RequestInterface, param string) (bool, interface{}) {

			if len(param) == 0 { return false, nil }

			val, err := strconv.Atoi(param)

      if val <= 0 { return false, nil }

			return err == nil, val
		},
		func (req RequestInterface, param interface{}) (bool, interface{}) {

			i, ok := param.(float64)

      if i <= 0 { return false, nil }

			return ok, int(i)
		},
	)
}

// Returns a validation object that checks for an int which parses correctly and is negative
func NegativeInt() *ValidationConfig {

	return NewValidationConfig(
		0,
		func (req RequestInterface, param string) (bool, interface{}) {

			if len(param) == 0 { return false, nil }

			val, err := strconv.Atoi(param)

      if val >= 0 { return false, nil }

			return err == nil, val
		},
		func (req RequestInterface, param interface{}) (bool, interface{}) {

			i, ok := param.(float64)

      if i >= 0 { return false, nil }

			return ok, int(i)
		},
	)
}

// Returns a validation object that checks for an int which parses correctly and is zero or above
func OptimisticInt() *ValidationConfig {

	return NewValidationConfig(
		0,
		func (req RequestInterface, param string) (bool, interface{}) {

			if len(param) == 0 { return false, nil }

			val, err := strconv.Atoi(param)

      if val < 0 { return false, nil }

			return err == nil, val
		},
		func (req RequestInterface, param interface{}) (bool, interface{}) {

			i, ok := param.(float64)

      if i < 0 { return false, nil }

			return ok, int(i)
		},
	)
}

// Returns a validation object that checks for an int which parses correctly and is zero or lower
func PessimisticInt() *ValidationConfig {

	return NewValidationConfig(
		0,
		func (req RequestInterface, param string) (bool, interface{}) {

			if len(param) == 0 { return false, nil }

			val, err := strconv.Atoi(param)

      if val > 0 { return false, nil }

			return err == nil, val
		},
		func (req RequestInterface, param interface{}) (bool, interface{}) {

			i, ok := param.(float64)

      if i > 0 { return false, nil }

			return ok, int(i)
		},
	)
}

// Returns a validation object that checks for an int64 which parses correctly
func Int64() *ValidationConfig {

	return NewValidationConfig(
		int64(0),
		func (req RequestInterface, param string) (bool, interface{}) {

			if len(param) == 0 { return false, nil }

			val, err := strconv.ParseInt(param, 10, 64)

			return err == nil, val
		},
		func (req RequestInterface, param interface{}) (bool, interface{}) {

			i, ok := param.(float64)

			return ok, int64(i)
		},
	)
}

// Returns a validation object that checks for a float64 which parses correctly
func Float64(ranges ...int) *ValidationConfig {

	var min float64
	var max float64

	switch len(ranges) {

		case 2:

			min = float64(ranges[0])
			max = float64(ranges[1])

	}

	cfg := NewValidationConfig(
		float64(0),
		func (req RequestInterface, param string) (bool, interface{}) {

			if len(param) == 0 { return false, nil }

			val, err := strconv.ParseFloat(param, 64); if err != nil { return false, nil }

			if min + max == 0 { return true, val }

			return !(val > max || val < min), val
		},
		func (req RequestInterface, param interface{}) (bool, interface{}) {

			val, ok := param.(float64); if !ok { return false, 0 }

			if min + max == 0 { return true, val }

			return !(val > max || val < min), val
		},
		min,
		max,
	)


	return cfg
}

// Returns a validation object that checks for a bool which parses correctly
func Bool() *ValidationConfig {

	return NewValidationConfig(
		true,
		func (req RequestInterface, param string) (bool, interface{}) {

			if len(param) == 0 { return false, nil }

			switch param {

				case "true":	return true, true
				case "false":	return true, false

			}

			return false, false
		},
		func (req RequestInterface, param interface{}) (bool, interface{}) {

			b, ok := param.(bool); if !ok { return false, nil }

			return true, b
		},
	)
}

// Returns a validation object for request body that checks a property to see if it's an object
func MapStringInterface() *ValidationConfig {

	return NewValidationConfig(
		map[string]interface{}{},
		nil,
		func (req RequestInterface, param interface{}) (bool, interface{}) {

			x, ok := param.(map[string]interface{}); if !ok { return false, nil }

			return true, x
		},
	)
}

// Returns a validation object for request body that checks a property to see if it's an array
func InterfaceArray() *ValidationConfig {

	return NewValidationConfig(
		[]interface{}{},
		nil,
		func (req RequestInterface, param interface{}) (bool, interface{}) {

			m, ok := param.([]interface{}); if !ok { return false, nil }

			return true, m
		},
	)
}

// Returns a validation object for request body that checks a property to see if it's an array
func StringInterfaceArray() *ValidationConfig {

	return NewValidationConfig(
		[]string{},
		nil,
		func (req RequestInterface, param interface{}) (bool, interface{}) {

			a, ok := param.([]interface{}); if !ok { return false, nil }

			list := make([]string, len(a))

			for i, x := range a {

				list[i], ok = x.(string); if !ok { return false, nil }

			}

			return true, list
		},
	)
}

// Returns a validation object that checks to see if it can resolve to a country struct
func CountryISO2() *ValidationConfig {
		
	return NewValidationConfig(
		&Country{},
		func (req RequestInterface, param string) (bool, interface{}) {

			lp := len(param)
			if lp == 0 || lp > 64 { return false, nil }

			param = strings.ToUpper(param)

			country := globalNode.Config.countries[param]

			return (country != nil), country
		},
		func (req RequestInterface, param interface{}) (bool, interface{}) {

			s, ok := param.(string); if !ok { return false, nil }

			lp := len(s)
			if lp > 64 { return false, nil }

			country := globalNode.Config.countries[strings.ToUpper(s)]

			return (country != nil), country
		},
	)
}

// Returns a validation object that checks to see if it can resolve to a country struct
func LanguageISO2() *ValidationConfig {
		
	return NewValidationConfig(
		&Language{},
		func (req RequestInterface, param string) (bool, interface{}) {

			lp := len(param)
			if lp == 0 || lp > 64 { return false, nil }

			param = strings.ToUpper(param)

			lang := globalNode.Config.languages[param]

			return (lang != nil), lang
		},
		func (req RequestInterface, param interface{}) (bool, interface{}) {

			s, ok := param.(string); if !ok { return false, nil }

			lp := len(s)
			if lp > 64 { return false, nil }

			lang := globalNode.Config.languages[strings.ToUpper(s)]

			return (lang != nil), lang
		},
	)
}

// Returns a validation object that checks to see if a valid phone number is provided
func OptionalPhoneNumber(countryCode string) *ValidationConfig {
		
	return NewValidationConfig(
		"",
		func (req RequestInterface, number string) (bool, interface{}) {

			number = strings.TrimSpace(number)

			if len(number) == 0 { return true, "" }

			if len(number) > 30 { return false, nil }

			num, err := libphonenumber.Parse(number, countryCode); if err != nil { fmt.Println(err); return false, nil }

			format := libphonenumber.Format(num, libphonenumber.NATIONAL)

			return true, format
		},
		func (req RequestInterface, param interface{}) (bool, interface{}) {

			number, ok := param.(string); if !ok { return false, nil }

			number = strings.TrimSpace(number)

			if len(number) == 0 { return true, "" }

			if len(number) > 30 { return false, nil }

			num, err := libphonenumber.Parse(number, countryCode); if err != nil { fmt.Println(err); return false, nil }

			format := libphonenumber.Format(num, libphonenumber.NATIONAL)

			return true, format
		},
	)
}

// Returns a validation object that checks to see if a valid phone number is provided
func PhoneNumber(countryCode string) *ValidationConfig {
		
	return NewValidationConfig(
		"",
		func (req RequestInterface, number string) (bool, interface{}) {

			number = strings.TrimSpace(number)

			if len(number) > 30 { return false, nil }

			num, err := libphonenumber.Parse(number, countryCode); if err != nil { fmt.Println(err); return false, nil }

			format := libphonenumber.Format(num, libphonenumber.NATIONAL)

			return true, format
		},
		func (req RequestInterface, param interface{}) (bool, interface{}) {

			number, ok := param.(string); if !ok { return false, nil }

			number = strings.TrimSpace(number)

			if len(number) > 30 { return false, nil }

			num, err := libphonenumber.Parse(number, countryCode); if err != nil { fmt.Println(err); return false, nil }

			format := libphonenumber.Format(num, libphonenumber.NATIONAL)

			return true, format
		},
	)
}


