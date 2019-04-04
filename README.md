# library-price

This is a quick tool that takes a `csv` file for input (via `os.Stdin`) and fetches prices using [isbn.nu](https://isbn.nu) 
that it then adds as additional columns to a `csv` output (written to `os.Stdout`). 

## Be Nice

[isbn.nu](https://isbn.nu) specifically states that their website is meant for human/manual lookups. The tool tries to be nice
by throttling and making requests very slowly. It's intended to be used along with a personal or small library. Let's say you wanted
to find out how much your book collection is worth for future insurance claims. [isbn.nu](https://isbn.nu) is well within their 
rights to block access. That said, I hope that no one behind [isbn.nu](https://isbn.nu) is going to be mad with this limited use 🙇‍♂️.
I'm really thankful for this site to exist as alternatives aren't straightforward or ideal for normal individuals.

## Requirements

This uses [Go 1.12](https://golang.org) or higher. 

## Install 

`go install github.com/alexandre-normand/library-pricer`

## Usage

You'll need a `csv` that has a `isbn` column. Find the column number holding the `isbn` and take note as you'll need
to specify it to run `library-pricer`. To get the full usage, just do:

```
library-pricer --help
usage: library-pricer --isbnColumn=ISBNCOLUMN [<flags>]

Flags:
      --help                   Show context-sensitive help (also try --help-long and --help-man).
      --isbnColumn=ISBNCOLUMN  The column containing the isbn to use when looking up prices (first
                               column starting at 1)
      --withHeaderRow          Indicates if a header row should be read from the input and written in
                               the output
  -t, --throttleTime=15s       Sleep time between calls to isbn.nu (be easy on them).
      --version                Show application version.

```