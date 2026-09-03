grammar myLang;

// Lexer rules

SEMI_COLON: ';' ;
LEFT_PAREN : '(' ;
RIGHT_PAREN : ')' ;
LEFT_CURLY : '{' ;
RIGHT_CURLY : '}' ;
IF : 'if' ;
SUM : '+' ;
SUBTRACTION: '-' ;
MULTIPLICATION: '*' ;
DIVISION: '/' ;
FLOAT : [0-9]+[.][0-9]+ ;
INTEGER : [0-9]+ ;
EXPO: '^' ;
ASSIGNMENT : '=' ;
EQUALS : '==' ;
GREATER_THAN : '>' ;
LESS_THAN : '<' ;
GREATER_EQ_THAN : '>=' ;
LESS_EQ_THAN : '<=' ;
COMMA: ',' ;
DTYPE_INT : 'int' ;
DTYPE_FLOAT : 'float' ;
IDENTIFIER : [a-zA-Z][a-zA-Z0-9_]* ;

NEWLINE : '\r'? '\n' -> skip ;
WHITESPACE : [ \t\f]+ -> skip ;

// Parser rules
program : statement* EOF;

statement : 
    var_declaration SEMI_COLON
    | func_call SEMI_COLON
    | if_stmt
    ;

var_declaration : type_declaration IDENTIFIER ASSIGNMENT expr;
if_stmt : IF condition LEFT_CURLY statement* RIGHT_CURLY ;

condition : expr comparison expr ;

// Precedence
expr : 
    factor ((SUM | SUBTRACTION) factor)*
    ;

factor :
    power ((MULTIPLICATION | DIVISION) power)*
    ;

power :
    primary (EXPO power)?
    ;

primary:
    IDENTIFIER
    | value_literal
    | func_call
    | LEFT_PAREN expr RIGHT_PAREN
    ;

arg_list:
    expr* (COMMA expr)*
    ;

func_call : IDENTIFIER LEFT_PAREN arg_list RIGHT_PAREN ;
comparison : EQUALS | GREATER_THAN | GREATER_EQ_THAN | LESS_THAN | LESS_EQ_THAN ;
type_declaration : DTYPE_INT | DTYPE_FLOAT ;
value_literal : INTEGER | FLOAT ;
