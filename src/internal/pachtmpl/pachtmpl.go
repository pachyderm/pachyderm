package pachtmpl

import (
	"io"
	"net/http"
	"strings"

	"github.com/google/go-jsonnet"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

// ParseArgs parses args of the form key=value
func ParseArgs(argStrs []string) (map[string]string, error) {
	ret := make(map[string]string)
	for _, argStr := range argStrs {
		kv := strings.SplitN(argStr, "=", 2)
		if len(kv) != 2 {
			return nil, errors.Errorf("invalid template argument %q: must have form \"key=value\"", argStr)
		}
		key, value := kv[0], kv[1]
		ret[key] = value
	}
	return ret, nil
}

// RenderTemplate renders the template tmpl, using args and returns the result.
func RenderTemplate(tmpl string, args map[string]string) (string, error) {
	vm := newVM(nil)
	for key, value := range args {
		vm.TLAVar(key, value)
	}
	output, err := vm.EvaluateAnonymousSnippet("main", string(tmpl))
	if err != nil {
		return "", errors.Wrapf(err, "template err")
	}
	return output, nil
}

// Eval evaluates the jsonnet at entrypointPath using fsContext to resolve imports
func Eval(fsContext map[string][]byte, entrypointPath string) ([]byte, error) {
	vm := newVM(fsContext)
	output, err := vm.EvaluateFile(entrypointPath)
	if err != nil {
		return nil, errors.Wrapf(err, "template err")
	}
	return []byte(output), nil
}

func newVM(fsContext map[string][]byte) *jsonnet.VM {
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
	return vm
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
		return jsonnet.Contents{}, "", errors.EnsureStack(err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return jsonnet.Contents{}, "", errors.Errorf("http code %d fetching %s", resp.StatusCode, importedPath)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return jsonnet.Contents{}, "", errors.EnsureStack(err)
	}
	return jsonnet.MakeContents(string(data)), importedPath, nil
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
			contents, from, err = rule.Importer.Import(importedFrom, importedPath)
			err = errors.EnsureStack(err)
			return contents, from, err
		}
	}
	contents, from, err = im[len(im)-1].Importer.Import(importedFrom, importedPath)
	err = errors.EnsureStack(err)
	return contents, from, err
}
