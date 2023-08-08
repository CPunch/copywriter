package main

import (
	"context"
	"flag"
	"os"

	"github.com/google/subcommands"
)

func main() {
	conf := flag.String("config", "", "copywriter config file")
	subcommands.ImportantFlag("config")
	trnd := flag.String("trend", "all", "trending category")
	subcommands.ImportantFlag("trend")
	cust := flag.String("custom", "", "custom prompt")
	subcommands.ImportantFlag("custom")
	imgs := flag.String("image", "", "image style appended to image prompt")
	subcommands.ImportantFlag("image")
	trndTopic := flag.String("trend-topic", "trends", "topic type for trending category")
	subcommands.ImportantFlag("trend-topic")
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&WriteCommand{}, "")
	flag.Parse()

	cfg := NewConfig(*trnd, *cust, *imgs, *trndTopic)
	if *conf != "" {
		cfg.LoadConfig(*conf)
	}

	ctx := context.WithValue(context.Background(), "conf", cfg)

	os.Exit(int(subcommands.Execute(ctx)))
}
