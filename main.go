package main

import (
	"archive/tar"
	"compress/gzip"

	// "encoding/json"
	// "flag"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
)

type Base struct {
	Branch     string
	Repository string
	Arch       string
	Path       string
	IndexUID   string
	LastSet    mapset.Set[string]
	NextSet    mapset.Set[string]
	Content    string
	Packages   []*Package
	temp       struct {
		authors    map[string]Maintainer
		re_author  *regexp.Regexp
		requires   map[string]string
		re_require *regexp.Regexp
	}
}

var base Base

type Package struct {
	CheckSum      string     `json:"id"` // for search and react
	Name          string     `json:"package"`
	Version       string     `json:"version"`
	Description   string     `json:"description"`
	FileSize      int        `json:"file_size"`
	InstalledSize int        `json:"installed_size"`
	ProjectURL    string     `json:"project"`
	License       string     `json:"license"`
	NotSub        bool       `json:"not_sub"`
	Origin        string     `json:"origin"`
	Depends       []string   `json:"depends"`
	Provides      []string   `json:"provides"`
	Repository    string     `json:"repository"`
	Commit        string     `json:"commit"`
	BuildTime     string     `json:"build_time"`
	Maintainer    Maintainer `json:"maintainer"`
}

type Maintainer struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

var test_path = "/home/qaq/rsync/v3.18/main/x86_64/APKINDEX.tar.gz"

func locked(path string) bool {
	_, err := os.Stat(path + "/cache.lock")
	if os.IsNotExist(err) {
		return false
	} else if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(path, "lock")
	}
	return true
}

func main() {
	base.Path = filepath.Dir(test_path)
	fmt.Println(base.Path)
	if locked(base.Path) {
		os.Exit(1)
	}
	defer os.Remove(base.Path + "/cache.lock")
	base.Init()
	base.LoadCache()
	if err := base.Read(); err != nil {
		fmt.Println(err)
		return
	}
	count := base.Parse()
	if count == 0 {
		fmt.Println("Nothing to update")
		return
	}
	fmt.Printf("index [%s - %s]\tUpdate:\t%d\n", base.Branch, base.Arch, count)
	fmt.Printf("%#v\n", base.Packages[100])
	// connect meilisearch
}

// empty when error
func (b *Base) Read() error {
	// file
	file, err := os.Open(b.Path + "/APKINDEX.tar.gz")
	if err != nil {
		return err
	}
	defer file.Close()
	// gzip
	gzip_reader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzip_reader.Close()
	// tar
	tar_reader := tar.NewReader(gzip_reader)
	for {
		header, err := tar_reader.Next()
		// if err == io.EOF {
		// 	// come to ending
		// 	fmt.Println("404 not found in tar")
		// 	break
		// }
		if err != nil {
			return err
		}
		if header.Name == "APKINDEX" {
			buffer := new(strings.Builder)
			buffer.Grow(int(header.Size))
			_, err := io.Copy(buffer, tar_reader)
			if err != nil {
				return err
			}
			b.Content = buffer.String()
			return nil
		}
	}
}
func (b *Base) Parse() int {
	sections := strings.Split(strings.TrimRight(b.Content, "\n"), "\n\n")
	outpkgs := make([]*Package, 0, len(sections))
	for _, section := range sections {
		pkg := parse_package(section)
		if pkg == nil {
			continue
		}
		// if len(outpkgs) == 3 {
		// 	break
		// }
		outpkgs = append(outpkgs, pkg)
	}
	b.Packages = outpkgs
	return len(outpkgs)
}
func parse_package(str string) *Package {
	// fmt.Println(str)
	// fmt.Println(base.NextSet.Cardinality())
	pkg := Package{
		Repository: base.Repository,
	}
	lines := strings.Split(str, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			fmt.Println("length of part is not two?")
			return nil
		}
		key, value := parts[0], parts[1]
		switch key {
		case "C":
			pkg.CheckSum = value
			if base.NotNew(value) {
				return nil
			}
		case "P":
			pkg.Name = value
		case "V":
			pkg.Version = value
		case "S":
			pkg.FileSize, _ = strconv.Atoi(value)
		case "I":
			pkg.InstalledSize, _ = strconv.Atoi(value)
		case "T":
			pkg.Description = value
		case "U":
			pkg.ProjectURL = value
		case "L":
			pkg.License = value
		case "o":
			pkg.Origin = value
			pkg.NotSub = (pkg.Name == pkg.Origin)
		case "m":
			pkg.Maintainer = get_author(value)
		case "t":
			pkg.BuildTime = value
		case "c":
			pkg.Commit = value
		case "D":
			pkg.Depends = get_require(value)
		case "p":
			pkg.Provides = get_require(value)
		default:
			// fmt.Println("parser key, continue:", key)
			continue
		}
	}
	return &pkg
}

func (b *Base) NotNew(uuid string) bool {
	b.NextSet.Add(uuid)
	if b.LastSet != nil && b.LastSet.Contains(uuid) {
		b.LastSet.Remove(uuid)
		return true
	}
	return false
}
func get_author(str string) Maintainer {
	maintainer := base.temp.authors
	re := base.temp.re_author
	if value, ok := maintainer[str]; ok {
		return value
	}
	match := re.FindStringSubmatch(str)
	if len(match) != 3 {
		return Maintainer{}
	}
	name := match[1]
	email := match[2]
	author := Maintainer{
		Name:  name,
		Email: email,
	}
	maintainer[str] = author
	return author
}

func get_require(str string) []string {
	depends := base.temp.requires
	re := base.temp.re_require
	outputs := mapset.NewSet[string]()
	inputs := strings.Split(str, " ")
	for _, input := range inputs {
		if outputs.Contains(input) {
			continue
		}
		if value, ok := depends[input]; ok {
			outputs.Add(value)
			continue
		}
		match := re.FindStringSubmatch(input)
		if len(match) != 2 {
			fmt.Println("regexp require fail:", input)
			continue
		}
		depends[input] = match[1]
		outputs.Add(match[1])
	}
	return outputs.ToSlice()
}

func (b *Base) Init() {
	// get author
	b.temp.authors = map[string]Maintainer{}
	b.temp.re_author = regexp.MustCompile(`^(\w+\s+\w+)\s+<(.+)>$`)
	// get require
	b.temp.requires = map[string]string{}
	b.temp.re_require = regexp.MustCompile(`(?:.*:)?([^=<>]*)`)
	// root path, but not init here
	// b.Path = Dir(apkindex_path)
	// branch, repository, arch
	p := strings.Split(b.Path, string(filepath.Separator))
	length := len(p)
	b.Branch = p[length-1]
	b.Repository = p[length-2]
	b.Arch = p[length-3]
	// search index name
	b.IndexUID = fmt.Sprintf("%s-%s", b.Branch, strings.ReplaceAll(b.Arch, ".", "_"))
	// NewSet for cache.gob
	b.NextSet = mapset.NewSet[string]()
}
func (b *Base) LoadCache() {
	f, err := os.Open(b.Path + "/cache.gob")
	if os.IsNotExist(err) {
		return
	} else if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	decoder := gob.NewDecoder(f)
	decoder.Decode(&b.LastSet) // TODO maybe bug
}
func (b *Base) SaveCache() {
	f, err := os.Create(b.Path + "/cache.gob")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	encoder := gob.NewEncoder(f)
	encoder.Encode(b.NextSet)
}
