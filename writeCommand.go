package main

import (
	"context"
	"flag"
	"strings"

	"git.openpunk.com/CPunch/copywriter/util"
	"github.com/google/subcommands"
)

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

	bw.setupOutDir(w.OutDir)

	// write the post
	if err := bw.WritePost(); err != nil {
		util.Fail("Failed to generate post: %v", err)
	}

	util.Success("Done!")
	return subcommands.ExitSuccess
}
