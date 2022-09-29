#todo
- rethink error handling
  - bail method in util
	  takes a message, writes it to stderr and bails
- add more templates
  - refs
	- submodules
	- lisence
- fix formatting of file mode, size
- add cmdline overrides for templates
  - header, footer, files, refs, log
- implement caching for diff crunching

it's very rough and needs refinement

like given an iterator and a transformation function, call the transformation fn on each or something

we don't care about cache, so we can just skip the log file header and write it straight up

# stagit
- manually parse args
- open the repodir
- get the head
- check for README, submodiles, lisence
## log html
- make "log.html" file
- make a commit folder w/ 777 perm
- write common header for log file (goes in a fn)
- start writing the table

# LATER
- open a read cache and a write cache

# stagit approach
## stagit-index
- takes a list of repos and builds html results in that order
- 

stagit-index [repodir]

# release checklist
- documentation
- error handling
