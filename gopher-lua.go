package main

import "C"
import (
	"errors"
	lua "github.com/yuin/gopher-lua"
	"log"
	"reflect"
	"time"
	"unsafe"
)

/*
#cgo pkg-config: lua

#include <lua.h>
#include <lualib.h>
#include <lauxlib.h>
#include <stdlib.h>
#include <stdio.h>
#include <string.h>

typedef struct MAP {
  int keyType;
  int valueType;

  void* key;
  void* value;

  struct MAP* previous;
  struct MAP* next;
} MAP;

static void store(void *destination, void *source, size_t size) {
    memcpy(destination, source, size);
}

static MAP* parseTable(lua_State *L, int t);

static int populateNode(MAP* node, int keyOrValue, lua_State *L, int index){
	void* destination;
	int result;
	switch (lua_type(L, index)) {
      case LUA_TNUMBER: ;
		lua_Number nb = lua_tonumber(L, index);
		destination = malloc(sizeof(lua_Number));
        store(destination, &nb, sizeof(lua_Number));
        result = LUA_TNUMBER;
		break;
      case LUA_TSTRING: ;
		char* str = lua_tostring(L, index);
		destination = malloc(sizeof(char*));
		store(destination, str, sizeof(char*));
        result = LUA_TSTRING;
		break;
      case LUA_TBOOLEAN: ;
        int b = lua_toboolean(L, index);
		destination = malloc(sizeof(int));
		store(destination, &b, sizeof(int));
        result = LUA_TBOOLEAN;
		break;
	  case LUA_TTABLE:
		destination = parseTable(L, index);
		result = LUA_TTABLE;
		break;
      default:
        result = -1;
		break;
    }

	if (keyOrValue == 0){
		node->key = destination;
	} else {
		node->value = destination;
	}

	return result;
}

static MAP* parseTable(lua_State *L, int t){
	lua_pushnil(L);
    MAP* current = malloc(sizeof(MAP));
	current->next = NULL;
	current->previous = NULL;
	current->key = NULL;
	current->value = NULL;

	while (lua_next(L, t) != 0) {
        lua_pushvalue(L, -2);
		current->keyType = populateNode(current, 0, L, lua_gettop(L));
		current->valueType = populateNode(current, 1, L, lua_gettop(L) - 1);
		if (current->keyType == -1 || current->valueType == -1){
			return NULL;
		}
		lua_pop(L, 2);

		current->next = malloc(sizeof(MAP));
		current->next->previous = current;
		current = current->next;
        current->next = NULL;
		current->key = NULL;
		current->value = NULL;
	}

	return current;
}

static void dumpstack (lua_State *L) {
  int top=lua_gettop(L);
  for (int i=1; i <= top; i++) {
    printf("%d\t%s\t", i, luaL_typename(L,i));
    switch (lua_type(L, i)) {
      case LUA_TNUMBER:
        printf("%g\n",lua_tonumber(L,i));
        break;
      case LUA_TSTRING:
        printf("%s\n",lua_tostring(L,i));
        break;
      case LUA_TBOOLEAN:
        printf("%s\n", (lua_toboolean(L, i) ? "true" : "false"));
        break;
      case LUA_TNIL:
        printf("%s\n", "nil");
        break;
      default:
        printf("pointer %p\n",lua_topointer(L,i));
        break;
    }
  }
}
*/
import "C"

var printtable = `
return testmp
`

var start = time.Now()
var elapsed = time.Since(start)

type Test struct {
	Z string
}

func main() {
	var L1 *C.struct_lua_State
	L1 = C.luaL_newstate()

	defer C.lua_close(L1)

	C.luaL_openlibs(L1)

	cKey := C.CString("key")
	cValue := C.CString("value")
	cPrintTable := C.CString(printtable)
	cTableName := C.CString("tb")

	defer C.free(unsafe.Pointer(cKey))
	defer C.free(unsafe.Pointer(cValue))
	defer C.free(unsafe.Pointer(cPrintTable))
	defer C.free(unsafe.Pointer(cTableName))

	C.lua_createtable(L1, 0, 0)
	var tableIndex C.int
	tableIndex = C.lua_gettop(L1)

	testmp := map[interface{}]interface{}{"a": 1, 1: "b", true: 10.0, "c": []string{"x", "y", "z"}, "d": Test{Z: "struct"}}

	Put(L1, "testmp", testmp)

	C.lua_pushvalue(L1, tableIndex)
	C.lua_setfield(L1, -10002, cTableName)
	C.lua_pushstring(L1, cKey)
	C.lua_pushstring(L1, cValue)
	//C.dumpstack(L1)
	C.lua_settable(L1, -3)

	C.luaL_loadstring(L1, cPrintTable)

	start = time.Now()
	result := C.lua_pcall(L1, 0, -1, 0)

	if int(result) != 0 {
		panic(C.GoString(C.lua_tolstring(L1, -1, nil)))
	}

	elapsed = time.Since(start)
	log.Printf("Lua C Bindings Execution took %s", elapsed)

	res, err := toGoValue(L1)
	if err != nil {
		panic(err)
	}
	println(res)

	mp := make(map[string]string)

	mp["key"] = "value"

	start = time.Now()
	mp["key"] = "abc"
	elapsed = time.Since(start)
	log.Printf("Native access Execution took %s", elapsed)

	L := lua.NewState()

	table := L.NewTable()

	table.RawSet(lua.LString("key"), lua.LString("value"))

	L.SetGlobal("tb", table)

	defer L.Close()
	start = time.Now()
	if err := L.DoString(printtable); err != nil {
		panic(err)
	}
	elapsed = time.Since(start)
	log.Printf("Gopher-lua Execution took %s", elapsed)
}

