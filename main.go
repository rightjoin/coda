package main

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/alecthomas/template"
	"github.com/rightjoin/dorm"

	"github.com/rightjoin/rutl/conv"
	"github.com/rightjoin/rutl/refl"
	"gitlabtc.futuregroup.in/tcommerce/backend/skeleton-svc/api"
)

var model = &api.Model{}

func main() {

	if reflect.TypeOf(model).Kind() != reflect.Ptr {
		fmt.Println("Address of model expected")
		return
	}

	// ask user to use id or uid in generated apis
	pkey := func() string {
		obj := reflect.TypeOf(model).Elem()
		_, okid := obj.FieldByName("ID")
		_, okuid := obj.FieldByName("UID")
		if okid && !okuid {
			return "id"
		} else if !okid && okuid {
			return "uid"
		}
		// Ask user
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Which key should be used? uid or id:")
		text, _ := reader.ReadString('\n')
		return strings.TrimSpace(text)
	}()

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
		"VarContr": abbr[0:1] + "v",

		// Core Model & Table
		"Model": name,
		"Table": dorm.Table(model),
		"Root":  strings.Replace(dorm.Table(model), "_", "-", -1),

		// Unique ID or Primary Key related
		"Key": pkey,
		"KeyRoute": func() string {
			if pkey == "id" {
				return "{id:[0-9]+}"
			}
			return "{uid}"
		}(),
		"KeyType": func() string {
			if pkey == "id" {
				return "uint"
			}
			return "string"
		}(),

		// Behaviors
		"IsDyn": refl.ComposedOf(model, dorm.DynamicField{}),
		"IsSM":  refl.ComposedOf(model, dorm.Stateful{}),
		"HasFile": func() bool {
			fileStr := refl.Signature(reflect.TypeOf(dorm.File{}))
			for _, fld := range refl.NestedFields(*model) {
				if refl.Signature(fld.Type) == "*"+fileStr {
					return true
				}
			}
			return false
		}(),
		"HasWho": refl.ComposedOf(model, dorm.WhosThat{}),
		"Files": func() map[string]map[string]string {
			out := map[string]map[string]string{}
			flds := refl.NestedFields(model)
			for _, f := range flds {
				if f.Type == reflect.TypeOf(&dorm.Files{}) {
					// map[whats-this] : WhatsThis
					meth := conv.CaseSentence(f.Name)
					out[f.Name] = map[string]string{
						"url":  conv.CaseURL(f.Name),
						"meth": strings.ToLower(meth[0:1]) + meth[1:],
						"db":   conv.CaseSnake(f.Name),
					}
					//out[conv.CaseSnake(f.Name)] = f.Name
				}
			}
			return out
		}(),
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

// // Table get the table name of the given model
// func Table(model interface{}) string {
// 	t := reflect.TypeOf(model)
// 	v := reflect.ValueOf(model)
// 	if t.Kind() == reflect.Ptr {
// 		t = t.Elem()
// 		v = v.Elem()
// 	}

// 	if _, ok := t.MethodByName("TableName"); ok {
// 		name := v.MethodByName("TableName").Call([]reflect.Value{})
// 		return name[0].String()
// 	}
// 	return conv.CaseSnake(t.Name())
// }
