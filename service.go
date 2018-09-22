package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	// "github.com/gorilla/websocket"
)

type SocketRequest struct {
	messageType int
	req         []byte
	res         chan<- []byte
}

type ServeInterface struct {
	In interface{}
}

type SocketError struct {
	Error string `json:"error,omitempty"`
}

// func mapify(val reflect.Value) interface{} {
// 	if val.Kind() == reflect.Invalid {
// 		return nil
// 	} else if val.Kind() == reflect.Bool {
// 		return val.Bool()
// 	} else if val.Kind() >= reflect.Int && val.Kind() <= reflect.Int64 {
// 		return val.Int()
// 	} else if val.Kind() >= reflect.Uint && val.Kind() <= reflect.Uint64 {
// 		return val.Uint()
// 	} else if val.Kind() >= reflect.Float32 && val.Kind() <= reflect.Float64 {
// 		return val.Float()
// 	} else if val.Kind() >= reflect.Complex64 && val.Kind() <= reflect.Complex128 {
// 		return val.Complex()
// 	} else if val.Kind() == reflect.Array || val.Kind() == reflect.Slice {
// 		ret := make([]interface{}, val.Len())
// 		for i := 0; i < val.Len(); i++ {
// 			ret[i] = mapify(val.Index(i))
// 		}
// 		return ret
// 	} else if val.Kind() == reflect.Map {
// 		ret := make(map[string]interface{})
// 		keys := val.MapKeys()
// 		for _, key := range keys {
// 			k := fmt.Sprintf("%v", key)
// 			ret[k] = mapify(val.MapIndex(key))
// 		}
// 		return ret
// 	} else if val.Kind() == reflect.String {
// 		return val.String()
// 	} else if val.Kind() == reflect.Func && val.Type().NumIn() == 0 {
// 		return mapify(val.Call([]reflect.Value{}))
// 	} else if val.Kind() == reflect.Ptr {
// 		return mapify(val.Elem())
// 	} else if val.Kind() == reflect.Struct || val.Kind() == reflect.Interface {
// 		ret := make(map[string]interface{})
// 		t := val.Type()
// 		if val.Kind() == reflect.Struct {
// 			for i := 0; i < val.NumField(); i++ {
// 				ret[t.Field(i).Name] = mapify(val.Field(i))
// 			}
// 		}
// 		for i := 0; i < val.NumMethod(); i++ {
// 			if m := val.Method(i); m.Type().NumIn() == 0 {
// 				ret[m.Type().Name()] = mapify(m.Call([]reflect.Value{}))
// 			}
// 		}
// 		return ret
// 	}
//
// 	return fmt.Sprintf("'%s' not supported", val.Kind().String())
// }

func (s *ServeInterface) Serve(path string, value interface{}) (ret interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	return servePath(reflect.ValueOf(s.In), strings.Split(path, "/"))
}

func servePath(val reflect.Value, path []string) (ret interface{}, err error) {
	if len(path) == 0 || path[0] == "" {
		if val.CanInterface() {
			ret = val.Interface()
		}
		return
	}

	p, rest := path[0], path[1:]

	if m := val.MethodByName(p); m != (reflect.Value{}) {
		if m.Type().NumIn() == 0 {
			os := m.Call([]reflect.Value{})
			if len(os) == 0 {
				os = append(os, reflect.Value{})
			}
			return servePath(os[0], rest)
		}
		err = fmt.Errorf("method '%s' accepts more than 0 parameters", p)
		return
	}

	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			err = fmt.Errorf("value is nil")
			return
		}
		return servePath(val.Elem(), path)
	}

	if val.Kind() != reflect.Struct {
		err = fmt.Errorf("path not found: %s because %s is %s", p, val.Type().Name(), val.Kind().String())
		return
	}

	if f := val.FieldByName(p); f == (reflect.Value{}) {
		err = fmt.Errorf("path not found: %s in %s", p, val.Type().Name())
		return
	} else {
		return servePath(f, rest)
	}

}

func (s *ServeInterface) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			http.Error(w, fmt.Sprintf("%v", r), 500)
		}
	}()

	log.Printf("Interface Server: %s", r.URL.Path)

	var val interface{}

	if r.Method == http.MethodPost {
		if b, err := ioutil.ReadAll(r.Body); err != nil {
			http.Error(w, err.Error(), 500)
		} else if exp, err := parser.ParseExpr(string(b)); err != nil {
			http.Error(w, err.Error(), 400)
		} else if lit, ok := exp.(*ast.BasicLit); !ok {
			http.Error(w, "body must be a basic literal", 400)
		} else {
			switch lit.Kind {
			case token.INT:
				val, _ = strconv.Atoi(lit.Value)
			case token.FLOAT:
				val, _ = strconv.ParseFloat(lit.Value, 64)
			case token.STRING:
				val = lit.Value
			default:
				http.Error(w, "body must be an integer, float or string", 400)
				return
			}
		}
	}

	if out, err := s.Serve(r.URL.Path, val); err != nil {
		http.Error(w, err.Error(), 500)
	} else {
		wrapped := make(map[string]interface{})
		wrapped[r.URL.Path] = out
		if b, err := json.Marshal(wrapped); err != nil {
			http.Error(w, err.Error(), 500)
		} else if _, err := w.Write(b); err != nil {
			log.Println(err)
		}
	}
}
