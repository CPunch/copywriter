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
	title = strings.ToLower(title)

	return title + ".md"
}

type Write struct {
	OutDir string
}

func (*Write) Name() string     { return "write" }
func (*Write) Synopsis() string { return "Write a post" }
func (w *Write) SetFlags(f *flag.FlagSet) {
	f.StringVar(&w.OutDir, "o", ".", "output directory")
}

func (*Write) Usage() string {
	return "write [-o outdir] <title>:\n\tWrite a post. If not title is provided, one will be generated based on previous post titles.\n"
}

func (w *Write) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	defer func() {
		if e := recover(); e != nil {
			Fail("%s", e)
		}
	}()

	config := ctx.Value("conf").(*Config)
	bw := NewBlogWriter(config, w.OutDir)

	// build title
	var title string
	for _, arg := range f.Args() {
		title += arg + " "
	}

	// generate the post
	post := bw.WritePost(title)
	os.MkdirAll(w.OutDir, 0600)
	outFile := w.OutDir + "/" + genBlogFileName(bw.Title)

	Info("Writing to file '%s'...", outFile)
	// write to outfile
	if err := os.WriteFile(outFile, []byte(post), 0644); err != nil {
		Fail("Failed to write to file '%s': %v", outFile, err)
	}

	Success("Done!")
	return subcommands.ExitSuccess
}

func main() {
	subcommands.Register(&Write{}, "")

	conf := flag.String("db", "copywriter.ini", "copywriter config file")
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")

	flag.Parse()
	config := LoadConfig(*conf)
	ctx := context.WithValue(context.Background(), "conf", config)
	os.Exit(int(subcommands.Execute(ctx)))
}
