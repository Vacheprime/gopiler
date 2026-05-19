package glushkov

import (
	gp "github.com/Vacheprime/gopiler"
)

type Symbol struct {
	SymbolId rune
	Repr     rune
}

type SymbolPair struct {
	S1 Symbol
	S2 Symbol
}

type SymbolInformation struct {
	StartSymbols gp.Set[Symbol]
	FinalSymbols gp.Set[Symbol]
	SymbolPairs  gp.Set[SymbolPair]
}
