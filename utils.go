package govue

import (
	"log"
	"net/http"
	"reflect"
	"unsafe"
)

func ferr(tag string, err error) bool {
	if err != nil {
		log.Fatalf("[%s] %s.\n", tag, err)
		return true
	}
	return false
}

func logerr(tag string, err error) bool {
	if err != nil {
		log.Printf("[%s] %s.\n", tag, err)
	}
	return false
}

func werr(err error, w http.ResponseWriter) bool {
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return true
	}
	return false
}

func getUnexportedField(field reflect.Value) interface{} {
	return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Interface()
}
