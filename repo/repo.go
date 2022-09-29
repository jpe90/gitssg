package repo

import (
	"html/template"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"jeskin.net/gitssg/templates"

	"github.com/go-git/go-git/v5"
)

var (
	readmeFiles  = []string{"README", "readme", "README.md", "readme.md"}
	lisenceFiles = []string{"LICENSE", "LICENSE.md", "COPYING"}
)

const (
	headerTpl = `{{define "header"}}<!DOCTYPE html>
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
<meta name="viewport" content="width=device-width, initial-scale=1" />
<title>{{.Title}}{{if (and .Title .StrippedName)}} - {{end}}{{.StrippedName}}{{with .Description}} - {{.}}{{end}}</title>
<link rel="icon" type="image/png" href="{{.RelPath}}favicon.png" />
<link rel="alternate" type="application/atom+xml" title="{{.Name}} Atom Feed" href="atom.xml" />
<link rel="alternate" type="application/atom+xml" title="{{.Name}} Atom Feed (tags)" href="tags.xml" />
<link rel="stylesheet" type="text/css" href="{{.RelPath}}style.css" />
</head>
<body>
<table><tr><td><a href="../{{.RelPath}}"><img src="{{.RelPath}}logo.png" alt="" width="32" height="32" /></a></td><td><h1>{{.StrippedName}}</h1><span class="desc">{{.Description}}
</span></td></tr>{{with .URL}}<tr class="url"><td></td><td>git clone <a href="{{.}}">{{.}}</a></td></tr>{{end}}<tr><td></td><td>
{{with .Readme }}<a href="file/{{.}}html">README</a> |{{end}} <a href="log.html">Log</a> | <a href="files.html">Files</a> | <a href="refs.html">Refs</a> |  <a href="file/LICENSE.html">LICENSE</a></td></tr></table>
<hr/>
<div id="content">
{{end}}`
	logTpl = `{{block "header" .}}{{end}}<table id="log"><thead>
<tr><td><b>Date</b></td>"
<td><b>Commit message</b></td>
<td><b>Author</b></td><td class="num" align="right"><b>Files</b></td>
<td class="num" align="right"><b>+</b></td>
<td class="num" align="right"><b>-</b></td></tr>
</thead><tbody>
<tr>{{range .Commits}}
<td>{{FormatTime .Author.When}}</td>
<td>{{TrimMessage .Message}}</td>
<td>{{.Author.Name }}</td>
<td>{{NumFiles .}}</td>
<td>+{{Added .}}</td>
<td>-{{Deleted .}}</td>
</tr></div></body></html>{{end}}
`
	filesTpl = `{{block "header" .}}{{end}}<table id="files"><thead>
<tr>
<td><b>Mode</b></td><td><b>Name</b></td>
<td class="num" align="right"><b>Size</b></td>
</tr>
</thead><tbody>
<tr>{{range .Files}}
<td>{{.Mode.ToOSFileMode}}</td>
<td>{{.Name}}</td>
<td>{{Size .}}</td>
</tr></div></body></html>{{end}}`
	refsTpl = `{{block "header" .}}{{end}}<h2>Branches</h2><table id="branches"><thead>
<tr><td><b>Name</b></td><td><b>Last commit date</b></td><td><b>Author</b></td>
</tr>
</thead><tbody>
<tr>{{range .Branches }}
<td>{{RefName .Ref.Name }}</td>
<td>{{FormatTime .LastCmt.Author.When }}</td>
<td>{{.LastCmt.Author.Name }}</td>
</tr>{{end}}
</tbody></table><br/>
<h2>Tags</h2><table id="tags"><thead>
<tr><td><b>Name</b></td><td><b>Last commit date</b></td><td><b>Author</b></td>
</tr>
</thead><tbody>
<tr>{{range .Tags}}
<td>{{RefName .Ref.Name }}</td>
<td>{{FormatTime .LastCmt.Author.When }}</td>
<td>{{.LastCmt.Author.Name }}</td>
</tr></div></body></html>{{end}}`
)

type headerData struct {
	Name, RelPath, URL, StrippedName, Description, Readme, License, Submodules, Title string
}

type logData struct {
	*headerData
	Commits []*object.Commit
}

type filesData struct {
	*headerData
	Files []*object.File
}

type refsData struct {
	*headerData
	Branches []refData
	Tags     []refData
}

type refData struct {
	Ref     *plumbing.Reference
	LastCmt *object.Commit
}

