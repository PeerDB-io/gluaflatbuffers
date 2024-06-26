package gluaflatbuffers

import (
	"github.com/yuin/gopher-lua"

	"github.com/PeerDB-io/glua64"
)

type View struct {
	ba        []byte
	pos       int // 0-based offset
	vtable    int // 0-based offset
	vtableEnd uint16
	hasv      bool
}

var LuaView = glua64.UserDataType[*View]{Name: "flatbuffers_view"}

func CheckOffset(ls *lua.LState, idx int) int {
	num := ls.CheckNumber(idx)
	if num < 0 || num > 42949672951 {
		ls.RaiseError("Offset is not valid")
	}
	return int(num)
}

func View_Loader(ls *lua.LState) int {
	m := ls.NewTable()
	ls.SetField(m, "New", ls.NewFunction(ViewNew))

	mt := LuaView.NewMetatable(ls)
	ls.SetField(mt, "__index", ls.NewFunction(ViewIndex))

	ls.Push(m)
	return 1
}

func ViewNew(ls *lua.LState) int {
	buf := ls.Get(1)
	var ba []byte
	switch val := buf.(type) {
	case lua.LString:
		ba = []byte(val)
	case *lua.LUserData:
		var ok bool
		ba, ok = val.Value.([]byte)
		if !ok {
			ls.RaiseError("invalid buf userdata passed to view.New")
			return 0
		}
	default:
		ls.RaiseError("invalid buf passed to view.New")
		return 0
	}
	ls.Push(LuaView.New(ls, &View{
		ba:  ba,
		pos: CheckOffset(ls, 2),
	}))
	return 1
}

func ViewIndex(ls *lua.LState) int {
	view, key := LuaView.StartIndex(ls)
	switch key {
	case "bytes":
		ls.Push(LuaBinaryArray.New(ls, view.ba))
	case "pos":
		ls.Push(lua.LNumber(view.pos))
	case "Offset":
		ls.Push(ls.NewFunction(ViewOffset))
	case "Indirect":
		ls.Push(ls.NewFunction(ViewIndirect))
	case "String":
		ls.Push(ls.NewFunction(ViewString))
	case "VectorLen":
		ls.Push(ls.NewFunction(ViewVectorLen))
	case "Vector":
		ls.Push(ls.NewFunction(ViewVector))
	case "VectorAsString":
		ls.Push(ls.NewFunction(ViewVectorAsString))
	case "Union":
		ls.Push(ls.NewFunction(ViewUnion))
	case "Get":
		ls.Push(ls.NewFunction(ViewGet))
	case "GetSlot":
		ls.Push(ls.NewFunction(ViewGetSlot))
	case "GetVOffsetTSlot":
		ls.Push(ls.NewFunction(ViewGetVOffsetTSlot))
	}
	return 1
}

func (view *View) Offset(vtoff uint16) uint16 {
	if !view.hasv {
		view.vtable = view.pos - int(int32(int32n.UnpackU64(view.ba[view.pos:])))
		view.vtableEnd = uint16(uint16n.UnpackU64(view.ba[view.vtable:]))
		view.hasv = true
	}
	if vtoff < view.vtableEnd {
		return uint16(uint16n.UnpackU64(view.ba[view.vtable+int(vtoff):]))
	} else {
		return 0
	}
}

func (view *View) Vector(off int) int {
	off += view.pos
	return off + int(uint32n.UnpackU64(view.ba[off:])) + 4
}

func (view *View) VectorLen(off int) uint32 {
	off += int(uint32n.UnpackU64(view.ba[view.pos+off:]))
	return uint32(uint32n.UnpackU64(view.ba[off:]))
}

func ViewOffset(ls *lua.LState) int {
	view := LuaView.StartMethod(ls)
	vtoff := uint16(CheckOffset(ls, 2))
	ls.Push(lua.LNumber(view.Offset(vtoff)))
	return 1
}

func ViewIndirect(ls *lua.LState) int {
	view := LuaView.StartMethod(ls)
	off := CheckOffset(ls, 2)
	ls.Push(lua.LNumber(off + int(uint32n.UnpackU64(view.ba[off:]))))
	return 1
}

func ViewString(ls *lua.LState) int {
	view := LuaView.StartMethod(ls)
	off := CheckOffset(ls, 2)
	off += int(uint32n.UnpackU64(view.ba[off:]))
	start := off + 4
	length := int(uint32n.UnpackU64(view.ba[off:]))
	ls.Push(lua.LString(view.ba[start : start+length]))
	return 1
}

func ViewVectorLen(ls *lua.LState) int {
	view := LuaView.StartMethod(ls)
	off := CheckOffset(ls, 2)
	ls.Push(lua.LNumber(view.VectorLen(off)))
	return 1
}

func ViewVector(ls *lua.LState) int {
	view := LuaView.StartMethod(ls)
	off := CheckOffset(ls, 2)
	ls.Push(lua.LNumber(view.Vector(off)))
	return 1
}

func ViewVectorAsString(ls *lua.LState) int {
	view := LuaView.StartMethod(ls)
	off := uint16(CheckOffset(ls, 2))
	o := view.Offset(off)
	if o == 0 {
		ls.Push(lua.LNil)
		return 1
	}
	var start, stop int
	lstart, ok := ls.Get(3).(lua.LNumber)
	if ok {
		start = int(lstart)
	}
	lstop, ok := ls.Get(4).(lua.LNumber)
	if ok {
		stop = int(lstop)
	} else {
		stop = int(view.VectorLen(int(o)))
	}
	a := view.Vector(int(o)) + start
	ls.Push(lua.LString(view.ba[a : a+stop-start]))
	return 1
}

func ViewUnion(ls *lua.LState) int {
	view := LuaView.StartMethod(ls)
	t2ud, t2 := LuaView.Check(ls, 2)
	off := CheckOffset(ls, 3)
	off += view.pos
	t2.pos = off + int(uint32n.UnpackU64(view.ba[off:]))
	t2.ba = view.ba
	t2ud.Value = t2
	return 0
}

func ViewGet(ls *lua.LState) int {
	view := LuaView.StartMethod(ls)
	_, n := LuaN.Check(ls, 2)
	off := CheckOffset(ls, 3)
	ls.Push(n.Unpack(ls, view.ba[off-1:]))
	return 1
}

func ViewGetSlot(ls *lua.LState) int {
	view := LuaView.StartMethod(ls)
	slot := uint16(CheckOffset(ls, 2))
	off := view.Offset(slot)
	if off == 0 {
		ls.Push(ls.Get(3))
		return 1
	}
	_, validatorFlags := LuaN.Check(ls, 4)
	ls.Push(validatorFlags.Unpack(ls, view.ba[view.pos+int(off):]))
	return 1
}

func ViewGetVOffsetTSlot(ls *lua.LState) int {
	view := LuaView.StartMethod(ls)
	slot := uint16(CheckOffset(ls, 2))
	off := view.Offset(slot)
	if off == 0 {
		ls.Push(ls.Get(3))
	} else {
		ls.Push(lua.LNumber(off))
	}
	return 1
}