func Put(L *C.struct_lua_State, name string, value interface{}) {
	pushValueToStack(L, value)
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	C.lua_setfield(L, -10002, cName)
}

func pushValueToStack(L *C.struct_lua_State, value interface{}) {
	switch val := reflect.ValueOf(value); val.Kind() {
	case reflect.Bool:
		var cValue C.int
		if value.(bool) {
			cValue = C.int(1)
		} else {
			cValue = C.int(0)
		}

		C.lua_pushboolean(L, cValue)
		break
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		cValue := C.long(val.Int())
		C.lua_pushinteger(L, cValue)
		break
	case reflect.Float32, reflect.Float64:
		cValue := C.double(val.Float())
		C.lua_pushnumber(L, cValue)
		break
	case reflect.Map:
		C.lua_createtable(L, 0, 0)
		for k, v := range value.(map[interface{}]interface{}) {
			pushValueToStack(L, k)
			pushValueToStack(L, v)
			C.lua_settable(L, -3)
		}
		break
	case reflect.Struct:
		C.lua_createtable(L, 0, 0)
		for i := 0; i < val.NumField(); i++ {
			if val.Field(i).CanInterface() {
				varValue := val.Field(i).Interface()
				pushValueToStack(L, val.Type().Field(i).Name)
				pushValueToStack(L, varValue)
				C.lua_settable(L, -3)
			}
		}
		break
	case reflect.Array, reflect.Slice:
		C.lua_createtable(L, 0, 0)
		for i := 0; i < val.Len(); i++ {
			v := val.Index(i).Interface()
			pushValueToStack(L, i)
			pushValueToStack(L, v)
			C.lua_settable(L, -3)
		}
		break
	case reflect.String:
		cValue := C.CString(value.(string))
		defer C.free(unsafe.Pointer(cValue))
		C.lua_pushstring(L, cValue)
		break
	default:
		panic(errors.New("not supported type"))
	}
}

func toGoValue(luaStack *C.struct_lua_State) (res interface{}, err error) {
	valueIdx := C.lua_gettop(luaStack)

	switch C.lua_type(luaStack, valueIdx) {
	case C.LUA_TNIL:
		return nil, nil
	case C.LUA_TBOOLEAN:
		return int(C.lua_toboolean(luaStack, valueIdx)) == 1, nil
	case C.LUA_TNUMBER:
		return float64(C.lua_tonumber(luaStack, valueIdx)), nil
	case C.LUA_TSTRING:
		return C.GoString(C.lua_tolstring(luaStack, valueIdx, nil)), nil
	case C.LUA_TTABLE:
		result := convertCTableToGoMap(C.parseTable(luaStack, valueIdx))
		return result, nil
	default:
		return nil, errors.New("unknown type")
	}
}

func convertCTableToGoMap(table *C.struct_MAP) map[interface{}]interface{} {
	result := make(map[interface{}]interface{})

	var last *C.struct_MAP

	for current := table.previous; current != nil; current, last = current.previous, current {
		if current.next != nil {
			C.free(unsafe.Pointer(current.next))
		}
		result[convertCValueToGoValue(unsafe.Pointer(current.key), int(current.keyType))] =
			convertCValueToGoValue(unsafe.Pointer(current.value), int(current.valueType))
		C.free(unsafe.Pointer(current.key))
		C.free(unsafe.Pointer(current.value))
	}

	C.free(unsafe.Pointer(last.key))
	C.free(unsafe.Pointer(last.value))
	C.free(unsafe.Pointer(last))

	return result
}

func convertCValueToGoValue(ptr unsafe.Pointer, valueType int) interface{} {
	switch valueType {
	case C.LUA_TBOOLEAN:
		return *(*int)(ptr) == 1
	case C.LUA_TSTRING:
		return C.GoString((*C.char)(ptr))
	case C.LUA_TNUMBER:
		return *((*float64)(ptr))
	case C.LUA_TTABLE:
		return convertCTableToGoMap((*C.struct_MAP)(ptr))
	default:
		return nil
	}
}
