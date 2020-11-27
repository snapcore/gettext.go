package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"

	"github.com/snapcore/go-gettext/internal/xgettext"
)

var formatTime = func() string {
	return time.Now().Format("2006-01-02 15:04-0700")
}

type options struct {
	FilesFrom string `short:"f" long:"files-from" value-name:"FILE" description:"get list of input files from FILE"`

	Directories []string `short:"D" long:"directory" value-name:"DIRECTORY" description:"add DIRECTORY to list for input files search"`

	Output string `short:"o" long:"output" value-name:"FILE" description:"output to specified file"`

	CommentTags []string `short:"c" long:"add-comments" optional:"true" optional-value:"" value-name:"TAG" description:"place all comment blocks preceding keyword lines in output file"`

	Keywords []string `short:"k" long:"keyword" optional:"true" optional-value:"" value-name:"WORD" description:"look for WORD as the keyword for singular strings"`

	NoLocation bool `long:"no-location" description:"do not write '#: filename:line' lines"`

	SortOutput bool `short:"s" long:"sort-output" description:"generate sorted output"`

	PackageName string `long:"package-name" value-name:"PACKAGE" description:"set package name in output"`

	MsgidBugsAddress string `long:"msgid-bugs-address" default:"EMAIL" value-name:"ADDRESS" description:"set report address for msgid bugs"`
}

func main() {
	// parse args
	var opts options
	args, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		os.Exit(1)
	}

	var files []string
	if opts.FilesFrom != "" {
		content, err := ioutil.ReadFile(opts.FilesFrom)
		if err != nil {
			log.Fatalf("cannot read file %v: %v", opts.FilesFrom, err)
		}
		content = bytes.TrimSpace(content)
		files = strings.Split(string(content), "\n")
	} else {
		files = args[1:]
	}

	extractor := xgettext.Extractor{
		Directories:      opts.Directories,
		CommentTags:      opts.CommentTags,
		SortOutput:       opts.SortOutput,
		NoLocation:       opts.NoLocation,
		PackageName:      opts.PackageName,
		MsgidBugsAddress: opts.MsgidBugsAddress,
		CreationDate:     formatTime(),
	}
	log.Printf("keywords: %#v", opts.Keywords)
	addDefaultKeywords := true
	for _, spec := range opts.Keywords {
		if spec == "" {
			// a bare "-k" option disables the default keywords
			addDefaultKeywords = false
			continue
		}
		kw, err := xgettext.ParseKeyword(spec)
		if err != nil {
			log.Fatalf("cannot parse keyword %s: %s", spec, err)
		}
		extractor.Keywords = append(extractor.Keywords, kw)
	}
	if addDefaultKeywords {
		extractor.AddDefaultKeywords()
	}

	for _, filename := range files {
		if err := extractor.ParseFile(filename); err != nil {
			log.Fatalf("cannot parse file %s: %s", filename, err)
		}
	}

	out := os.Stdout
	if opts.Output != "" {
		var err error
		out, err = os.Create(opts.Output)
		if err != nil {
			log.Fatalf("failed to create %s: %s", opts.Output, err)
		}
	}
	if err := extractor.Write(out); err != nil {
		log.Fatalf("failed to write po template: %s", err)
	}
}
