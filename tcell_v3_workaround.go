package main

import (
	"reflect"
	"unsafe"
	"github.com/gdamore/tcell/v3"
)

func Decompose(s tcell.Style) (fg, bg tcell.Color, attrs tcell.AttrMask) {
	rv := reflect.ValueOf(&s).Elem()

	readField := func(name string) uint64 {
		f := rv.FieldByName(name)
		f = reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
		return f.Uint()
	}

	fg = tcell.Color(readField("fg"))
	bg = tcell.Color(readField("bg"))
	attrs = tcell.AttrMask(readField("attrs"))

	return
}
