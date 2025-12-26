package compiler

import (
	"github.com/arc-language/core-builder/ir"
)

// Symbol represents a named value in the symbol table
type Symbol struct {
	Name      string
	Value     ir.Value
	IsConst   bool
	Namespace string // Which namespace this symbol belongs to
}

// Scope represents a lexical scope with symbol table
type Scope struct {
	parent  *Scope
	symbols map[string]*Symbol
}

// NewScope creates a new scope
func NewScope(parent *Scope) *Scope {
	return &Scope{
		parent:  parent,
		symbols: make(map[string]*Symbol),
	}
}

// Define adds a symbol to the current scope
func (s *Scope) Define(name string, value ir.Value) {
	s.symbols[name] = &Symbol{
		Name:    name,
		Value:   value,
		IsConst: false,
	}
}

// DefineConst adds a constant symbol to the current scope
func (s *Scope) DefineConst(name string, value ir.Value) {
	s.symbols[name] = &Symbol{
		Name:    name,
		Value:   value,
		IsConst: true,
	}
}

// DefineInNamespace adds a symbol with namespace information
func (s *Scope) DefineInNamespace(name string, value ir.Value, namespace string) {
	s.symbols[name] = &Symbol{
		Name:      name,
		Value:     value,
		IsConst:   false,
		Namespace: namespace,
	}
}

// Lookup searches for a symbol in the current scope and parent scopes
func (s *Scope) Lookup(name string) (*Symbol, bool) {
	// Check current scope
	if sym, ok := s.symbols[name]; ok {
		return sym, true
	}
	
	// Check parent scopes
	if s.parent != nil {
		return s.parent.Lookup(name)
	}
	
	return nil, false
}

// LookupLocal searches only the current scope (not parents)
func (s *Scope) LookupLocal(name string) (*Symbol, bool) {
	sym, ok := s.symbols[name]
	return sym, ok
}

// IsDefined checks if a symbol exists in current scope
func (s *Scope) IsDefined(name string) bool {
	_, ok := s.symbols[name]
	return ok
}