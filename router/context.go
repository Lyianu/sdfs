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

func (c *Context) StatusCode(code int) int {
	c.status = code
	return code
}

func (c *Context) SetContentType(cType string) string {
	c.w.Header().Set("Content-Type", cType)
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
