package main

import (
	"context"
	"flag"
	"os"
	"strings"
	"unicode"

	"github.com/google/subcommands"
)

func genBlogFileName(title string) string {
	// strip any non-alphanumeric characters
	title = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsSpace(r) {
			return r
		}
		return -1
	}, title)

	title = strings.TrimSpace(title)
	title = strings.ReplaceAll(title, " ", "-")

	return title + ".md"
}

type Write struct {
	Outdir string
}

func (*Write) Name() string     { return "write" }
func (*Write) Synopsis() string { return "Write a post" }
func (w *Write) SetFlags(f *flag.FlagSet) {
	f.StringVar(&w.Outdir, "o", "content/post", "output file")
}

func (*Write) Usage() string {
	return "write [-o outdir] <title>:\n\tWrite a post. If not title is provided, one will be generated based on previous post titles.\n"
}

func (w *Write) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	var title string
	for _, arg := range f.Args() {
		title += arg + " "
	}

	if title == "" {
		title = GenBlogTitle()
	}

	Info("Generating post '%s'...", title)

	// generate the post
	outFile := w.Outdir + "/" + genBlogFileName(title) + ".md"
	post := GenBlogPost(title)
	Success("Generated post '%s'!", title)

	Info("Writing to file '%s'...", outFile)
	// write to outfile
	if err := os.WriteFile(outFile, []byte(post), 0644); err != nil {
		Fail("Failed to write to file '%s': %v", outFile, err)
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
