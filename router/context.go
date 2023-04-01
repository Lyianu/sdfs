package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type Context struct {
	w   http.ResponseWriter
	req *http.Request

	status int
}

type H map[string]interface{}

type Range struct {
	Start int64
	End   int64
}

func NewContext(w http.ResponseWriter, req *http.Request) *Context {
	c := &Context{
		w:   w,
		req: req,
	}
	return c
}

func (c *Context) StatusCode(code int) int {
	c.status = code
	return code
}

func (c *Context) SetHeader(key, value string) string {
	c.w.Header().Set(key, value)
	return key
}

func (c *Context) SetContentType(cType string) string {
	c.SetHeader("Content-Type", cType)
	return cType
}

func (c *Context) String(code int, message string, a ...interface{}) error {
	c.StatusCode(code)
	c.SetContentType("text/plain")
	c.w.WriteHeader(c.status)
	_, err := fmt.Fprintf(c.w, message, a...)
	return err
}

func (c *Context) JSON(code int, h H) error {
	c.StatusCode(code)
	c.SetContentType("application/JSON")
	c.w.WriteHeader(c.status)
	encoder := json.NewEncoder(c.w)
	err := encoder.Encode(h)
	if err != nil {
		c.StatusCode(http.StatusInternalServerError)
	}
	return err
}

func (c *Context) Query(queryString string) string {
	return c.req.URL.Query().Get(queryString)
}

func (c *Context) Header(headerString string) string {
	return c.req.Header.Get(headerString)
}

func (c *Context) ParseRange(size int64) ([]Range, error) {
	rangeHeader := c.Header("Range")
	if rangeHeader == "" {
		return nil, nil
	}

	if rangeHeader[:6] != "bytes=" {
		return nil, fmt.Errorf("invalid range header: %s", rangeHeader)
	}

	rangeHeader = rangeHeader[6:]
	ranges := strings.Split(rangeHeader, ",")

	var result []Range
	for _, r := range ranges {
		var start, end int64
		var err error

		parts := strings.Split(r, "-")
		if len(parts) == 1 {
			start, err = strconv.ParseInt(parts[0], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid range header: %s", rangeHeader)
			}
			end = size - 1
		} else if len(parts) == 2 {
			if parts[0] == "" {
				end, err = strconv.ParseInt(parts[1], 10, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid range header: %s", rangeHeader)
				}
				start = size - end
				end = size - 1
			} else if parts[1] == "" {
				start, err = strconv.ParseInt(parts[0], 10, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid range header: %s", rangeHeader)
				}
				end = size - 1
			} else {
				start, err = strconv.ParseInt(parts[0], 10, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid range header: %s", rangeHeader)
				}
				end, err = strconv.ParseInt(parts[1], 10, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid range header: %s", rangeHeader)
				}
				if end < start {
					return nil, fmt.Errorf("invalid range header: %s", rangeHeader)
				}
			}
		} else {
			return nil, fmt.Errorf("invalid range header: %s", rangeHeader)
		}

		if start >= size || end >= size {
			return nil, fmt.Errorf("invalid range header: %s", rangeHeader)
		}

		result = append(result, Range{Start: start, End: end + 1})
	}

	return result, nil
}
