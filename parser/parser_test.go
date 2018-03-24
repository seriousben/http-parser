package parser

import (
	"strings"
	"testing"

	"github.com/gotestyourself/gotestyourself/assert"
	"github.com/gotestyourself/gotestyourself/assert/cmp"
	"github.com/gotestyourself/gotestyourself/golden"
	"github.com/opsidian/parsley/parsley"
	"github.com/opsidian/parsley/text"
)

func TestStatusCode(t *testing.T) {
	parsedData := new(ResponseData)

	parser := statusLineParser()

	s := parsley.NewSentence(parser)
	_, _, err := s.Evaluate(text.NewReader([]byte("HTTP/1.1 200 OK"), "", true), parsedData)
	assert.NilError(t, err)

	// fmt.Printf("%+v", parsedData)
	assert.Equal(t, parsedData.StatusCode, 200)
	assert.Equal(t, parsedData.StatusString, "OK")
	assert.Equal(t, parsedData.ProtocolVersion, "HTTP/1.1")
}

func TestHeader(t *testing.T) {
	parser := headerParser()

	s := parsley.NewSentence(parser)
	val, _, err := s.Evaluate(text.NewReader([]byte("Content-Type: application/json"), "", true), nil)
	assert.NilError(t, err)

	valArray := val.([]string)

	//fmt.Printf("%+v", valArray)
	assert.Assert(t, cmp.Len(valArray, 2))
	assert.Equal(t, valArray[0], "Content-Type")
	assert.Equal(t, valArray[1], "application/json")
}

func TestHeaders(t *testing.T) {
	parsedData := NewResponseData()

	parser := headersParser()

	s := parsley.NewSentence(parser)

	headers := []string{
		"Content-Type: application/json",
		"Date: Mon, 23 May 2005 22:38:34 GMT",
	}
	_, _, err := s.Evaluate(text.NewReader([]byte(strings.Join(headers, "\n")+"\n"), "", true), parsedData)
	assert.NilError(t, err)

	//fmt.Printf("%+v", parsedData)
	assert.Assert(t, cmp.Len(parsedData.Headers, 2))
	assert.Equal(t, parsedData.Headers["Content-Type"], "application/json")
}

func TestParseLines(t *testing.T) {
	b := golden.Get(t, "simple-http.text")

	parsedData := NewResponseData()

	lines := linesParser()
	s := parsley.NewSentence(lines)
	_, _, err := s.Evaluate(text.NewReader(b, "", true), parsedData)
	assert.NilError(t, err)
	assert.Check(t, golden.String(parsedData.ToHTTP()+"\n", "simple-http.text"))
}

func TestParse(t *testing.T) {
	b := golden.Get(t, "simple-http.text")
	parsedData, err := Parse(b)

	assert.NilError(t, err)
	//fmt.Printf("%+v", parsedData)
	assert.Equal(t, parsedData.StatusCode, 200)
	assert.Equal(t, parsedData.StatusString, "OK")
	assert.Equal(t, parsedData.ProtocolVersion, "HTTP/1.1")
	assert.Equal(t, parsedData.Headers["Content-Type"], "text/html; charset=UTF-8")
	assert.Equal(t, parsedData.Headers["Connection"], "close")
	assert.Check(t, golden.String(parsedData.ToHTTP()+"\n", "simple-http.text"))
}
