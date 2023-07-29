package main

import (
	"context"
	"flag"
	"os"

	"github.com/google/subcommands"
)

type Write struct {
	Outfile string
}

func (*Write) Name() string     { return "write" }
func (*Write) Synopsis() string { return "Write a post" }
func (w *Write) SetFlags(f *flag.FlagSet) {
	f.StringVar(&w.Outfile, "o", "out.md", "output file")
}

func (*Write) Usage() string {
	return "write [-o outfile] <title>:\n\tWrite a post. If not title is provided, one will be generated based on previous post titles.\n"
}

func (w *Write) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	var title string
	for _, arg := range f.Args() {
		title += arg + " "
	}

	if title == "" {
		title = genBlogTitle()
	}

	// generate the post
	post := genBlogPost(title)
	Success("Generated post '%s'!", title)

	Info("Writing to file '%s'...", w.Outfile)
	// write to outfile
	if err := os.WriteFile(w.Outfile, []byte(post), 0644); err != nil {
		Fail("Failed to write to file '%s': %v", w.Outfile, err)
	}

	Info("Adding post to DB...")
	addPost(title, post)
	Success("Done!")
	return subcommands.ExitSuccess
}

func main() {
	subcommands.Register(&Write{}, "")

	conf := flag.String("db", "default.db", "SQLite DB file")
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")

	flag.Parse()
	openDB(*conf)
	os.Exit(int(subcommands.Execute(context.Background())))
}
