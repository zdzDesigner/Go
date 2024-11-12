package main

import (
	"reflect"
	"time"
)

type Material struct {
	ID         int       `json:"id"`
	TableZh    string    `json:"table_zh"`
	TableEn    string    `json:"table_en"`
	TableDescr string    `json:"table_descr"`
	Createtime time.Time `json:"createtime"`
	Updatetime time.Time `json:"updatetime"`
}

func main() {

	keys := []string{"id", "table_zh", "table_en", "table_descr", "createtime", "updatetime"}
	model(keys)

	reflectModel(&Material{}, keys)
}

func model(keys []string) (material *Material, vals []any) {
	vals = []any{}
	material = &Material{}
	for _, key := range keys {
		if key == "id" {
			vals = append(vals, &material.ID)
		}
		if key == "table_en" {
			vals = append(vals, &material.TableEn)
		}
		if key == "table_zh" {
			vals = append(vals, &material.TableZh)
		}
		if key == "table_descr" {
			vals = append(vals, &material.TableDescr)
		}
		if key == "createtime" {
			vals = append(vals, &material.Createtime)
		}
		if key == "updatetime" {
			vals = append(vals, &material.Updatetime)
		}
	}
	return
}

func reflectModel(tb any, keys []string) (vals []any) {
	rt := reflect.TypeOf(tb)
	rv := reflect.ValueOf(tb)
	// 指针
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
		rv = rv.Elem()
	}
	// fmt.Println("reflect TypeOf:", rt.NumField(), "name:", rt.Name())
	// fmt.Println("reflect ValueOf:", rv)

	for i := 0; i < rt.NumField(); i++ {
		t := rt.Field(i)
		v := rv.Field(i)
		tag := t.Tag.Get("json")
		// fmt.Printf("t:%+v,v:%+v\n:", t, v)
		// fmt.Println("t:", t, t.Tag, "tag.Get:", t.Tag.Get("json"), "v:", v)
		// fmt.Println("t.Tag.Get:", t.Tag.Get("json"))
		// fmt.Println("type:", reflect.New(t.Type))

		if Contains[string](keys, tag) {
			vals = append(vals, v.Addr().Interface())
		}
	}
	// fmt.Println("vals len:", len(vals))
	return

}

func Contains[T comparable](collection []T, element T) bool {
	for _, item := range collection {
		if item == element {
			return true
		}
	}

	return false
}
