package util

import (
	"encoding/json"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/go-playground/validator.v9"
	"github.com/valyala/fasthttp"
)

func ReadPostAny(c *fasthttp.RequestCtx, i interface{}, opts ...interface{}) error {
	frm, err := c.Request.MultipartForm()
	if err != nil { return ReadPostJSON(c, i, opts...) }
	var m map[string]interface{}
	for k, v := range frm.Value {
		m[k] = v[0]
	}
	for k, v := range frm.File {
		m[k] = v[0]
	}
	if frm.Value["payload_json"] != nil {
		err := json.Unmarshal([]byte(frm.Value["payload_json"][0]), i)
		if err != nil { return err }
	}
	dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{TagName: "json", Result: i, WeaklyTypedInput: true})
	if err != nil {
		return err
	}
	err = dec.Decode(m)
	if err != nil { return err }
	if len(opts) == 0 || !opts[0].(bool) {
		v := validator.New()
		err = v.Struct(i)
		if err != nil {
			return err
		}
	}
	return nil
}

func ReadPostJSON(c *fasthttp.RequestCtx, i interface{}, opts ...interface{}) error {
	err := json.Unmarshal(c.PostBody(), i)
	if err != nil {
		return err
	}
	if len(opts) == 0 || !opts[0].(bool) {
		v := validator.New()
		err = v.Struct(i)
		if err != nil {
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
