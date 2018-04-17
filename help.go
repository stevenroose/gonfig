package gonfig

import (
	"fmt"
	"io"
	"os"
	"path"
	"reflect"
	"strings"
)

const (
	defaultHelpDescription = "print this help menu"
	defaultHelpMessage     = "Usage of __EXEC__:"
)

func typeString(t reflect.Type) string {
	if t.Implements(typeOfTextUnmarshaler) {
		return "string"
	}

	if t == typeOfByteSlice {
		return "string"
	}

	switch t.Kind() {
	case reflect.String:
		return "string"

	case reflect.Bool:
		return "bool"

	case reflect.Float32, reflect.Float64:
		return "float"

	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		return "int"

	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		return "uint"

	case reflect.Slice:
		subTypeStr := typeString(t.Elem())
		return subTypeStr + "..."
	}

	return ""
}

// unquoteDescription extracts a back-quoted name from the description
// string and returns it and the un-quoted description.
// Given "a `name` to show" it returns ("name", "a name to show").
// If there are no back quotes, the name is empty.
func unquoteDescription(desc string) (string, string) {
	for i := 0; i < len(desc); i++ {
		if desc[i] == '`' {
			for j := i + 1; j < len(desc); j++ {
				if desc[j] == '`' {
					name := desc[i+1 : j]
					newDesc := desc[:i] + name + desc[j+1:]
					return name, newDesc
				}
			}
			break // only one back quote
		}
	}

	return "", desc
}

// wrapN splits the string `s` on whitespace into an initial substring up
// to `i` runes in length and the remainder. Will go `slop` over `i` if
// that encompasses the entire string (which allows the caller to
// avoid short orphan words on the final line).
func wrapN(i, slop int, s string) (string, string) {
	if i+slop > len(s) {
		return s, ""
	}

	w := strings.LastIndexAny(s[:i], " \t")
	if w <= 0 {
		return s, ""
	}

	return s[:w], s[w+1:]
}

// wrap wraps the string `s` to a maximum width `w` with leading indent
// `i`. The first line is not indented (this is assumed to be done by
// caller). Pass `w` == 0 to do no wrapping
func wrap(i, w int, s string) string {
	if w == 0 {
		return s
	}

	// space between indent i and end of line width w into which
	// we should wrap the text.
	wrap := w - i

	var r, l string

	// Not enough space for sensible wrapping. Wrap as a block on
	// the next line instead.
	if wrap < 24 {
		i = 16
		wrap = w - i
		r += "\n" + strings.Repeat(" ", i)
	}
	// If still not enough space then don't even try to wrap.
	if wrap < 24 {
		return s
	}

	// Try to avoid short orphan words on the final line, by
	// allowing wrapN to go a bit over if that would fit in the
	// remainder of the line.
	slop := 5
	wrap = wrap - slop

	// Handle first line, which is indented by the caller (or the
	// special case above)
	l, s = wrapN(wrap, slop, s)
	r = r + l

	// Now wrap the rest
	for s != "" {
		var t string

		t, s = wrapN(wrap, slop, s)
		r = r + "\n" + strings.Repeat(" ", i) + t
	}

	return r

}

// writeHelpMessage writes the help message to the writer.
//
// This implementation is borrowed from https://github.com/spf13/pflag
func writeHelpMessage(s *setup, w io.Writer) {
	lines := make([]string, 0, len(s.allOpts))

	maxlen := 0
	for _, opt := range s.allOpts {
		if opt.isParent {
			continue
		}
		if opt.hasFieldOpt(fieldOptHidden) {
			continue
		}

		line := ""
		if opt.short != "" {
			line = fmt.Sprintf("  -%v, --%v", opt.short, opt.fullID())
		} else {
			line = fmt.Sprintf("      --%v", opt.fullID())
		}

		typeStr := typeString(opt.value.Type())
		varname, desc := unquoteDescription(opt.desc)
		if opt.isMap {
			line += ".<key> <value>"
		} else {
			if varname == "" {
				varname = typeStr
				if varname == "bool" {
					// We don't want to show a varname for bools.
					varname = ""
				}
			}

			if varname != "" {
				line += " " + varname
			}
		}

		// This special character will be replaced with spacing once the
		// correct alignment is calculated
		line += "\x00"
		if len(line) > maxlen {
			maxlen = len(line)
		}

		line += desc
		if opt.defaul != "" {
			if len(typeStr) >= 6 && typeStr[0:6] == "string" {
				// Put quotes around string types.
				line += fmt.Sprintf(" (default %q)", opt.defaul)
			} else {
				line += fmt.Sprintf(" (default %v)", opt.defaul)
			}
		}

		lines = append(lines, line)
	}

	helpFlagDesc := s.conf.HelpDescription
	if helpFlagDesc == "" {
		helpFlagDesc = defaultHelpDescription
	}
	lines = append(lines, "  -h, --help\x00"+helpFlagDesc)

	message := s.conf.HelpMessage
	if message == "" {
		exec := path.Base(os.Args[0])
		message = strings.Replace(defaultHelpMessage, "__EXEC__", exec, 1)
	}
	fmt.Fprintln(w, message)

	terminalWidth := getTerminalWidth()

	for _, line := range lines {
		sidx := strings.Index(line, "\x00")
		spacing := strings.Repeat(" ", maxlen-sidx)
		// maxlen + 2 comes from + 1 for the \x00 and + 1 for the (deliberate)
		// off-by-one in maxlen-sidx
		fmt.Fprintln(w, line[:sidx], spacing,
			wrap(maxlen+2, terminalWidth, line[sidx+1:]))
	}
}

// printHelpAndExit prints the help message and exits the program.
func printHelpAndExit(s *setup) {
	writeHelpMessage(s, os.Stdout)
	os.Exit(2)
}
