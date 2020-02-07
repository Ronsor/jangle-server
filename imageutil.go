package main

import (
	"fmt"
	"path"
	"bytes"
	"strings"

	"image"
	"image/png"
	"image/jpeg"
	"image/gif"
)

type ImageUploadOptions struct {
	ForcePNG, ForceJPEG bool
	AllowGIF bool
	MaxWidth, MaxHeight int
}

func ImageBytesUpload(f FileStore, upath string, data []byte, opts ImageUploadOptions) (realPath string, err error) {
	cfg, typ, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil { return "", err }
	if opts.MaxWidth != 0 && opts.MaxHeight != 0 && (cfg.Width > opts.MaxWidth || cfg.Height > opts.MaxHeight) {
		return "", fmt.Errorf("Image too big.")
	}

	outTyp := typ
	if opts.ForcePNG || (typ == "gif" && !opts.AllowGIF) { outTyp = "png" }

	upath, name := path.Split(upath)
	if path.Ext(name) != "" { name = strings.TrimRight(name, path.Ext(name)) }
	name = name + "." + outTyp

	var img interface{}
	switch typ + "->" + outTyp {
		case "gif->gif":
			img, err = gif.DecodeAll(bytes.NewReader(data))
		default:
			img, _, err = image.Decode(bytes.NewReader(data))
	}
	if err != nil { return "", err }

	// TODO: animated GIF
	outBuf := bytes.Buffer{}
	switch outTyp {
		case "png":
			err = png.Encode(&outBuf, img.(image.Image))
		case "jpeg":
			err = jpeg.Encode(&outBuf, img.(image.Image), &jpeg.Options{Quality: 80})
		case "gif":
			err = gif.EncodeAll(&outBuf, img.(*gif.GIF))
		default:
			panic("Impossible state")
	}
	if err != nil { return "", err }
	return BytesUpload(f, upath, outBuf.Bytes())
}
