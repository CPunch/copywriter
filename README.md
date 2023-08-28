# Copywriter

Copywriter is a very niche program to write complete articles (including images!) for a [Hugo-based](https://gohugo.io/) site. It uses GPT4 to generate the article content and [sdxl](https://replicate.com/stability-ai/sdxl/api) to generate the images.

You can tell copywriter to write a blog about a specific topic by passing a title to the `write` command:
```sh
> ./copywriter write -o . "Why investing in DogeCoin is a great financial decision"
[*] Title: 'Why investing in DogeCoin is a great financial decision'...
[*] Generating image for query 'a happy investor viewing a rising graph representing the increase in value of DogeCoin'...
[*] Using replicate.ai to generate image...
[*] Downloading https://pbxt.replicate.delivery/7ZQdS1TXk1JxOBEtBcwVtXtceoa5u2Mm2RQ9YiAjkVOVlUvIA/out-0.png to 'why-investing-in-dogecoin-is-a-great-financial-decision/file_1.jpg'...
[*] Generating blog post contents...
[*] Populating images...
[*] Generating image for query 'a graphical representation of the DogeCoin's price increase in 2021'...
[*] Using replicate.ai to generate image...
[*] Downloading https://pbxt.replicate.delivery/h29UixSkIk4AMZ2BeKfYdCr8GeqcYZrcqY6ZRaQGjPi4XS9iA/out-0.png to 'why-investing-in-dogecoin-is-a-great-financial-decision/file_2.jpg'...
[*] Generating image for query 'an image showing a weighing scale on which the advantages and risks of investing in DogeCoin are measured'...
[*] Using replicate.ai to generate image...
[*] Downloading https://pbxt.replicate.delivery/AAmnfHfGSKkZ30bYjLZ8x3t019uYe6bfFrYQtZ3ZT1QegJ1LC/out-0.png to 'why-investing-in-dogecoin-is-a-great-financial-decision/file_3.jpg'...
[*] Generating tags...
[SUCCESS] Generated post!
[*] Writing to file 'why-investing-in-dogecoin-is-a-great-financial-decision/index.md'...
[SUCCESS] Done!
```
> Copywriter will create a new slug-like directory in the out path you specified and write all of the content there.

You'll need to populate a few environment variables before running copywriter however, including your OpenAI API Key and [replicate](https://replicate.com) API Key:
```sh
export OPENAI_API_KEY=sk-################################################
export REPLICATE_API_KEY=r8_#####################################
```
> I put this in a file called `.env.local`, and just use `source .env.local` for my local environment

If you omit the `REPLICATE_API_KEY`, image prompts will just be scraped from google with mixed results.

## Usage

```
Usage: copywriter <flags> <subcommand> <subcommand args>

Subcommands:
        commands         list all command names
        flags            describe all known top-level flags
        help             describe subcommands and their syntax
        write            Write a post

Top-level flags (use "copywriter flags" for a full list):
  -config=: copywriter config file
  -custom=: custom prompt
  -image=: image style appended to image prompt
  -trend=all: trending category
  -trend-topic=trends: topic type for trending category
```

As stated before, the `write` command will write an article using a provided title, or if one is omitted a title will be generated based on Google Trend data. For more info about the expected configuration check the `copywriter.ini` example in this repository, any configs in the provided config file (eg. the file passed to `-config`) will overwrite any passed command line arguments, so be careful.

## Compiling

Just cd to the project root and compile it like any other Go program:
```sh
> go build -o copywriter
```