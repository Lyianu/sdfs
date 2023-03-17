package router

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Context struct {
	w   http.ResponseWriter
	req *http.Request

	status int
}

type H map[string]interface{}

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
	_, err := fmt.Fprintf(c.w, message, a)
	return err
}

func (c *Context) JSON(code int, h H) error {
	c.StatusCode(code)
	c.SetContentType("application/JSON")
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
