// +build tools

package main

import (
	"fmt"
	"log"
	"os"
	"text/template"
)

func main() {
	data := map[string]string{
		"Object":   "*Object",
		"Storager": "Storager",
		"Part":     "*Part",
		"Block":    "*Block",
	}

	generateT(tmpl, "types/iterator.generated.go", data)
}

func generateT(tmpl *template.Template, filePath string, data interface{}) {
	errorMsg := fmt.Sprintf("generate template %s to %s", tmpl.Name(), filePath) + ": %v"

	file, err := os.Create(filePath)
	if err != nil {
		log.Fatalf(errorMsg, err)
	}
	err = tmpl.Execute(file, data)
	if err != nil {
		log.Fatalf(errorMsg, err)
	}
}

var tmpl = template.Must(template.New("iterator").Parse(`
// Code generated by go generate internal/cmd; DO NOT EDIT.
package types

import (
	"context"
	"errors"
	"fmt"
)

type Continuable interface {
	ContinuationToken() string
}

{{- range $k, $v := . }}
/*
NextObjectFunc is the func used in iterator.

Notes
- ErrDone should be return while there are no items any more.
- Input objects slice should be set every time.
*/
type Next{{$k}}Func func(ctx context.Context, page *{{$k}}Page) error

type {{$k}}Page struct {
	Status Continuable
	Data  []{{$v}}
}

type {{$k}}Iterator struct {
	ctx context.Context
	next Next{{$k}}Func

	index int
	done  bool

	o {{$k}}Page
}

func New{{$k}}Iterator(ctx context.Context, next Next{{$k}}Func, status Continuable) *{{$k}}Iterator {
	return &{{$k}}Iterator{
		ctx: ctx,
		next:  next,
		index: 0,
		done:  false,
		o:     {{$k}}Page{
			Status: status,
		},
	}
}

func (it *{{$k}}Iterator) ContinuationToken() string {
	return it.o.Status.ContinuationToken()
}

func (it *{{$k}}Iterator) Next() (object {{$v}}, err error) {
	// Consume Data via index.
	if it.index < len(it.o.Data) {
		it.index++
		return it.o.Data[it.index-1], nil
	}
	// Return IterateDone if iterator is already done.
	if it.done {
		return nil, IterateDone
	}

	// Reset buf before call next.
	it.o.Data = it.o.Data[:0]

	err = it.next(it.ctx ,&it.o)
	if err != nil && !errors.Is(err, IterateDone) {
		return nil, fmt.Errorf("iterator next failed: %w", err)
	}
	// Make iterator to done so that we will not fetch from upstream anymore.
	if err != nil {
		it.done = true
	}
	// Return IterateDone directly if we don't have any data.
	if len(it.o.Data) == 0 {
		return nil, IterateDone
	}
	// Return the first object.
	it.index = 1
	return it.o.Data[0], nil
}
{{- end }}
`))
