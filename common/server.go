package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"hfrpc/interceptor"
	"reflect"
	"strings"
	"sync"
)

type Method struct {
	Name       string
	ParamsType reflect.Type
	ResultType reflect.Type
	Method     reflect.Method
}

type Service struct {
	Name string
	V    reflect.Value
	T    reflect.Type
	Mm   map[string]*Method
}

type Server struct {
	Sm          sync.Map
	Interceptor []interceptor.HandlerFunc
}

func (svr *Server) Register(s interface{}) error {
	svc := new(Service)
	svc.V = reflect.ValueOf(s)
	svc.T = reflect.TypeOf(s)
	sname := reflect.Indirect(svc.V).Type().Name()
	svc.Name = sname
	svc.Mm = RegisterMethods(svc.T)
	if _, err := svr.Sm.LoadOrStore(sname, svc); err {
		return errors.New("rpc: service already defined: " + sname)
	}
	return nil
}

func RegisterMethods(s reflect.Type) map[string]*Method {
	mm := make(map[string]*Method)
	for m := 0; m < s.NumMethod(); m++ {
		rm := s.Method(m)
		if mt := RegisterMethod(rm); mt != nil {
			mm[rm.Name] = mt
		}
	}
	return mm
}

func RegisterMethod(rm reflect.Method) *Method {
	var (
		msg string
	)
	rmt := rm.Type
	rmn := rm.Name

	if !(rm.Type.NumIn() == 2 || rm.Type.NumIn() == 3) {
		msg = fmt.Sprintf("RegisterMethod: method %q has %d input parameters; needs exactly three", rmn, rmt.NumIn())
		Debug(msg)
		return nil
	}
	p := rmt.In(1)
	if p.Kind() != reflect.Ptr {
		msg = fmt.Sprintf("RegisterMethod: Params type of method %q is not a reflect.Ptr:%q", rmn, p)
		Debug(msg)
		return nil
	}

	if rm.Type.NumOut() != 2 {
		msg = fmt.Sprintf("RegisterMethod: Method %q has %d output parameters; needs exactly one", rmn, rmt.NumOut())
		Debug(msg)
		return nil
	}
	r := rmt.Out(0)
	if r.Kind() != reflect.Interface {
		msg = fmt.Sprintf("RegisterMethod: Return type of method %q is not a must be Interface:%q", rmn, r)
		Debug(msg)
		return nil
	}
	ret := rmt.Out(1)
	if ret != reflect.TypeOf((*error)(nil)).Elem() {
		msg = fmt.Sprintf("RegisterMethod: Return type of method %q is not a must be error:%q", rmn, ret)
		Debug(msg)
		return nil
	}
	m := &Method{rmn, p, r, rm}
	return m
}

func (svr *Server) Handler(b []byte) []byte {
	data, err := ParseRequestBody(b)
	if err != nil {
		return jsonE(nil, JsonRpc, ParseError)
	}
	var res interface{}
	if reflect.ValueOf(data).Kind() == reflect.Slice {
		var resList []interface{}
		for _, v := range data.([]interface{}) {
			r := svr.SingleHandler(v.(map[string]interface{}))
			resList = append(resList, r)
		}
		res = resList
	} else if reflect.ValueOf(data).Kind() == reflect.Map {
		r := svr.SingleHandler(data.(map[string]interface{}))
		res = r
	} else {
		return jsonE(nil, JsonRpc, InvalidRequest)
	}
	response, _ := json.Marshal(res)
	return response
}

func (svr *Server) SingleHandler(jsonMap map[string]interface{}) interface{} {
	id, jsonRpc, method, paramsData, errCode, content := ParseSingleRequestBody(jsonMap)
	if errCode != WithoutError {
		return E(id, jsonRpc, errCode, map[string]interface{}{})
	}
	sName, mName, err := ParseRequestMethod(method)
	if err != nil {
		return E(id, jsonRpc, MethodNotFound, map[string]interface{}{})
	}

	if len(svr.Interceptor) > 0 {
		for _, handleFun := range svr.Interceptor {
			err = handleFun(&content)
			if err != nil {
				return E(id, jsonRpc, InvalidRequest, map[string]interface{}{})
			}
		}
	}

	s, ok := svr.Sm.Load(sName)
	if !ok {
		sName = lineToHump(sName) // support HelloWorld and hello_world
		s, ok = svr.Sm.Load(sName)
		if !ok {
			return E(id, jsonRpc, MethodNotFound, map[string]interface{}{})
		}
	}
	m, ok := s.(*Service).Mm[mName]
	if !ok {
		return E(id, jsonRpc, MethodNotFound, map[string]interface{}{})
	}

	params := reflect.New(m.ParamsType.Elem())
	pv := params.Interface()
	err = GetStruct(paramsData, pv)
	if err != nil {
		return E(id, jsonRpc, InvalidParams, map[string]interface{}{})
	}
	var r []reflect.Value
	if m.Method.Type.NumIn() == 2 {
		r = m.Method.Func.Call([]reflect.Value{s.(*Service).V, params})
	} else if m.Method.Type.NumIn() == 3 {
		r = m.Method.Func.Call([]reflect.Value{s.(*Service).V, params, reflect.ValueOf(content)})
	}
	if len(r) <= 0 {
		return E(id, jsonRpc, InternalError, map[string]interface{}{})
	}
	if i := r[1].Interface(); i != nil {
		Debug(i.(error))
		return E(id, jsonRpc, InternalError, map[string]interface{}{})
	}
	return S(id, jsonRpc, r[0].Interface(), map[string]interface{}{})
}

func lineToHump(in string) string {
	s := strings.Split(in, "_")
	for k, v := range s {
		s[k] = Capitalize(v)
	}
	return strings.Join(s, "")
}

func Capitalize(str string) string {
	var upperStr string
	vv := []rune(str)
	for i := 0; i < len(vv); i++ {
		if i == 0 {
			if vv[i] >= 97 && vv[i] <= 122 {
				vv[i] -= 32
				upperStr += string(vv[i])
			} else {
				return str
			}
		} else {
			upperStr += string(vv[i])
		}
	}
	return upperStr
}
