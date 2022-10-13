package env

import (
	"fmt"
	"gitee.com/hasika/gots"
	"gitee.com/hasika/v8go"
	"reflect"
)

type ServerEnv struct {
	*gots.TsEnv
	goCallTsEventClassValue *v8go.Object
	goCallTsEventFunc       map[string]*v8go.Value
}

func NewServerEnv(dir string) *ServerEnv {
	tsEnv := gots.NewTsEnv(dir)
	return &ServerEnv{
		TsEnv:             tsEnv,
		goCallTsEventFunc: map[string]*v8go.Value{},
	}
}
func (t *ServerEnv) GetGoEventClass() *v8go.Object {
	if t.goCallTsEventClassValue != nil {
		return t.goCallTsEventClassValue
	}
	GoCallTsEventClassValue, err := t.Ctx.Global().Get("GoCallTsEvent")
	if err != nil {
		panic(err)
	}
	GoCallTsEventClass, err := GoCallTsEventClassValue.AsObject()
	if err != nil {
		panic(err)
	}
	t.goCallTsEventClassValue = GoCallTsEventClass
	return GoCallTsEventClass
}

func (t *ServerEnv) GetFunc(name string) *v8go.Value {
	f, ex := t.goCallTsEventFunc[name]
	if ex {
		return f
	}
	GoCallTsEventClass := t.GetGoEventClass()
	f, err := GoCallTsEventClass.Get(name)
	if err != nil {
		panic(err)
	}
	t.goCallTsEventFunc[name] = f
	return f
}

func (t *ServerEnv) GoCallTsEvent(name string, args ...interface{}) (ret *v8go.Value) {
	functionValue := t.GetFunc(name)
	function, err := functionValue.AsFunction()
	if err != nil {
		panic(err)
	}
	jsArgs := make([]v8go.Valuer, 0)
	for _, arg := range args {
		rfType := reflect.TypeOf(arg)
		rfValue := reflect.ValueOf(arg)
		kind := rfValue.Kind()
		if rfType == JsValuePtrType {
			jsArgs = append(jsArgs, arg.(v8go.Valuer))
		} else if kind == reflect.Slice || kind == reflect.Array {
			ar := make([]interface{}, 0)
			for arrayIndex := 0; arrayIndex < rfValue.Len(); arrayIndex++ {
				eleValue0 := rfValue.Index(arrayIndex)
				ar = append(ar, eleValue0.Interface())
			}
			jsValue := t.createJsArrayFromArray(ar)
			jsArgs = append(jsArgs, jsValue)

		} else {
			jsValue, jsValueError := v8go.NewValue(t.Iso, arg)
			if jsValueError != nil {
				panic(jsValueError)
			}
			jsArgs = append(jsArgs, jsValue)
		}
	}
	defer func() {
		tmp := make([]*v8go.Value, len(jsArgs))
		for i, v := range jsArgs {
			tmp[i] = v.(*v8go.Value)
		}
		if len(tmp) > 0 {
			t.Iso.BatchMarkCanReleaseInC(tmp...)
		}
	}()
	ret, err = function.Call(v8go.Undefined(t.Iso), jsArgs...)
	if err != nil {
		fmt.Printf("JS ERROR %+v", err)
		panic(err)
	}
	return ret
}

func (t *ServerEnv) createJsArrayFromArray(data []interface{}) *v8go.Value {
	jsArgs := make([]interface{}, 0)
	for _, arg := range data {
		jsValue, jsValueError := v8go.NewValue(t.Iso, arg)
		if jsValueError != nil {
			panic(jsValueError)
		}
		jsArgs = append(jsArgs, jsValue)
	}
	defer func() {
		tmp := make([]*v8go.Value, len(jsArgs))
		for i, v := range jsArgs {
			tmp[i] = v.(*v8go.Value)
		}
		if len(tmp) > 0 {
			t.Iso.BatchMarkCanReleaseInC(tmp...)
		}
	}()
	return t.GoCallTsEvent("createArray", jsArgs...)
}

//
//func (g *TsEnv) RegClass(obj interface{}, asNameSpace bool) {
//	ty := reflect.TypeOf(obj)
//	numMethod := ty.NumMethod()
//	for index := 0; index < numMethod; index++ {
//		method := ty.Method(index)
//		if asNameSpace {
//			nameSpaceObject, _ := g.CreateEmptyObject()
//			nameSpaceObject.Set(method.Name, v8go.NewFunctionTemplate(g.Iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
//				//method.Func.Call([]reflect.Value{nil})
//				return nil
//			}))
//		}
//	}
//}
