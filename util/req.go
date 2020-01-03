package util

import (
	"encoding/json"

	"gopkg.in/go-playground/validator.v9"
	"github.com/valyala/fasthttp"
)

func ReadPostJSON(c *fasthttp.RequestCtx, i interface{}, opts ...interface{}) error {
	err := json.Unmarshal(c.PostBody(), i)
	if err != nil {
		return err
	}
	if len(opts) == 0 || !opts[0].(bool) {
		v := validator.New()
		err = v.Struct(i)
		if err != nil {
			panic(err)
			return err
		}
	}
	return nil
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
