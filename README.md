# Dudu

### What is Dudu?

Dudu is a static site generator. It takes [pandoc](https://pandoc.org/) markdown files and renders them into templated HTML, ready to be served as a static site.

### Why Dudu?

The motivation for Dudu came from realizing there are simply [not enough static site generators](https://staticsitegenerators.net/). In all seriousness, it was inspired by my good friend [Jacob Strieb](https://jstrieb.github.io)'s [similar project](https://github.com/jstrieb/personal-site/) for his personal website. In fact, I borrowed directly from his bash script to replicate the functionality of compiling pandoc markdown in Dudu as `dudu build`. The point of making Dudu was really just because I thought it would be fun to take the work done by Jacob and rebuild it in Go and add new functionality.

Oh, did you mean why the name? No reason, just thought it sounded cool.

### Features

- Hot Reloading
- Incremental Builds
- Project Scaffolding