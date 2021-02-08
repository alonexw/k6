/*
 *
 * k6 - a next-generation load testing tool
 * Copyright (C) 2020 Load Impact
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package data

import (
	"github.com/dop251/goja"
)

// TODO fix it not working really well with setupData or just make it more broken
// TODO fix it working with console.log
type sharedArray struct {
	arr []string
}

type wrappedSharedArray struct {
	sharedArray

	rt       *goja.Runtime
	freeze   goja.Callable
	isFrozen goja.Callable
	parse    goja.Callable
}

func (s sharedArray) wrap(rt *goja.Runtime) goja.Value {
	freeze, _ := goja.AssertFunction(rt.GlobalObject().Get("Object").ToObject(rt).Get("freeze"))
	isFrozen, _ := goja.AssertFunction(rt.GlobalObject().Get("Object").ToObject(rt).Get("isFrozen"))
	parse, _ := goja.AssertFunction(rt.GlobalObject().Get("JSON").ToObject(rt).Get("parse"))
	return rt.NewDynamicArray(wrappedSharedArray{
		sharedArray: s,
		rt:          rt,
		freeze:      freeze,
		isFrozen:    isFrozen,
		parse:       parse,
	})
}

func (s wrappedSharedArray) Set(index int, val goja.Value) bool {
	panic(s.rt.NewTypeError("SharedArray is immutable")) // this is specifically a type error
}

func (s wrappedSharedArray) SetLen(len int) bool {
	panic(s.rt.NewTypeError("SharedArray is immutable")) // this is specifically a type error
}

func (s wrappedSharedArray) Get(index int) goja.Value {
	if index < 0 || index >= len(s.arr) {
		return goja.Undefined()
	}
	val, err := s.parse(goja.Undefined(), s.rt.ToValue(s.arr[index]))
	if err != nil {
		panic(err)
	}
	s.deepFreeze(s.rt, val)

	return val
}

func (s wrappedSharedArray) Len() int {
	return len(s.arr)
}

func (s wrappedSharedArray) deepFreeze(rt *goja.Runtime, val goja.Value) {
	_, err := s.freeze(goja.Undefined(), val)
	if err != nil {
		panic(s.rt.NewTypeError(err))
	}

	if goja.IsUndefined(val) {
		return
	}

	o := val.ToObject(rt)
	for _, key := range o.Keys() {
		prop := o.Get(key)
		if prop != nil {
			frozen, err := s.isFrozen(goja.Undefined(), prop)
			if err != nil {
				panic(s.rt.NewTypeError(err))
			}
			if !frozen.ToBoolean() {
				s.deepFreeze(rt, prop)
			}
		}
	}
}
