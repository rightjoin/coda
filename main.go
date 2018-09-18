package main

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/alecthomas/template"
	"github.com/rightjoin/dorm"

	"github.com/rightjoin/admire/api"
	"github.com/rightjoin/fuel"
	"github.com/rightjoin/utila/refl"
	"github.com/rightjoin/utila/txt"
)

var model = &api.Filter{}

func main() {

	if reflect.TypeOf(model).Kind() != reflect.Ptr {
		fmt.Println("Address of model expected")
		return
	}

	name := reflect.TypeOf(model).Elem().Name()
	abbr := func() string {
		out := ""
		for _, c := range name {
			s := string(c)
			if s == strings.ToUpper(s) {
				out += s
			}
		}
		return strings.ToLower(out)
	}()

	data := map[string]interface{}{
		// Variable Naming
		"VarSing":  abbr,
		"VarPlur":  abbr + "s",
		"VarContr": abbr[0:1] + "c",

		// Core Model & Table
		"Model": name,
		"Table": tableName(model),

		// Behaviors
		"IsDyn":      refl.ComposedOf(model, dorm.DynamicField{}),
		"IsSM":       refl.ComposedOf(model, dorm.Stateful{}),
		"IsImgMulti": refl.ComposedOf(model, dorm.ImgMulti{}),
		"HasIns": func() bool {
			var intf interface{} = *model
			_, valid := (intf).(insertChecks)
			return valid
		}(),
		"HasUpd": func() bool {
			var intf interface{} = *model
			_, valid := (intf).(updateChecks)
			return valid
		}(),
		"HasInsUpd": func() bool {
			var intf interface{} = *model
			_, valid := (intf).(writeChecks)
			return valid
		}(),
		"HasImg": func() bool {
			imgStr := refl.Signature(reflect.TypeOf(dorm.Img{}))
			for _, fld := range refl.NestedFields(*model) {
				if refl.Signature(fld.Type) == imgStr {
					return true
				}
			}
			return false
		}(),
		"HasWho": refl.ComposedOf(model, dorm.WhosThat{}),
	}

	t, err := template.New("codegen.txt").ParseFiles("./codegen.txt")
	if err != nil {
		fmt.Println(err)
		return
	}

	err = t.Execute(os.Stdout, data)
	if err != nil {
		fmt.Println(err)
	}

	//fmt.Println(data)
}

// tableName gets the table name of the given model
func tableName(model interface{}) string {
	t := reflect.TypeOf(model)
	v := reflect.ValueOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	if _, ok := t.MethodByName("TableName"); ok {
		name := v.MethodByName("TableName").Call([]reflect.Value{})
		return name[0].String()
	}
	return txt.CaseSnake(t.Name())
}

type insertChecks interface {
	BeforeInsert(a fuel.Aide) error
}

type updateChecks interface {
	BeforeUpdate(a fuel.Aide) error
}

type writeChecks interface {
	BeforeInsertUpdate(a fuel.Aide) error
}
