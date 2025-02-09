package adapter

import (
	"encoding/base64"
	"fmt"
	"sync/atomic"

	"github.com/aarzilli/golua/lua"
)

// TODO: aSec sync is enough?
type Lua struct {
	state         *lua.State
	packetCounter atomic.Int64
	base64LuaCode string
}

type LuaParams struct {
	Base64LuaCode string
}

func NewLua(params LuaParams) (*Lua, error) {
	luaCode, err := base64.StdEncoding.DecodeString(params.Base64LuaCode)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(luaCode))

	state := lua.NewState()
	state.OpenLibs()

	if err := state.DoString(string(luaCode)); err != nil {
		return nil, fmt.Errorf("Error loading Lua code: %v\n", err)
	}
	return &Lua{state: state, base64LuaCode: params.Base64LuaCode}, nil
}

func (l *Lua) Close() {
	l.state.Close()
}

func (l *Lua) Generate(
	msgType int64,
	data []byte,
) ([]byte, error) {
	l.state.GetGlobal("d_gen")

	l.state.PushInteger(msgType)
	l.state.PushBytes(data)
	l.state.PushInteger(l.packetCounter.Add(1))

	if err := l.state.Call(3, 1); err != nil {
		return nil, fmt.Errorf("Error calling Lua function: %v\n", err)
	}

	result := l.state.ToBytes(-1)
	l.state.Pop(1)

	return result, nil
}

func (l *Lua) Parse(data []byte) ([]byte, error) {
	l.state.GetGlobal("d_parse")

	l.state.PushBytes(data)
	if err := l.state.Call(1, 1); err != nil {
		return nil, fmt.Errorf("Error calling Lua function: %v\n", err)
	}

	result := l.state.ToBytes(-1)
	l.state.Pop(1)
	// copy(data, result)

	return result, nil
}

func (l *Lua) Base64LuaCode() string {
	return l.base64LuaCode
}
