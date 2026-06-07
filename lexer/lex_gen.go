package lexer

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strings"
	"unicode"
)

const (
	ErrInvalidClassDefinition   string = "Invalid syntax in token class or fragment definition."
	ErrClassRedefined           string = "Class fragment or definition was redefined."
	ErrInvalidRePlaceholder     string = "Invalid syntax for the regex placeholder."
	ErrInvalidRePlaceholderName string = "Invalid identifier for regex placeholder."
	ErrClassDefNotDefined       string = "Class definition not defined."
	ErrUnclosedRePlaceholder    string = "Regex placeholder is never closed."
)

type ParseError struct {
	Message string
	Line    int
}

func (pe *ParseError) Error() string {
	return fmt.Sprintf("%s Line: %d", pe.Message, pe.Line)
}

type definitionType int

const (
	CLASS definitionType = iota
	FRAG
)

type definition struct {
	identifier string
	regex      string
	defType    definitionType
}

func ParseDefinitions(patternFilePath string) ([]definition, error) {
	file, err := os.Open(patternFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	classDefs := []definition{}
	lineCount := 1
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		def, err := parseDefinition(line)
		if err != nil {
			err.Line = lineCount
			return nil, err
		}
		found := slices.ContainsFunc(classDefs, func(d definition) bool {
			if d.identifier == def.identifier {
				return true
			}
			return false
		})
		if found {
			return nil, &ParseError{ErrClassRedefined, lineCount}
		}
		substituted, err := substituteRegexPlaceholders(def.regex, classDefs)
		if err != nil {
			err.Line = lineCount
			return nil, err
		}
		def.regex = substituted
		lineCount++
	}
	return filterDefinitionsByType(classDefs, CLASS), nil
}

func filterDefinitionsByType(defs []definition, defType definitionType) []definition {
	filtered := make([]definition, 0, len(defs))
	for _, d := range defs {
		if d.defType == defType {
			filtered = append(filtered, d)
		}
	}
	return filtered
}

func parseDefinition(line string) (definition, *ParseError) {
	components := strings.Fields(line)
	if len(components) != 2 {
		return definition{}, &ParseError{Message: ErrInvalidClassDefinition}
	}
	className := components[0]
	regexDef := components[1]
	defType := CLASS
	if strings.HasPrefix(className, "!") {
		defType = FRAG
	}
	return definition{className, regexDef, defType}, nil
}

func substituteRegexPlaceholders(regex string, classDefs []definition) (string, *ParseError) {
	chars := []rune(regex)
	var substitutedRegexBuilder strings.Builder
	for i := 0; i < len(chars); i++ {
		c := chars[i]
		if c == '\\' {
			if i+1 < len(chars) && chars[i+1] != '{' {
				substitutedRegexBuilder.WriteRune('\\')
				substitutedRegexBuilder.WriteRune(chars[i+1])
			} else {
				substitutedRegexBuilder.WriteRune(chars[i+1])
			}
			i++ // Next char is escaped
		} else if c == '{' {
			if i+1 >= len(chars) {
				return "", &ParseError{Message: ErrInvalidRePlaceholder}
			}
			placeholderId, n, err := parseRegexPlaceholder(chars[i+1:])
			if err != nil {
				return "", err
			}
			foundIdx := slices.IndexFunc(classDefs, func(d definition) bool {
				return d.identifier == placeholderId
			})
			if foundIdx == -1 {
				return "", &ParseError{Message: ErrClassDefNotDefined}
			}
			substituteRegex := classDefs[foundIdx].regex
			substitutedRegexBuilder.WriteString(substituteRegex)
			i += n
		} else {
			substitutedRegexBuilder.WriteRune(c)
		}

	}
	return substitutedRegexBuilder.String(), nil
}

func parseRegexPlaceholder(nextChars []rune) (string, int, *ParseError) {
	var identifierBuilder strings.Builder
	charCount := 0
	wasClosed := false
	for _, c := range nextChars {
		charCount++
		if c == '}' {
			wasClosed = true
			break
		}
		identifierBuilder.WriteRune(c)
	}
	if !wasClosed {
		return "", 0, &ParseError{Message: ErrUnclosedRePlaceholder}
	}
	identifier := strings.TrimSpace(identifierBuilder.String())
	if strings.IndexFunc(identifier, isWhitespace) != -1 {
		return "", 0, &ParseError{Message: ErrInvalidRePlaceholderName}
	}
	return identifier, charCount, nil
}

/* Checks if a rune is a whitespace excluding newlines */
func isWhitespace(r rune) bool {
	return unicode.IsSpace(r) && r != '\n'
}
