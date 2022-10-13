package env

import (
	"gitee.com/hasika/v8go"
	"reflect"
)

var JsValuePtrType = reflect.TypeOf(&v8go.Value{})
