package main

import (
	"context"
	"flag"
	"os"
	"path"
	"strings"
	"unicode"

	"git.openpunk.com/CPunch/copywriter/util"
	"github.com/google/subcommands"
)

func genBlogFilePath(title string) string {
	// strip any non-alphanumeric characters
	title = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsSpace(r) {
			return r
		}
		return -1
	}, title)

	title = strings.TrimSpace(title)
	title = strings.ReplaceAll(title, " ", "-")
	title = strings.ToLower(title)

	return title
}

type WriteCommand struct {
	OutDir string
}

func (*WriteCommand) Name() string     { return "write" }
func (*WriteCommand) Synopsis() string { return "Write a post" }
func (w *WriteCommand) SetFlags(f *flag.FlagSet) {
	f.StringVar(&w.OutDir, "o", ".", "output directory")
}

func (*WriteCommand) Usage() string {
	return "write [-o outdir] <title>:\n\tWrite a post. If title is not provided, one will be generated based on previous post titles.\n"
}

func (w *WriteCommand) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	config := ctx.Value("conf").(*ConfigData)

	// build title
	var title string
	for _, arg := range f.Args() {
		title += arg + " "
	}

	title = strings.TrimSpace(title)

	// create the blog writer, set the title and output directory
	bw := NewBlogWriter(config)
	if err := bw.setTitle(title); err != nil {
		util.Fail("Failed to set title: %v", err)
	}

	dirPath := path.Join(w.OutDir, genBlogFilePath(bw.Title))
	if err := os.MkdirAll(dirPath, 0777); err != nil {
		util.Fail("Failed to create directory '%s': %v", dirPath, err)
	}
	bw.setOutDir(dirPath)

	// generate the post
	post, err := bw.WritePost()
	if err != nil {
		util.Fail("Failed to generate post: %v", err)
	}
	outFile := path.Join(dirPath, "index.md")

	util.Info("Writing to file '%s'...", outFile)
	if err := os.WriteFile(outFile, []byte(post), 0644); err != nil {
		util.Fail("Failed to write to file '%s': %v", outFile, err)
	}

	util.Success("Done!")
	return subcommands.ExitSuccess
}
