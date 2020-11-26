package xgettext

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"
)

var (
	ErrNotString  = errors.New("not a string constant")
	ErrBadKeyword = errors.New("bad keyword")
	ErrOutOfRange = errors.New("argument index out of range")
)

// stringConstant evaluates an ast.Expr representing a string constant
//
// In addition to
func stringConstant(expr ast.Expr) (string, error) {
	switch val := expr.(type) {
	case *ast.BasicLit:
		if val.Kind != token.STRING {
			return "", ErrNotString
		}
		s, err := strconv.Unquote(val.Value)
		if err != nil {
			return "", err
		}
		return s, nil
	// Support simple string concatenation
	case *ast.BinaryExpr:
		// we only support string concat
		if val.Op != token.ADD {
			return "", ErrNotString
		}
		left, err := stringConstant(val.X)
		if err != nil {
			return "", err
		}
		right, err := stringConstant(val.Y)
		if err != nil {
			return "", err
		}
		return left + right, nil
	// Support parenthesised expressions
	case *ast.ParenExpr:
		return stringConstant(val.X)
	}
	return "", ErrNotString
}

type Keyword struct {
	name, pkg                      string
	msgid, msgidPlural, msgContext int
}

func ParseKeyword(spec string) (*Keyword, error) {
	// Keyword spec is of form [PKG.]FUNC[:ARG,...]
	idx := strings.IndexByte(spec, ':')
	var function, pkg string
	var args []string
	if idx >= 0 {
		function = spec[:idx]
		args = strings.Split(spec[idx+1:], ",")
	} else {
		function = spec
	}

	idx = strings.IndexByte(function, '.')
	if idx >= 0 {
		pkg = function[:idx]
		function = function[idx+1:]
		if strings.IndexByte(function, '.') >= 0 {
			return nil, ErrBadKeyword
		}
	}

	k := &Keyword{
		name:        function,
		pkg:         pkg,
		msgid:       0,
		msgidPlural: -1,
		msgContext:  -1,
	}

	// Now process arguments
	processed := 0
	for _, arg := range args {
		if arg[len(arg)-1] == 'c' {
			// This is the context
			val, err := strconv.Atoi(arg[:len(arg)-1])
			if err != nil {
				return nil, err
			}
			k.msgContext = val - 1
			continue
		}

		val, err := strconv.Atoi(arg)
		if err != nil {
			return nil, err
		}
		switch processed {
		case 0:
			k.msgid = val - 1
		case 1:
			k.msgidPlural = val - 1
		default:
			return nil, ErrBadKeyword
		}
		processed += 1
	}

	return k, nil
}

func (k *Keyword) Match(call *ast.CallExpr) bool {
	var pkg, name string

	switch e := call.Fun.(type) {
	case *ast.Ident:
		name = e.Name
	case *ast.SelectorExpr:
		name = e.Sel.Name
		if ident, ok := e.X.(*ast.Ident); ok {
			pkg = ident.Name
		}
	default:
		return false
	}

	if name != k.name {
		return false
	}
	// If the keyword includes a package qualifier, make sure it matches
	return k.pkg == "" || k.pkg == pkg
}

func (k *Keyword) Extract(call *ast.CallExpr) (msg Message, err error) {
	if k.msgid >= len(call.Args) {
		return Message{}, ErrOutOfRange
	}
	msg.msgid, err = stringConstant(call.Args[k.msgid])
	if err != nil {
		return Message{}, err
	}
	if k.msgidPlural >= 0 {
		if k.msgidPlural >= len(call.Args) {
			return Message{}, ErrOutOfRange
		}
		msg.msgidPlural, err = stringConstant(call.Args[k.msgidPlural])
		if err != nil {
			return Message{}, err
		}
	}
	if k.msgContext >= 0 {
		if k.msgContext >= len(call.Args) {
			return Message{}, ErrOutOfRange
		}
		msg.msgContext, err = stringConstant(call.Args[k.msgContext])
		if err != nil {
			return Message{}, err
		}
	}
	return msg, nil
}

