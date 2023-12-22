package main

import (
	"archive/tar"
	"compress/gzip"
	"encoding/gob"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/meilisearch/meilisearch-go"
)

type Base struct {
	Branch       string
	Repository   string
	Architecture string
	RootPath     string
	IndexUID     string
	LastSet      mapset.Set[string]
	NextSet      mapset.Set[string]
	Content      string
	Packages     []Package
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

var (
	key  string
	url  string
	path string
)

var (
	authors    = map[string]Maintainer{}
	re_author  = regexp.MustCompile(`^(\w+\s+\w+)\s+<(.+)>$`)
	requires   = map[string]string{}
	re_require = regexp.MustCompile(`(?:.*:)?([^=<>]*)`)
)

func (b *Base) Lock() {
	_, err := os.Stat(b.RootPath + "/cache.lock")
	if errors.Is(err, os.ErrNotExist) {
		f, _ := os.Create(b.RootPath + "/cache.lock")
		f.Close()
	} else if err != nil {
		log.Panicln(err)
	} else {
		log.Panicf("%s has been locked, wait existing process or remove cache.lock\n", path)
	}
}

func (b *Base) UnLock() {
	os.Remove(b.RootPath + "/cache.lock")
}

func init() {
	flag.StringVar(&url, "url", "http://localhost:7700", "meilisearch address")
	flag.StringVar(&key, "key", "", "meilisearch master key")
	flag.StringVar(&path, "path", "", "path of APKINDEX.tar.gz")
	flag.Parse()

	if key == "" || path == "" {
		log.Panicln("KEY and PATH are required")
	}

	base.Init(path)
}

func main() {
	fmt.Println("Here", base.RootPath)
	base.Lock()

	if err := base.LoadCache(); err != nil {
		log.Panicln(err)
	}

	if err := base.Read(); err != nil {
		log.Panicln(err)
	}
	count := base.Parse()
	if count == 0 {
		log.Println("Nothing to update")
		return
	}
	if err := base.SaveCache(); err != nil {
		log.Panicln(err)
	}

	fmt.Printf("%9s in index %14s, updated with %5d, removed %5d, now a total of %5d\n",
		base.Repository, base.IndexUID, count,
		base.LastSet.Cardinality(), base.NextSet.Cardinality())

	// connect meilisearch
	client := meilisearch.NewClient(meilisearch.ClientConfig{
		Host:   url,
		APIKey: key,
	})
	if _, err := client.GetKeys(nil); err != nil {
		log.Panicln(err)
	}
	delete_ids, _ := base.LastSet.MarshalJSON()
	task1, err := client.Index(base.IndexUID).DeleteDocumentsByFilter("id IN " + string(delete_ids))
	if err != nil {
		fmt.Println(err)
	}
	task2, err := client.WaitForTask(task1.TaskUID)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(task2)
	task3, err := client.Index(base.IndexUID).AddDocuments(base.Packages)
	if err != nil {
		fmt.Println(err)
	}
	task4, err := client.WaitForTask(task3.TaskUID)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(task4)

	base.UnLock()
	fmt.Println("Done", base.RootPath)
}

// empty when error
func (b *Base) Read() error {
	// file
	file, err := os.Open(b.RootPath + "/APKINDEX.tar.gz")
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
	out_pkgs := make([]Package, 0, len(sections))
	for _, section := range sections {
		pkg := parse_package(section)
		if pkg == nil {
			continue
		}
		out_pkgs = append(out_pkgs, *pkg)
	}
	b.Packages = out_pkgs
	return len(out_pkgs)
}

func parse_package(str string) *Package {
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
			uuid, old := base.NotNew(value)
			if old {
				return nil
			}
			pkg.CheckSum = uuid
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

func (b *Base) NotNew(check_sum string) (string, bool) {
	uuid := strings.TrimRight(check_sum, "=")[2:]
	uuid = strings.ReplaceAll(uuid, "+", "-")
	uuid = strings.ReplaceAll(uuid, "/", "_")
	b.NextSet.Add(uuid)
	if b.LastSet != nil && b.LastSet.Contains(uuid) {
		b.LastSet.Remove(uuid)
		return uuid, true
	}
	return uuid, false
}

func get_author(str string) Maintainer {
	if value, ok := authors[str]; ok {
		return value
	}
	match := re_author.FindStringSubmatch(str)
	if len(match) != 3 {
		return Maintainer{}
	}
	name := match[1]
	email := match[2]
	author := Maintainer{
		Name:  name,
		Email: email,
	}
	authors[str] = author
	return author
}

func get_require(str string) []string {
	outputs := mapset.NewSet[string]()
	inputs := strings.Split(str, " ")
	for _, input := range inputs {
		if outputs.Contains(input) {
			continue
		}
		if value, ok := requires[input]; ok {
			outputs.Add(value)
			continue
		}
		match := re_require.FindStringSubmatch(input)
		if len(match) != 2 {
			fmt.Println("regexp require fail:", input)
			continue
		}
		requires[input] = match[1]
		outputs.Add(match[1])
	}
	return outputs.ToSlice()
}

func (b *Base) Init(path string) {
	// root path, like /home/qaq/rsync/v3.19/main/aarch64
	b.RootPath = filepath.Dir(path)

	// branch, repository, architecture
	patrs := strings.Split(b.RootPath, string(filepath.Separator))
	length := len(patrs)
	b.Architecture = patrs[length-1]
	b.Repository = patrs[length-2]
	b.Branch = patrs[length-3]

	// search index name
	b.IndexUID = fmt.Sprintf("%s_%s", strings.ReplaceAll(b.Branch, ".", "_"), b.Architecture)

	// NewSet for cache.gob
	b.NextSet = mapset.NewThreadUnsafeSet[string]()
	b.LastSet = mapset.NewThreadUnsafeSet[string]()
}

func (b *Base) LoadCache() error {
	f, err := os.Open(b.RootPath + "/cache.gob")
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()

	decoder := gob.NewDecoder(f)
	return decoder.Decode(&b.LastSet) // TODO maybe bug
}

func (b *Base) SaveCache() error {
	f, err := os.Create(b.RootPath + "/cache.gob")
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := gob.NewEncoder(f)
	return encoder.Encode(b.NextSet)
}
