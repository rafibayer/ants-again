//go:build !js

// this file defines stubs for non-wasm compilation targets
package control

var (
	// stub Controller global, always returns nil
	Control Controller = Stub{}
)

type Stub struct{}

var _ Controller = Stub{}

func (Stub) State() *State {
	return nil
}