type Message struct {
	msgid       string
	msgidPlural string
	msgContext  string
}

func (m *Message) Less(other *Message) bool {
	if m.msgid != other.msgid {
		return m.msgid < other.msgid
	}
	if m.msgidPlural != other.msgidPlural {
		return m.msgidPlural < other.msgidPlural
	}
	return m.msgContext < other.msgContext
}

type Location struct {
	file     string
	line     int
	comments string
	format   bool
}

type visitor struct {
	*Extractor

	fset *token.FileSet
	file *ast.File
}

func commentGroupContent(cg *ast.CommentGroup) string {
	var lines []string
	for _, comment := range cg.List {
		for _, line := range strings.Split(comment.Text, "\n") {
			line = strings.TrimPrefix(line, "//")
			line = strings.TrimPrefix(line, "/*")
			line = strings.TrimSuffix(line, "*/")
			line = strings.TrimSpace(line)
			if line != "" {
				lines = append(lines, "#. "+line+"\n")
			}
		}
	}
	return strings.Join(lines, "")
}

func (v *visitor) findCommentsBefore(pos token.Position) string {
	for i := len(v.file.Comments) - 1; i >= 0; i-- {
		cg := v.file.Comments[i]
		cgPos := v.fset.Position(cg.End())
		if cgPos.Line+1 == pos.Line {
			return commentGroupContent(cg)
		}
	}
	return ""
}

func (v *visitor) Visit(node ast.Node) ast.Visitor {
	// We're only interested in calls
	call, ok := node.(*ast.CallExpr)
	if !ok {
		return v
	}

	for _, k := range v.Keywords {
		if !k.Match(call) {
			continue
		}

		msg, err := k.Extract(call)
		if err != nil {
			break
		}

		pos := v.fset.Position(node.Pos())
		var comments string
		if len(v.CommentTags) != 0 {
			comments = v.findCommentsBefore(pos)
			keep := false
			for _, tag := range v.CommentTags {
				if strings.HasPrefix(comments, "#. "+tag) {
					keep = true
					break
				}
			}
			if !keep {
				comments = ""
			}
		}

		v.Messages[msg] = append(v.Messages[msg], Location{
			file:     pos.Filename,
			line:     pos.Line,
			comments: comments,
			// FIXME: too simplistic, should check if call
			// used as a format argument.
			format:   strings.IndexByte(msg.msgid, '%') >= 0,
		})
		break
	}
	return v
}

type Extractor struct {
	Messages    map[Message][]Location
	Keywords    []*Keyword
	CommentTags []string
	Directories []string
	SortOutput  bool
	NoLocation  bool

	PackageName      string
	MsgidBugsAddress string
	CreationDate     string
}

func (e *Extractor) AddDefaultKeywords() {
	for _, spec := range []string{
		"Gettext:1",
		"NGettext:1,2",
		"PGettext:1c,2",
		"NPGettext:1c,2,3",
	} {
		kw, err := ParseKeyword(spec)
		if err != nil {
			panic(err)
		}
		e.Keywords = append(e.Keywords, kw)
	}
}

func (e *Extractor) openFile(filename string) (f *os.File, err error) {
	if len(e.Directories) == 0 || filepath.IsAbs(filename) {
		return os.Open(filename)
	}
	for _, dir := range e.Directories {
		f, err = os.Open(filepath.Join(dir, filename))
		if !os.IsNotExist(err) {
			break
		}
	}
	return f, err
}

func (e *Extractor) parseStream(filename string, r io.Reader) (err error) {
	var v visitor
	v.Extractor = e
	v.fset = token.NewFileSet()
	v.file, err = parser.ParseFile(v.fset, filename, r, parser.ParseComments)
	if err != nil {
		return err
	}

	if e.Messages == nil {
		e.Messages = make(map[Message][]Location)
	}
	ast.Walk(&v, v.file)
	return nil
}

