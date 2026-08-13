package regex

// Symbol represents a regex symbol.
type Symbol struct {
	SymbolId int
	Token    *RegexToken
	CharSet  *CharacterSet
	Label    string
}

// BuildSymbol creates a regex symbol from a regex token.
func BuildSymbol(tk *RegexToken, symbolId int, label string) (Symbol, error) {
	sym := Symbol{SymbolId: symbolId, Token: tk, Label: label}
	switch tk.Class {
	case SINGLE_CHAR:
		c := tk.Repr[0]
		ranges := []CharacterRange{{Start: c, End: c}}
		sym.CharSet = &CharacterSet{Ranges: ranges}
	case CHAR_CLASS, META_CHAR:
		charSet, _ := NewCharacterSet(*tk)
		sym.CharSet = &charSet
	}
	return sym, nil
}
