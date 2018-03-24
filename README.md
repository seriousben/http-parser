# http-parser

HTTP response parser using [parser combinators](https://en.wikipedia.org/wiki/Parser_combinator).

## Parses [RFC2068](https://tools.ietf.org/html/rfc2068) compliant payload

```
HTTP/1.1 200 OK
Date: Mon, 23 May 2005 22:38:34 GMT
Content-Type: text/html; charset=UTF-8
Content-Encoding: UTF-8
Content-Length: 138
Last-Modified: Wed, 08 Jan 2003 23:11:55 GMT
Server: Apache/1.3.3.7 (Unix) (Red-Hat/Linux)
ETag: "3f80f-1b6-3e1cb03b"
Accept-Ranges: bytes
Connection: close
<html>
<head>
  <title>An Example Page</title>
</head>
<body>
  Hello World, this is a very simple HTML document.
</body>
</html>
```

## Example

```go
b, _ := ioutil.ReadFile("./parser/testdata/simple-http.text")
parsedData, _ := Parse(b)
fmt.Printf("Status Code: %d", parsedData.StatusCode)
```
