/*
Package parser defines method for parsing an HTTP response based on https://tools.ietf.org/html/rfc2068.
*/
package parser

import (
	"fmt"
	"strings"

	"github.com/opsidian/parsley/ast"
	"github.com/opsidian/parsley/ast/builder"
	"github.com/opsidian/parsley/combinator"
	"github.com/opsidian/parsley/parser"
	"github.com/opsidian/parsley/parsley"
	"github.com/opsidian/parsley/reader"
	"github.com/opsidian/parsley/text"
	"github.com/opsidian/parsley/text/terminal"
)

// ResponseData is a representation of a https://tools.ietf.org/html/rfc2068 compliant payload.
type ResponseData struct {
	ProtocolVersion   string
	StatusCode        int
	StatusString      string
	OrderedHeaderKeys []string
	Headers           map[string]string
	Unparsed          []string
}

// NewResponseData creates a new ResponsaData with default values.
func NewResponseData() *ResponseData {
	return &ResponseData{
		Headers: make(map[string]string),
	}
}

// ToHTTP transforms data into an https://tools.ietf.org/html/rfc2068 compliant payload.
func (r *ResponseData) ToHTTP() string {
	payload := []string{}
	if r.ProtocolVersion != "" {
		// HTTP/1.1 200 OK
		statusLine := fmt.Sprintf("%s %d %s", r.ProtocolVersion, r.StatusCode, r.StatusString)
		payload = append(payload, statusLine)
	}
	for _, key := range r.OrderedHeaderKeys {
		headerLine := fmt.Sprintf("%s: %s", key, r.Headers[key])
		payload = append(payload, headerLine)
	}
	payload = append(payload, r.Unparsed...)
	return strings.Join(payload, "\n")
}

func statusLineParser() parser.Func {
	return combinator.Seq(
		builder.All(
			"STATUS_LINE",
			ast.InterpreterFunc(func(ctx interface{}, nodes []ast.Node) (interface{}, reader.Error) {
				responseData := ctx.(*ResponseData)
				semver, _ := nodes[1].Value(ctx)
				responseData.ProtocolVersion = fmt.Sprintf("HTTP/%s", semver)

				statusCodeRaw, _ := nodes[2].Value(ctx)
				responseData.StatusCode = statusCodeRaw.(int)

				statusStringRaw, _ := nodes[3].Value(ctx)
				responseData.StatusString = statusStringRaw.(string)
				return nil, nil
			}),
		),
		terminal.Word("HTTP", "HTTP", "HTTP"),
		terminal.Regexp("protocol version", "/([^ ]+)", false, 1, "PROTOCOL VERSION"),
		terminal.Integer(),
		terminal.Regexp("any char", ".*", false, 0, "CONTENT"),
	)
}

func headerParser() parser.Func {
	return combinator.Seq(
		builder.All(
			"HEADER",
			ast.InterpreterFunc(func(ctx interface{}, nodes []ast.Node) (interface{}, reader.Error) {
				value0, _ := nodes[0].Value(ctx)
				value2, _ := nodes[2].Value(ctx)

				return []string{value0.(string), value2.(string)}, nil
			}),
		),
		terminal.Regexp("header name", "[^:]*", false, 0, "HEADER_NAME"),
		terminal.Rune(':', "COLON"),
		terminal.Regexp("any char", ".*", false, 0, "HEADER_VALUE"),
	)
}

func headersParser() parser.Func {
	newline := terminal.Regexp("newline", "\n", true, 0, "NEWLINE")
	line := combinator.Seq(
		builder.All(
			"LINE",
			ast.InterpreterFunc(func(ctx interface{}, nodes []ast.Node) (interface{}, reader.Error) {
				value0, _ := nodes[0].Value(ctx)
				return value0, nil
			}),
		),
		headerParser(),
		newline,
	)

	return combinator.Many(
		builder.All(
			"HEADERS",
			ast.InterpreterFunc(func(ctx interface{}, nodes []ast.Node) (interface{}, reader.Error) {
				responseData := ctx.(*ResponseData)
				for _, node := range nodes {
					val, _ := node.Value(ctx)
					headerVals, _ := val.([]string)
					responseData.OrderedHeaderKeys = append(responseData.OrderedHeaderKeys, headerVals[0])
					responseData.Headers[headerVals[0]] = headerVals[1]
				}
				return nil, nil
			}),
		),
		line,
	)
}

func linesParser() parser.Func {
	newline := terminal.Regexp("newline", "\n", true, 0, "NEWLINE")
	line := combinator.Seq(
		builder.All(
			"LINE",
			ast.InterpreterFunc(func(ctx interface{}, nodes []ast.Node) (interface{}, reader.Error) {
				value0, _ := nodes[0].Value(ctx)

				return value0.(string), nil
			}),
		),
		terminal.Regexp("any char", "[^\n]*", true, 0, "CONTENT"),
		newline,
	)

	return combinator.Many(
		builder.All(
			"LINES",
			ast.InterpreterFunc(func(ctx interface{}, nodes []ast.Node) (interface{}, reader.Error) {
				responseData := ctx.(*ResponseData)
				var res string
				for _, node := range nodes {
					val, _ := node.Value(ctx)
					responseData.Unparsed = append(responseData.Unparsed, val.(string))
				}
				return res, nil
			}),
		),
		line,
	)
}

func httpResponseParser() parser.Func {
	return combinator.Seq(
		builder.All(
			"LINES",
			ast.InterpreterFunc(func(ctx interface{}, nodes []ast.Node) (interface{}, reader.Error) {
				for _, node := range nodes {
					_, _ = node.Value(ctx)
				}
				return nil, nil
			}),
		),
		statusLineParser(),
		headersParser(),
		linesParser(),
	)
}

// Parse bytes into an HTTP Response
func Parse(b []byte) (*ResponseData, error) {
	parsedData := NewResponseData()

	s := parsley.NewSentence(httpResponseParser())
	_, _, err := s.Evaluate(text.NewReader(b, "", true), parsedData)
	if err != nil {
		return parsedData, err
	}

	return parsedData, nil
}
