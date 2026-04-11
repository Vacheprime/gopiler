package lexer

// Implement the regex: /abc[0-9]x/

// DFA would look like s1 (a) -> s2(b) -> s3(c) -> s4(digit 0-9) -> s5(x)
// /".*"/
// /(const|var) [a-zA-Z_]+ (string|int) = /
const SEARCH_STR string = " tsabc8sabc7x "
