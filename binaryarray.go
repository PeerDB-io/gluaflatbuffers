package gluaflatbuffers

import (
	"github.com/yuin/gopher-lua"

	"github.com/PeerDB-io/glua64"
)

var LuaBinaryArray = glua64.UserDataType[[]byte]{Name: "flatbuffers_binaryarray"}

func BinaryArray_Loader(ls *lua.LState) int {
	m := ls.NewTable()
	ls.SetField(m, "New", ls.NewFunction(BinaryArrayNew))

	mt := LuaBinaryArray.NewMetatable(ls)
	ls.SetField(mt, "__index", ls.NewFunction(BinaryArrayIndex))
	ls.SetField(mt, "__len", ls.NewFunction(BinaryArrayLen))
	ls.SetField(mt, "Slice", ls.NewFunction(BinaryArraySlice))
	ls.SetField(mt, "Grow", ls.NewFunction(BinaryArrayGrow))
	ls.SetField(mt, "Pad", ls.NewFunction(BinaryArrayPad))
	ls.SetField(mt, "Set", ls.NewFunction(BinaryArraySet))

	ls.Push(m)
	return 1
}

func BinaryArrayNew(ls *lua.LState) int {
	lval := ls.Get(1)
	var ba []byte
	switch val := lval.(type) {
	case lua.LString:
		ba = []byte(val)
	case lua.LNumber:
		ba = make([]byte, int(val))
	default:
		ls.RaiseError("Expect a integer size value or string to construct a binary array")
		return 0
	}
	ls.Push(LuaBinaryArray.New(ls, ba))
	return 1
}

func BinaryArrayLen(ls *lua.LState) int {
	ba := LuaBinaryArray.StartMethod(ls)
	ls.Push(lua.LNumber(len(ba)))
	return 1
}

func BinaryArrayIndex(ls *lua.LState) int {
	ba, key := LuaBinaryArray.StartIndex(ls)
	switch key {
	case "size":
		ls.Push(lua.LNumber(len(ba)))
	case "str":
		ls.Push(lua.LString(ba))
	case "data":
		ls.RaiseError("BinaryArray data property inaccessible")
		return 0
	default:
		ls.Push(ls.GetField(LuaBinaryArray.Metatable(ls), key))
	}
	return 1
}

func BinaryArraySlice(ls *lua.LState) int {
	var startPos, endPos int
	ba := LuaBinaryArray.StartMethod(ls)
	if luaStartPos, ok := ls.Get(2).(lua.LNumber); ok {
		startPos = int(luaStartPos)
		if startPos < 0 {
			startPos = 0
		}
	} else {
		startPos = 0
	}
	if luaEndPos, ok := ls.Get(3).(lua.LNumber); ok {
		endPos = min(int(luaEndPos), len(ba))
	} else {
		endPos = len(ba)
	}
	ls.Push(lua.LString(ba[startPos:endPos]))
	return 1
}

func BinaryArrayGrow(ls *lua.LState) int {
	baud, ba := LuaBinaryArray.Check(ls, 1)
	newsize := int(ls.CheckNumber(2))
	if newsize > len(ba) {
		newdata := make([]byte, newsize)
		copy(newdata[newsize-len(ba):], ba)
		baud.Value = newdata
	}
	return 0
}

func Pad(ba []byte, n int, start int) {
	for i := range n {
		ba[start+i] = 0
	}
}

func BinaryArrayPad(ls *lua.LState) int {
	ba := LuaBinaryArray.StartMethod(ls)
	n := int(ls.CheckNumber(2))
	startPos := int(ls.CheckNumber(3))
	Pad(ba, n, startPos)
	return 0
}

func BinaryArraySet(ls *lua.LState) int {
	ba := LuaBinaryArray.StartMethod(ls)
	idx := int(ls.CheckNumber(3))
	value := ls.Get(2)
	if num, ok := value.(lua.LNumber); ok {
		ba[idx] = byte(num)
	}
	if str, ok := value.(lua.LString); ok {
		ba[idx] = str[0]
	}
	return 0
}
