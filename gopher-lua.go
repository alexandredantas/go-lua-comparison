package main

/*
#cgo pkg-config: lua

#include <lua.h>
#include <lualib.h>
#include <lauxlib.h>
#include <stdlib.h>

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
        printf("%p\n",lua_topointer(L,i));
        break;
    }
  }
}
*/
import "C"
import (
	lua "github.com/yuin/gopher-lua"
	"log"
	"time"
	"unsafe"
)

var printtable = `
tb["key"] = "abc"
`

var start = time.Now()
var elapsed = time.Since(start)

func main() {
	L1 := C.luaL_newstate()
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
	C.lua_pushvalue(L1, -1)
	C.lua_setfield(L1, -10002, cTableName)
	C.lua_pushstring(L1, cKey)
	C.lua_pushstring(L1, cValue)
	//C.dumpstack(L1)
	C.lua_settable(L1, -3)

	C.luaL_loadstring(L1, cPrintTable)

	start = time.Now()
	result := C.lua_pcall(L1, 0, -1, 0)

	if int(result) != 0 {
		panic("script error")
	}

	elapsed = time.Since(start)
	log.Printf("Lua C Bindings Execution took %s", elapsed)

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