func formatRefName(rn plumbing.ReferenceName) string {
	s := rn.String()
	lastInd := strings.LastIndex(s, "/") + 1
	return s[lastInd:]
}

func numFilesFromCommit(c *object.Commit) string {
	fs, err := c.Stats()
	if err != nil {
		log.Println("Error getting filestats: ", err)
		return ""
	}
	numFiles := strconv.Itoa(len(fs))
	return numFiles
}

func additionsFromCommit(c *object.Commit) string {
	fs, err := c.Stats()
	if err != nil {
		log.Println("Error getting filestats: ", err)
		return ""
	}
	var totalAdded int
	for _, filestat := range fs {
		totalAdded = totalAdded + filestat.Addition
	}
	return strconv.Itoa(totalAdded)
	// return fs.String()
}

func deletionsFromCommit(c *object.Commit) string {
	fs, err := c.Stats()
	if err != nil {
		log.Println("Error getting filestats: ", err)
		return ""
	}
	var totalDeleted int
	for _, filestat := range fs {
		totalDeleted = totalDeleted + filestat.Deletion
	}
	return strconv.Itoa(totalDeleted)
}

func sizeStr(f *object.File) string {
	return strconv.Itoa(int(f.Blob.Size))
}

func Run(repodir string) {

	var readme, lisence, submodules string

	abs, err := filepath.Abs(repodir)

	if err != nil {
		log.Fatalln(repodir, " is an invalid file path: ", err)
	}

	filename := filepath.Base(abs)
	filenameStripped := strings.TrimSuffix(filename, filepath.Ext(filename))

	r, err := git.PlainOpen(abs)
	if err != nil {
		log.Fatalln(repodir, " is not a git repository: ", repodir, err)
	}

	descrData, _ := os.ReadFile(path.Join(repodir, "/description"))
	description := string(descrData)
	urlData, _ := os.ReadFile(path.Join(repodir, "/url"))
	url := string(urlData)

	head, err := r.Head()
	if err != nil {
		log.Fatalln("could not find HEAD: ", err)
	}

	commit, err := r.CommitObject(head.Hash())
	if err != nil {
		log.Fatalln("could not hash ref for HEAD: ", err)
	}

	for _, file := range readmeFiles {
		found, _ := commit.File(file)
		if found != nil {
			readme = file
		}
	}

	for _, file := range lisenceFiles {
		found, _ := commit.File(file)
		if found != nil {
			lisence = file
		}
	}

	submodulesFile, _ := commit.File(".gitmodules")
	if submodulesFile != nil {
		submodules = ".gitmodules"
	}

	commits := make([]*object.Commit, 0)

	cIter, err := r.Log(&git.LogOptions{From: head.Hash()})
	if err != nil {
		log.Fatalln("could not obtain commit log: ", err)
	}

	err = cIter.ForEach(func(c *object.Commit) error {
		commits = append(commits, c)
		return nil
	})

	logPageData := logData{
		headerData: &headerData{
			Name:         filename,
			StrippedName: filenameStripped,
			Description:  description,
			RelPath:      "",
			Readme:       readme,
			License:      lisence,
			Submodules:   submodules,
			Title:        "Log",
			URL:          url,
		},
		Commits: commits,
	}

	logPageTmpl, err := template.New("logPageTmpl").Funcs(template.FuncMap{
		"TrimMessage": func(c string) string {
			return strings.TrimSuffix(c, "\n")
		},
		"FormatTime": templates.FormatTime,
		"NumFiles": func(c *object.Commit) string {
			fs, err := c.Stats()
			if err != nil {
				log.Println("couldn't get commit stats: ", err)
				return ""
			}
			numFiles := strconv.Itoa(len(fs))
			return numFiles
		},
		"Added":   additionsFromCommit,
		"Deleted": deletionsFromCommit,
	}).Parse(logTpl)

	if err != nil {
		log.Println("unable to load log page template: ", err)
	}

	headerTmpl, err := template.Must(logPageTmpl.Clone()).Parse(headerTpl)
	if err != nil {
		log.Println("unable to load header for log page template: ", err)
	}

	out, err := os.Create("log.html")
	if err != nil {
		log.Println("unable to create log page file: ", err)
	}

	if err := headerTmpl.Execute(out, logPageData); err != nil {
		log.Println("unable to fill log page template: ", err)
	}
	err = out.Close()
	if err != nil {
		log.Println("unable to close log page file: ", err)
	}

	tree, err := commit.Tree()
	if err != nil {
		log.Println("couldn't get tree for head commit: ", err)
	}

	files := make([]*object.File, 0)
	tree.Files().ForEach(func(f *object.File) error {
		files = append(files, f)
		return nil
	})

	filesPageData := filesData{
		headerData: &headerData{
			Name:         filename,
			StrippedName: filenameStripped,
			Description:  description,
			RelPath:      "",
			Readme:       readme,
			License:      lisence,
			Submodules:   submodules,
			Title:        "Files",
			URL:          url,
		},
		Files: files,
	}

	filesPageTmpl, err := template.New("file").Funcs(template.FuncMap{
		"Size": sizeStr,
	}).Parse(filesTpl)

	if err != nil {
		log.Println("unable to load files page template: ", err)
	}

	headerTmpl, err = template.Must(filesPageTmpl.Clone()).Parse(headerTpl)

	if err != nil {
		log.Println("unable to load header for files page template: ", err)
	}

	out, err = os.Create("files.html")
	if err != nil {
		log.Println("unable to create files page file: ", err)
	}

	if err := headerTmpl.Execute(out, filesPageData); err != nil {
		log.Println("unable to fill files page template: ", err)
	}
	err = out.Close()
	if err != nil {
		log.Println("unable to close files page file: ", err)
	}

	tags := make([]refData, 0)
	branches := make([]refData, 0)

	refs, err := r.References()
	if err != nil {
		log.Fatalln("failed to get refs: ", err)
	}

	err = refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().IsBranch() {
			branchIter, err := r.Log(&git.LogOptions{From: ref.Hash(), Order: git.LogOrderCommitterTime})
			if err != nil {
				log.Println("failed to get branch log for ", ref.Hash(), ": ", err)
				return nil
			}
			lastCmt, err := branchIter.Next()
			if err != nil {
				log.Println("failed to get latest branch commit for ", ref.Hash(), ": ", err)
				return nil
			}
			branches = append(branches, refData{Ref: ref, LastCmt: lastCmt})
		} else if ref.Name().IsTag() {
			// couldn't find a single approach to grabbing commit info for annotated
			// and lightweight tags, so check for each and handle separately
			obj, err := r.TagObject(ref.Hash())
			// absence of error indicates a annotated tag
			if err == nil {
				lastCmt, err := obj.Commit()
				if err != nil {
					log.Println("failed to get commit for annotated tag", ref.Hash(), ": ", err)
				}
				tags = append(tags, refData{Ref: ref, LastCmt: lastCmt})
				return nil
			} else {
				// lightweight tags
				tagIter, err := r.Log(&git.LogOptions{From: ref.Hash(), Order: git.LogOrderCommitterTime})
				if err != nil {
					log.Println("failed to get tag log for ", ref, ": ", err)
					return nil
				}
				lastCmt, err := tagIter.Next()
				if err != nil {
					log.Println("failed to get latest tag commit for ", ref.Hash(), ": ", err)
					return nil
				}
				tags = append(tags, refData{Ref: ref, LastCmt: lastCmt})
			}
		}
		// ignore refs that aren't tags or branches
		return nil
	})
	if err != nil {
		log.Println("failed to read refs: ", err)
	}

	refsPageData := refsData{
		headerData: &headerData{
			Name:         filename,
			StrippedName: filenameStripped,
			Description:  description,
			RelPath:      "",
			Readme:       readme,
			License:      lisence,
			Submodules:   submodules,
			Title:        "Refs",
			URL:          url,
		},
		Branches: branches,
		Tags:     tags,
	}

	refsPageTmpl, err := template.New("refs").Funcs(template.FuncMap{
		"RefName":    formatRefName,
		"FormatTime": templates.FormatTime}).Parse(refsTpl)

	if err != nil {
		log.Println("unable to load refs page template: ", err)
	}

	headerTmpl, err = template.Must(refsPageTmpl.Clone()).Parse(headerTpl)
	if err != nil {
		log.Println("unable to load header for refs page template: ", err)
	}

	out, err = os.Create("refs.html")
	if err != nil {
		log.Println("unable to create refs page file: ", err)
	}

	if err := headerTmpl.Execute(out, refsPageData); err != nil {
		log.Println("unable to fill refs page template: ", err)
	}
	err = out.Close()
	if err != nil {
		log.Println("unable to close refs page file: ", err)
	}
}
