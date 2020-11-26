package xgettext

import (
	"bytes"
	"go/ast"
	"go/parser"
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) {
	TestingT(t)
}

var _ = Suite(xgettextSuite{})

type xgettextSuite struct{}

func (xgettextSuite) TestStringConstant(c *C) {
	for _, test := range []struct {
		code, expected string
	}{
		{`"Hello world"`, "Hello world"},
		{"`Hello world`", "Hello world"},
		{"\"Hello \" + `world`", "Hello world"},
		{`"Line 1\nLine 2"`, "Line 1\nLine 2"},
		{"`Line 1\\nLine 1`", "Line 1\\nLine 1"},
		{`("Hello")`, "Hello"},
		{`("a"+"b")+("c"+"d")`, "abcd"},
	} {
		comment := Commentf("expression: %s", test.code)
		expr, err := parser.ParseExpr(test.code)
		c.Assert(err, IsNil, comment)
		result, err := stringConstant(expr)
		if !c.Check(err, IsNil, comment) {
			continue
		}
		c.Check(result, Equals, test.expected, comment)
	}

	for _, code := range []string{
		"1",
		"'x'",
		"`xyz`+2",
		`("a"+"b")+("c"+42)`,
	} {
		expr, err := parser.ParseExpr(code)
		c.Assert(err, IsNil)
		result, err := stringConstant(expr)
		c.Check(err, NotNil, Commentf("expression %s evaluated to %q", code, result))
	}
}

func (xgettextSuite) TestParseKeyword(c *C) {
	for _, test := range []struct {
		spec string
		kw   Keyword
	}{
		{"Gettext", Keyword{"Gettext", "", 0, -1, -1}},
		{"NGettext:1,2", Keyword{"NGettext", "", 0, 1, -1}},
		{"PGettext:1c,2", Keyword{"PGettext", "", 1, -1, 0}},
		{"PGettext:2,1c", Keyword{"PGettext", "", 1, -1, 0}},
		{"NPGettext:1c,2,3", Keyword{"NPGettext", "", 1, 2, 0}},
		{"NPGettext:2,3,1c", Keyword{"NPGettext", "", 1, 2, 0}},
		{"NPGettext:2,1c,3", Keyword{"NPGettext", "", 1, 2, 0}},
		{"i18n.G", Keyword{"G", "i18n", 0, -1, -1}},
		{"i18n.NG:1,2", Keyword{"NG", "i18n", 0, 1, -1}},
	} {
		comment := Commentf("keyword spec: %s", test.spec)
		kw, err := ParseKeyword(test.spec)
		if !c.Check(err, IsNil, comment) {
			continue
		}
		c.Check(*kw, Equals, test.kw, comment)
	}

	for _, spec := range []string{
		"foo:1,2,3",
		"bar:1c,2,3,4",
		"foo:bar",
		"foo:50x,2",
	} {
		kw, err := ParseKeyword(spec)
		c.Check(err, NotNil, Commentf("spec %s evaluated to %#v", spec, kw))
	}
}

func (xgettextSuite) TestKeywordMatch(c *C) {
	for _, test := range []struct {
		spec string
		code string
		ok   bool
	}{
		{"Gettext", "Gettext()", true},
		{"Gettext", "foo.Gettext()", true},
		{"Gettext", "foo.bar.Gettext()", true},
		{"Gettext", "NotGettext()", false},
		{"i18n.G", "G()", false},
		{"i18n.G", "i18n.G()", true},
		{"i18n.G", "foo.i18n.G()", false},
	} {
		comment := Commentf("spec: %s, expr: %s", test.spec, test.code)
		kw, err := ParseKeyword(test.spec)
		c.Assert(err, IsNil, comment)
		expr, err := parser.ParseExpr(test.code)
		c.Assert(err, IsNil, comment)

		c.Check(kw.Match(expr.(*ast.CallExpr)), Equals, test.ok, comment)
	}
}

func (xgettextSuite) TestKeywordExtract(c *C) {
	for _, test := range []struct {
		spec string
		code string
		ok   bool
		msg  Message
	}{
		{"Gettext", `Gettext("foo\tbar")`, true, Message{msgid: "foo\tbar"}},
		{"Gettext", `Gettext(foo())`, false, Message{}},
		{"NGettext:1,2", `NGettext("foo", "bar", n)`, true, Message{msgid: "foo", msgidPlural: "bar"}},
		{"NGettext:1,2", `NGettext(foo(), "bar", n)`, false, Message{}},
		{"NGettext:1,2", `NGettext("foo", bar(), n)`, false, Message{}},
		{"PGettext:1c,2", `PGettext("foo", "bar")`, true, Message{msgid: "bar", msgContext: "foo"}},
		{"NPGettext:1c,2,3", `NPGettext("foo", "bar", "baz", n)`, true, Message{msgid: "bar", msgidPlural: "baz", msgContext: "foo"}},
		{"NPGettext:1c,2,3", `NPGettext(foo(), "bar", "baz", n)`, false, Message{}},
		{"NPGettext:1c,2,3", `NPGettext("foo", bar(), "baz", n)`, false, Message{}},
		{"NPGettext:1c,2,3", `NPGettext("foo", "bar", baz(), n)`, false, Message{}},

		// out of bounds argument index
		{"Gettext:1", `Gettext()`, false, Message{}},
		{"NGettext:1,2", `NGettext("foo")`, false, Message{}},
		{"PGettext:1,2c", `PGettext("foo")`, false, Message{}},
	} {
		comment := Commentf("spec: %s, expr: %s", test.spec, test.code)
		kw, err := ParseKeyword(test.spec)
		c.Assert(err, IsNil, comment)
		expr, err := parser.ParseExpr(test.code)
		c.Assert(err, IsNil, comment)

		msg, err := kw.Extract(expr.(*ast.CallExpr))
		c.Check(err == nil, Equals, test.ok, comment)
		c.Check(msg, Equals, test.msg, comment)
	}
}

