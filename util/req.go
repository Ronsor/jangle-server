package util

import (
	"encoding/json"

	"github.com/valyala/fasthttp"
)

func PostJSON(c *fasthttp.RequestCtx, i interface{}) error {
	return json.Unmarshal(c.PostBody(), i)
}

func WriteJSON(c *fasthttp.RequestCtx, i interface{}) error {
	b, e := json.Marshal(i)
	if e != nil { return e }
	_, e = c.Write(b)
	return e
}

func WriteJSONStatus(c *fasthttp.RequestCtx, n int, i interface{}) error {
	c.SetStatusCode(n)
	return WriteJSON(c, i)
}
