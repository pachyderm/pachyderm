package pachtmpl

import (
	"io"
	"net/http"
	"strings"

	"github.com/google/go-jsonnet"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

// Eval evaluates the jsonnet at entrypointPath using fsContext to resolve imports
func Eval(fsContext map[string][]byte, entrypointPath string, args []string) ([]byte, error) {
	vm := jsonnet.MakeVM()
	// setup importer for fs
	memImp := jsonnet.MemoryImporter{
		Data: make(map[string]jsonnet.Contents),
	}
	for k, v := range fsContext {
		memImp.Data[k] = jsonnet.MakeContents(string(v))
	}
	httpImp := newHTTPImporter()
	imp := importerMux{
		{Prefix: "http://", Importer: httpImp},
		{Prefix: "https://", Importer: httpImp},
		{Importer: &memImp},
	}
	vm.Importer(&imp)
	// parse arguments
	for _, argStr := range args {
		kv := strings.SplitN(argStr, "=", 2)
		if len(kv) != 2 {
			return nil, errors.Errorf("invalid template argument %q: must have form \"key=value\"", argStr)
		}
		vm.TLAVar(kv[0], kv[1])
	}
	output, err := vm.EvaluateFile(entrypointPath)
	if err != nil {
		return nil, errors.Wrapf(err, "template err")
	}
	return []byte(output), nil
}

type httpImporter struct {
	hc *http.Client
}

func newHTTPImporter() *httpImporter {
	return &httpImporter{
		hc: http.DefaultClient,
	}
}

func (im httpImporter) Import(importedFrom, importedPath string) (contents jsonnet.Contents, from string, err error) {
	resp, err := im.hc.Get(importedPath)
	if err != nil {
		return jsonnet.Contents{}, "", err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return jsonnet.Contents{}, "", err
	}
	return jsonnet.MakeContents(string(data)), importedPath, err
}

type muxRule struct {
	Prefix   string
	Importer jsonnet.Importer
}

type importerMux []muxRule

func (im importerMux) Import(importedFrom, importedPath string) (contents jsonnet.Contents, from string, err error) {
	for i := 0; i < len(im)-1; i++ {
		rule := im[i]
		if strings.HasPrefix(importedPath, rule.Prefix) {
			return rule.Importer.Import(importedFrom, importedPath)
		}
	}
	return im[len(im)-1].Importer.Import(importedFrom, importedPath)
}
