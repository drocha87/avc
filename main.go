package main

import (
	"fmt"
	"os"
	"path"
	"strings"
	"unicode"

	"golang.org/x/net/html"
)

type VueFile struct {
	Folder  string
	Name    string
	Content *html.Tokenizer
	Counter uint32
	UsedBy  []string
}

var components []VueFile
var views []VueFile

func main() {
	// TODO: check if the src/components folder exists

	CollectVueFiles("src/components", &components)
	CollectVueFiles("src/views", &views)

	for _, view := range views {
		for index := range components {
			CheckVueFileForComponent(view, &components[index])
		}
	}

	ReportComponentsUsage()
}

func CollectVueFiles(directory string, list *[]VueFile) {
	files, err := os.ReadDir(directory)
	if err != nil {
		panic(fmt.Sprintf("could not read the %s directory", directory))
	}
	for _, f := range files {
		fullPath := directory + "/" + f.Name()
		if f.IsDir() {
			CollectVueFiles(fullPath, list)
		} else {
			if path.Ext(f.Name()) == ".vue" {
				content, err := ReadVueFile(fullPath)
				if err != nil {
					panic(fmt.Sprintf("Could not read the content of the vue file %s\n", fullPath))
				}
				*list = append(*list, VueFile{
					Folder:  directory,
					Name:    strings.TrimSuffix(f.Name(), ".vue"),
					Content: content,
					Counter: 0,
				})
			}
		}
	}
}

func ReadVueFile(file string) (*html.Tokenizer, error) {
	bs, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return html.NewTokenizer(strings.NewReader(string(bs))), nil
}

func CheckVueFileForComponent(file VueFile, component *VueFile) {
	for {
		tokenType := file.Content.Next()
		if tokenType == html.ErrorToken {
			return
		}
		if tokenType == html.StartTagToken {
			// NOTE: in vue a component like `ComponentA` can be accessed as <ComponentA /> or <component-a></component-a>
			token := file.Content.Token()

			componentInKebabCase := VueComponentInKebabCase(component.Name)
			if token.Data == strings.ToLower(component.Name) || token.Data == componentInKebabCase {
				// keep a distinct list of vue files
				if !StringSliceContains(component.UsedBy, file.Name) {
					component.UsedBy = append(component.UsedBy, file.Name)
				}
				component.Counter += 1
			}
		}
	}
}

// this function generates the kebab-case version of component name
func VueComponentInKebabCase(name string) string {
	var result []rune
	for index, ch := range name {
		if unicode.IsUpper(ch) && index >= 1 {
			result = append(result, '-')
		}
		result = append(result, unicode.ToLower(ch))
	}
	return string(result)
}

func ReportComponentsUsage() {
	for _, component := range components {
		if component.Counter > 0 {
			fmt.Printf("Component %v.vue is used %v ", component.Name, component.Counter)
			if component.Counter > 1 {
				fmt.Printf("times\n")
			} else {
				fmt.Printf("time\n")
			}
			fmt.Printf("\t")
			for _, c := range component.UsedBy {
				fmt.Printf("%v ", c)
			}
			fmt.Printf("\n")
		} else {
			fmt.Printf("Component %v.vue is not in use\n", component.Name)
		}
	}
}

func StringSliceContains(list []string, item string) bool {
	for _, s := range list {
		if s == item {
			return true
		}
	}
	return false
}