func (e *Extractor) ParseFile(filename string) error {
	f, err := e.openFile(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return e.parseStream(filename, f)
}

const poTemplateData = `# SOME DESCRIPTIVE TITLE.
# Copyright (C) YEAR THE PACKAGE'S COPYRIGHT HOLDER
# This file is distributed under the same license as the PACKAGE package.
# FIRST AUTHOR <EMAIL@ADDRESS>, YEAR.
#
#, fuzzy
msgid ""
msgstr ""
"Project-Id-Version: {{ or .Extractor.PackageName "PACKAGE" }}\n"
{{ if .Extractor.MsgidBugsAddress -}}
"Report-Msgid-Bugs-To: {{ .Extractor.MsgidBugsAddress }}\n"
{{ end -}}
"POT-Creation-Date: {{ .Extractor.CreationDate }}\n"
"PO-Revision-Date: YEAR-MO-DA HO:MI+ZONE\n"
"Last-Translator: FULL NAME <EMAIL@ADDRESS>\n"
"Language-Team: LANGUAGE <LL@li.org>\n"
"Language: \n"
"MIME-Version: 1.0\n"
"Content-Type: text/plain; charset=CHARSET\n"
"Content-Transfer-Encoding: 8bit\n"
{{ range .Messages -}}
{{ "\n" -}}
{{ .Comments -}}
{{ .Positions -}}
{{ if .Format -}}
#, c-format
{{ end -}}
{{ if .MsgContext -}}
msgctxt {{ .MsgContext }}
{{ end -}}
msgid {{ .Msgid }}
{{ if .MsgidPlural -}}
msgid_plural {{ .MsgidPlural }}
{{ end -}}
{{ if .MsgidPlural -}}
msgstr[0] ""
msgstr[1] ""
{{ else -}}
msgstr ""
{{ end -}}
{{end -}}
`

var poTemplate = template.Must(template.New("po").Parse(poTemplateData))

type messageData struct {
	msg         Message
	Msgid       string
	MsgidPlural string
	MsgContext  string

	Positions string
	Comments  string
	Format    bool
}

func quoteMsgid(msg string) string {
	if len(msg) == 0 {
		return ""
	}

	quoted := []string{`""`}
	for _, line := range strings.SplitAfter(msg, "\n") {
		if len(line) == 0 {
			continue
		}
		quoted = append(quoted, strconv.Quote(line))
	}

	if len(quoted) == 2 {
		return quoted[1]
	}
	return strings.Join(quoted, "\n")
}

func (e *Extractor) Write(w io.Writer) error {
	msgData := make([]*messageData, 0, len(e.Messages))
	for msg, locs := range e.Messages {
		data := &messageData{msg: msg}
		msgData = append(msgData, data)
		data.Msgid = quoteMsgid(msg.msgid)
		data.MsgidPlural = quoteMsgid(msg.msgidPlural)
		data.MsgContext = quoteMsgid(msg.msgContext)
		if e.SortOutput {
			sort.Slice(locs, func(i, j int) bool {
				return locs[i].file < locs[j].file || locs[i].file == locs[j].file && locs[i].line < locs[j].line
			})
		}
		var positions string
		for _, loc := range locs {
			data.Comments += loc.comments
			if loc.format {
				data.Format = true
			}

			if e.NoLocation {
				continue
			}
			pos := fmt.Sprintf("%s:%d", loc.file, loc.line)
			if len(positions)+len(pos) > 75 {
				data.Positions += "#:" + positions + "\n"
				positions = ""
			}
			positions += " " + pos
		}
		if len(positions) > 0 {
			data.Positions += "#:" + positions + "\n"
		}
	}

	if e.SortOutput {
		sort.Slice(msgData, func(i, j int) bool {
			return msgData[i].msg.Less(&msgData[j].msg)
		})
	}

	return poTemplate.Execute(w, struct {
		Extractor *Extractor
		Messages  []*messageData
	}{
		Extractor: e,
		Messages:  msgData,
	})
	return nil
}
