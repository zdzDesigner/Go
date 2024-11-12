func FieldFilter2(tb any, keys []string) []any {
	rt := reflect.TypeOf(tb)
	rv := reflect.ValueOf(tb)
	fmt.Println("typeof tb", rt, rt.Kind() == reflect.Map)
	fmt.Println("valueof tb", rv, rv.Kind() == reflect.Map)

	fmt.Println("mapkeys:", rv.MapKeys())
	vals := make([]any, 0, len(keys))
	fmt.Printf("tb:%+v\n", tb)

	// if pm, ok := tb.(IsMaper); ok {
	// 	fmt.Println("IsMaper:", pm, pm.Map())
	// 	m := pm.Map()
	// 	for _, key := range keys {
	// 		// fmt.Println(m[key])
	// 		vals = append(vals, m[key])
	//
	// 	}
	// }
	// it := rv.MapRange()
	// for it.Next() {
	// 	fmt.Println(it.Key(), it.Key().CanAddr(), it.Value(), it.Value().CanAddr())
	// }

	for _, key := range keys {
		for _, k := range rv.MapKeys() {
			if key != k.String() {
				continue
			}

			// key := k.Convert(newInstance.Type().Key())
			val := rv.MapIndex(k)
			fmt.Println("elem:", k.Kind(), k.String(), k.CanAddr(), val, val.CanAddr(), val.Kind())
			if val.IsValid() {
				// fmt.Println(val.Addr().Interface())
			}
			// fmt.Println(newval)

			// newval.Elem().Set(reflect.ValueOf(val.Interface()))

			// v := reflect.ValueOf(val.Interface())
			// fmt.Println(v, v.CanAddr(), v.Kind())
			// vals = append(vals, v.Interface())
			vals = append(vals, val.Interface())

			// // fmt.Println(val, val.String(), val.Kind(), val.Elem().Interface(), val.Elem().Elem())
			// // if util.Contains[string](keys, k.String()) {
			// newval := reflect.New(val.Type())
			// rv.SetMapIndex(k, newval)
			// // vals = append(vals, &v)
			// // vals = append(vals, &val)
			// vals = append(vals, newval.Interface())
			// // }
		}
	}

	// fmt.Println("mapkeys:", rv.MapRange())
	// mm := reflect.MakeMap(rt)
	// fmt.Println(mm.MapKeys())

	// vals := make([]any, 0, len(keys))
	// fmt.Printf("tb:%+v\n", tb)
	// mapkv, ok := tb.(map[string]any)
	//
	// for k, v := range mapkv {
	// 	fmt.Println(k)
	// 	if util.Contains[string](keys, k) {
	// 		mapkv[k] = &v
	// 		vals = append(vals, &v)
	// 	}
	// }
	// fmt.Println("vals:", vals, len(vals), mapkv, ok)
	return vals

}




func FieldFilter(tb any, keys []string) []any {
	vals := make([]any, 0, len(keys))
	fmt.Printf("tb:%+v\n", tb)

	if pm, ok := tb.(IsMaper); ok {
		// fmt.Println("IsMaper:", pm, pm.Map())
		m := pm.MapNew()
		for _, key := range keys {
			vals = append(vals, m[key])
		}
	}

	// rt := reflect.TypeOf(tb)
	// rv := reflect.ValueOf(tb)
	// fmt.Println("typeof tb", rt, rt.Kind() == reflect.Map)
	// fmt.Println("valueof tb", rv, rv.Kind() == reflect.Map)
	//
	// fmt.Println("mapkeys:", rv.MapKeys())
	// for _, key := range keys {
	// 	for _, k := range rv.MapKeys() {
	// 		if key != k.String() {
	// 			continue
	// 		}
	//
	// 		val := rv.MapIndex(k)
	// 		vals = append(vals, val.Interface())
	// 	}
	// }

	return vals

}


func FieldFilter(tb any, keys []string) []any {
	vals := make([]any, 0, len(keys))

	// = map ================
	if pm, ok := tb.(IsMaper); ok {
		m := pm.MapNew()
		for _, key := range keys {
			vals = append(vals, m[key])
		}
		return vals
	}

	// = struct ================
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

		if util.Contains[string](keys, tag) {
			vals = append(vals, v.Addr().Interface())
		}
	}
	// fmt.Println("vals len:", len(vals))
	return vals

}

