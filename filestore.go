package main

import (
	"os"
	"io"
	"bytes"
	"path/filepath"

	"github.com/valyala/fasthttp"
	"github.com/fasthttp/router"
)

var gFileStore = FileStore(&BogusCDN{})

type FileStore interface {
	Init(...interface{}) error
	PerformUpload(relpath string, pipe io.Reader) (string, error)
}

func FileUpload(f FileStore, path string, pipe io.Reader) (string, error) {
	out, err := f.PerformUpload(path, pipe)
	return *flgCDNServeBase + out, err
}

func BytesUpload(f FileStore, path string, data []byte) (string, error) {
	return FileUpload(f, path, bytes.NewReader(data))
}

type BogusCDN struct {}

func (b *BogusCDN) Init(opts ...interface{}) error {
	os.MkdirAll(*flgFileServerPath, 0755)
	opts[0].(*router.Router).GET("/boguscdn/*filepath", (&fasthttp.FS{
		Root: *flgFileServerPath,
		GenerateIndexPages: true,
		PathRewrite: fasthttp.NewPathSlashesStripper(1),
	}).NewRequestHandler())
	return nil
}


func (b *BogusCDN) PerformUpload(path string, pipe io.Reader) (string, error) {
	os.MkdirAll(*flgFileServerPath + "/" + filepath.Dir(path), 0755)
	file, err := os.Create(*flgFileServerPath + "/" + path)
	if err != nil { return "", err }
	_, err = io.Copy(file, pipe)
	if err != nil { return "", err }
	file.Close()
	return path, nil
}
