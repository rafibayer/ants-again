package control

// shared interface between wasm/stub implementations
type Controller interface {
	State() *State
}
