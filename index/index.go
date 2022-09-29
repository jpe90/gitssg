package index

import (
	"html/template"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/go-git/go-git/v5/plumbing/object"
	"jeskin.net/gitssg/templates"

	"github.com/go-git/go-git/v5"
)

type Repository struct {
	Name        string
	Description string
	LastCommit  time.Time
	Owner       string
	LandingPage string
}

const indexPage = `<!DOCTYPE html>
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
<meta name="viewport" content="width=device-width, initial-scale=1" />
<title>Repositories</title>
<link rel="icon" type="image/png" href="favicon.png" />
<link rel="stylesheet" type="text/css" href="style.css" />
</head>
<body>
<table>
<tr>
<td><img src="logo.png" alt="" width="32" height="32" /></td>
<td><span class="desc">Repositories</span></td>
</tr>
<tr>
<td></td>
<td></td>
</tr>
</table>
<hr/>
<div id="content">
<table id="index"><thead>
<tr>
<td><b>Name</b></td>
<td><b>Description</b></td>
<td><b>Owner</b></td>
<td><b>Last commit</b></td>
</tr></thead><tbody>
{{range .}}<tr>
<td><a href="{{.LandingPage}}">{{.Name}}</a></td></td>
<td>{{.Description}}</td>
<td>{{FormatTime .LastCommit}}</td>
<td>{{.Owner}}</td></tr>{{end}}</tbody>
</table>
</div>
</body>
</html>
`

func Run(repodirs []string) {
	var repositories []Repository

	for _, repodir := range repodirs {
		var readmeFilename string

		abs, err := filepath.Abs(repodir)

		if err != nil {
			log.Fatalln(repodir, " does not have a valid file path: ", err)
			continue
		}

		filename := filepath.Base(abs)
		filenameStripped := strings.TrimSuffix(filename, filepath.Ext(filename))
		r, err := git.PlainOpen(abs)

		if err != nil {
			log.Fatalln(repodir, " is not a git repository.", err)
			continue
		}

		descrData, err := os.ReadFile(path.Join(repodir, "/description"))
		if err != nil {
			log.Println("description file not found: ", err)
			continue
		}
		description := string(descrData)

		ownerData, err := os.ReadFile(path.Join(repodir, "/owner"))
		if err != nil {
			log.Println("owner file not found: ", err)
		}
		owner := string(ownerData)

		head, err := r.Head()
		if err != nil {
			log.Fatalln("could not find HEAD: ", err)
		}

		commit, err := r.CommitObject(head.Hash())
		if err != nil {
			log.Fatalln("could not hash ref for HEAD: ", err)
		}
		time := commit.Author.When

		tree, err := commit.Tree()
		if err != nil {
			log.Println("couldn't get tree for head commit: ", err)
		}

		// If the latest commit has a README, link to it.
		// Otherwise link to repo's reflog.
		tree.Files().ForEach(func(f *object.File) error {
			fname := f.Name
			res, e := regexp.MatchString("readme", strings.ToLower(fname))
			if e == nil && res {
				readmeFilename = fname
			}
			return nil
		})

		landingPage := filenameStripped + "/"
		if readmeFilename != "" {
			landingPage += readmeFilename
		} else {
			landingPage += "log.html"
		}

		newRepository := Repository{
			Name:        filenameStripped,
			Description: description,
			LastCommit:  time,
			Owner:       owner,
			LandingPage: landingPage,
		}
		repositories = append(repositories, newRepository)
	}

	tmpl, err := template.New("index").Funcs(template.FuncMap{
		"FormatTime": templates.FormatTime}).Parse(indexPage)
	if err != nil {
		log.Println("unable to load index page template: ", err)
	}

	err = tmpl.Execute(os.Stdout, repositories)
	if err != nil {
		log.Println("unable to fill index page template: ", err)
	}
}
