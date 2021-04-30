# Dudu

## What is Dudu?

Dudu is a static site generator. It takes [pandoc](https://pandoc.org/) markdown files and renders them into templated HTML, ready to be served as a static site.

## Why Dudu?

The motivation for Dudu came from realizing there are simply [not enough static site generators](https://staticsitegenerators.net/). In all seriousness, it was inspired by my good friend [Jacob Strieb](https://jstrieb.github.io)'s [similar project](https://github.com/jstrieb/personal-site/) for his personal website. In fact, I borrowed directly from his bash script to replicate the functionality of compiling pandoc markdown in Dudu as `dudu build`. The point of making Dudu was really just because I thought it would be fun to take the work done by Jacob and rebuild it in Go and add new functionality.

Oh, did you mean why the name? No reason, just thought it sounded cool.

## Features

- Hot Reloading
- Incremental Builds
- Project Scaffolding

## Getting Started

First, consider why you are using this obscure and opinionated static site generator when so many objectively better alternatives exist.

This package uses the `embed` package, which is only available in Go version 1.16 and higher.
### Installation (using `go install`)
- `go install github.com/lsnow99/dudu/cmd/dudu@latest`

### Installation from source
- `git clone https://github.com/lsnow99/dudu`
- `go build -o dudu cmd/dudu`
- `mv dudu $GOBIN`

(Assumes `$GOBIN` is set and in your `$PATH`)

You can now run `dudu new` to create a new project

Within the new project folder, some default files will be created for you:
```bash
.
├── md                          # Here is where your markdown content lives
│   ├── index.html              # Landing page. Does not hot reload
│   ├── 404.html                # Standard 404 page, redirects to /
│   ├── style.css               # Global style file
└── resources                   # This folder contains files for templating
    ├── code-highlight.theme    # Theming file for code markdown
    ├── footer.html             # Footer to be injected at the bottom of each content page
    ├── hotreload.html          # Script injected when running dudu serve
    ├── navbar.html             # Navbar to be injected at the top of each content page
    └── template.html           # Main template file for all content pages
```

A common pattern is to have subfolders within `md` to organize your content, and within these subfolders add your posts/pages as `.md` files.

## Usage
- `dudu new` - Scaffolds a new Dudu project
- `dudu serve` - Start the hot-reloading webserver, available at http://localhost:8080
- `dudu build` - Generates the static site, by default outputting to `static/`
- `dudu help` - Show all command options
