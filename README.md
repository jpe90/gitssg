# Notice - abandoned

Go templating is way too slow for this use case, so I don't think the project is going to pan out. Making a series of writes while parsing the git repository similar to Stagit's approach would drastically improve performance, but I'd really rather support some form of HTML templating.

I might make another attempt with a different language that has a good implementation of performant, pre-compiled templates as well as a library capability syntax highlighting with an HTML backend, or I may work on those subproblems individually if I don't find them.

# gitssg

gitssg is a static site generator for git repositories. It is in an exploratory stage and is not ready for use.

# TODO

- add html templates and process for files and commits
- achieve acceptable performance
- parse README markdown files and render as HTML
- add syntax highlighting to source files

## Contributing

Contributions are welcome! Feel free to send a pull request on [Github](https://github.com/jpe90/gitssg)
or email a patch directly to eskinjp at gmail dot com.
