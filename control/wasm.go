//go:build js && wasm

package control

import (
	"log"
	"strconv"
	"syscall/js"
)

var (
	// wasm Controller global, returns dom values
	Control Controller = WasmController{domCache: make(map[string]js.Value)}
)

type WasmController struct {
	// map from id to input to cache dom queries and prevent GC
	domCache map[string]js.Value
}

var _ Controller = WasmController{}

func (c WasmController) State() *State {
	return &State{
		AntSpeed:             c.getValue("antSpeed"),
		AntRotation:          c.getValue("antRotation"),
		PheromoneSenseRadius: c.getValue("pheromoneSenseRadius"),
	}
}

// gets value of input, caches in ref. Ref should be global to prevent GC
func (c WasmController) getValue(inputId string) float64 {
	// query dom, populate cache
	ref, ok := c.domCache[inputId]
	if !ok || ref.IsUndefined() {
		ref = js.Global().Get("document").Call("getElementById", inputId)
		c.domCache[inputId] = ref
	}

	val, err := strconv.ParseFloat(ref.Get("value").String(), 64)
	if err != nil {
		log.Fatalf("failed to parse %+v as float64: %+v", ref, err)
	}

	return val

}