func (xgettextSuite) TestExtractorParseStream(c *C) {
	const fooContent = `package main

func foo() {
	println(Gettext("msg"))
	println(PGettext("context1", "msg"))
	// Not a translator comment
	println(NGettext("single %d", "plural %d", 0))
}
`
	const barContent = `package main

func bar() {
	// TRANS: bar
	println(PGettext("context2", "msg"))
	// TRANSLATORS: xyz
	println(Gettext("msg"))
}
`

	var e Extractor
	e.AddDefaultKeywords()
	e.CommentTags = append(e.CommentTags, "TRANSLATORS:", "TRANS:")
	err := e.parseStream("foo.go", bytes.NewReader([]byte(fooContent)))
	c.Assert(err, IsNil)
	err = e.parseStream("bar.go", bytes.NewReader([]byte(barContent)))
	c.Assert(err, IsNil)

	c.Check(e.Messages, DeepEquals, map[Message][]Location{
		{"msg", "", ""}: {
			{"foo.go", 4, "", false},
			{"bar.go", 7, "#. TRANSLATORS: xyz\n", false},
		},
		{"msg", "", "context1"}: {
			{"foo.go", 5, "", false},
		},
		{"msg", "", "context2"}: {
			{"bar.go", 5, "#. TRANS: bar\n", false},
		},
		{"single %d", "plural %d", ""}: {
			{"foo.go", 7, "", true},
		},
	})
}

func (xgettextSuite) TestExtractorWrite(c *C) {
	e := Extractor{
		SortOutput:       true,
		PackageName:      "testing",
		MsgidBugsAddress: "bugs@example.org",
		CreationDate:     "1970-01-01 TT:TT+00:00",
		Messages: map[Message][]Location{
			{"one line", "", ""}: {
				{"foo.go", 4, "#. comment foo\n", false},
				{"bar.go", 42, "#. comment bar\n", true},
			},
			{"two\nlines", "", ""}: {
				{"file.go", 100, "", false},
			},
			{"single", "plural", ""}: {
				{"file.go", 10, "#. xyz\n", false},
			},
			{"foo", "", "context"}: {
				{"file.go", 50, "", false},
			},
			{"hello\tworld", "", ""}: {
				{"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.go", 10, "", false},
				{"yyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy.go", 20, "", false},
			},
		},
	}

	var buffer bytes.Buffer
	c.Assert(e.Write(&buffer), IsNil)

	const expectedPot = `# SOME DESCRIPTIVE TITLE.
# Copyright (C) YEAR THE PACKAGE'S COPYRIGHT HOLDER
# This file is distributed under the same license as the PACKAGE package.
# FIRST AUTHOR <EMAIL@ADDRESS>, YEAR.
#
#, fuzzy
msgid ""
msgstr ""
"Project-Id-Version: testing\n"
"Report-Msgid-Bugs-To: bugs@example.org\n"
"POT-Creation-Date: 1970-01-01 TT:TT+00:00\n"
"PO-Revision-Date: YEAR-MO-DA HO:MI+ZONE\n"
"Last-Translator: FULL NAME <EMAIL@ADDRESS>\n"
"Language-Team: LANGUAGE <LL@li.org>\n"
"Language: \n"
"MIME-Version: 1.0\n"
"Content-Type: text/plain; charset=CHARSET\n"
"Content-Transfer-Encoding: 8bit\n"

#: file.go:50
msgctxt "context"
msgid "foo"
msgstr ""

#: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.go:10
#: yyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy.go:20
msgid "hello\tworld"
msgstr ""

#. comment bar
#. comment foo
#: bar.go:42 foo.go:4
#, c-format
msgid "one line"
msgstr ""

#. xyz
#: file.go:10
msgid "single"
msgid_plural "plural"
msgstr[0] ""
msgstr[1] ""

#: file.go:100
msgid ""
"two\n"
"lines"
msgstr ""
`
	c.Check(buffer.String(), Equals, expectedPot)
}
