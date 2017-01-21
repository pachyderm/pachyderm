package pkgtmpl

import (
	"io"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
)

type templater struct {
	baseDirPath   string
	funcMap       template.FuncMap
	templateCache map[string]*template.Template
	templateLock  *sync.RWMutex
	noCache       bool
}

func newTemplater(baseDirPath string) *templater {
	return &templater{
		baseDirPath,
		template.FuncMap{
			"lowercase": strings.ToLower,
			"uppercase": strings.ToUpper,
		},
		make(map[string]*template.Template),
		&sync.RWMutex{},
		false,
	}
}

func (h *templater) WithFuncs(funcMap template.FuncMap) Templater {
	newFuncMap := make(template.FuncMap)
	for key, value := range h.funcMap {
		newFuncMap[key] = value
	}
	for key, value := range funcMap {
		newFuncMap[key] = value
	}
	return &templater{
		h.baseDirPath,
		newFuncMap,
		make(map[string]*template.Template),
		&sync.RWMutex{},
		h.noCache,
	}
}

func (h *templater) NoCache() Templater {
	return &templater{
		h.baseDirPath,
		h.funcMap,
		make(map[string]*template.Template),
		&sync.RWMutex{},
		true,
	}
}

func (h *templater) Execute(writer io.Writer, name string, data interface{}) error {
	t, err := h.getTemplate(name)
	if err != nil {
		return err
	}
	return t.Execute(writer, data)
}

func (h *templater) getTemplate(name string) (*template.Template, error) {
	if !h.noCache {
		h.templateLock.RLock()
		if t, ok := h.templateCache[name]; ok {
			h.templateLock.RUnlock()
			return t, nil
		}
		h.templateLock.RUnlock()
	}
	t, err := template.New(name).Funcs(h.funcMap).ParseFiles(filepath.Join(h.baseDirPath, name))
	if err != nil {
		return nil, err
	}
	if !h.noCache {
		h.templateLock.Lock()
		h.templateCache[name] = t
		h.templateLock.Unlock()
	}
	return t, nil
}
