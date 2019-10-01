package guintf

import (
	"fmt"
	"log"
	"path/filepath"
	r "reflect"
	"strings"
)

type Func struct {
	Name      string
	ParamType r.Type

	strParamType,
	strHandlerType,
	strInstructionGetFromStr string
}

type unit struct {
	delphiHandlersTypes map[string]string
	functions           []Func
	types               *unitTypes
	name                string
}

func newUnit(name string, functions []Func) *unit {

	var extension = filepath.Ext(name)
	name = name[0 : len(name)-len(extension)]

	x := &unit{
		types:               new(unitTypes),
		delphiHandlersTypes: make(map[string]string),
		name:                name,
	}

	for _, fun := range functions {

		if fun.ParamType.Kind() == r.Struct && fun.ParamType.NumField() == 0 {
			fun.strHandlerType = "TProcedure"
			x.functions = append(x.functions, fun)
			x.delphiHandlersTypes["TProcedure"] = "reference to procedure"
			continue
		}

		t, err := x.types.addType(fun.ParamType)

		if err != nil {
			log.Fatalln("notify_service:", fun.Name, "error:", err)
		}

		handlerTypeName := strings.Title(t.TypeName() + "Handler")
		if handlerTypeName[0] != 'T' {
			handlerTypeName = "T" + handlerTypeName
		}

		fun.strParamType = t.TypeName()
		fun.strHandlerType = handlerTypeName

		switch fun.ParamType.Kind() {

		case r.String:
			fun.strInstructionGetFromStr = "str"

		case r.Int,
			r.Int8, r.Int16, r.Int32, r.Int64,
			r.Uint8, r.Uint16, r.Uint32, r.Uint64:
			fun.strInstructionGetFromStr = "StrToInt(str)"

		case r.Float32, r.Float64:
			fun.strInstructionGetFromStr = "str_to_float(str)"

		case r.Bool:
			fun.strInstructionGetFromStr = "StrToBool(str)"

		case r.Struct:
			fun.strInstructionGetFromStr = fmt.Sprintf("_deserializer.deserialize<%s>(str)", t.TypeName())
		default:
			panic(fmt.Sprintf("wrong type %q: %+v", fun.Name, fun.ParamType))
		}

		x.delphiHandlersTypes[fun.strHandlerType] = fmt.Sprintf("reference to procedure (x:%s)", fun.strParamType)
		x.functions = append(x.functions, fun)
	}
	return x
}
