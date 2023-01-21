package main

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"unicode"

	"golang.org/x/net/html"
)

type VueFile struct {
	Folder        string
	Name          string
	KebabCaseName string
	Tags          []string
	Counter       uint32
	UsedBy        []string
	Recursive     bool
}

var components []VueFile
var views []VueFile

func main() {
	// TODO: check if we are in vue project

	// TODO: check the component also in the component folder
	// TODO: check for cyclic components
	// TODO: check for consistency in case on components usage

	CollectVueFiles("src/components", &components)
	CollectVueFiles("src/views", &views)

	for i := range views {
		for j := range components {
			CheckVueFileForComponent(&views[i], &components[j])
		}
	}

	for i := range components {
		for j := range components {
			CheckVueFileForComponent(&components[i], &components[j])
		}
	}

	ReportComponentsUsage()
}

func CollectVueFiles(directory string, list *[]VueFile) {
	// FIXME: handle the case where we import the component in the script section

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
				name := strings.TrimSuffix(f.Name(), ".vue")
				*list = append(*list, VueFile{
					Folder:        directory,
					Name:          name,
					KebabCaseName: VueComponentInKebabCase(name),
					Tags:          content,
					Counter:       0,
				})
			}
		}
	}
}

func ReadVueFile(file string) ([]string, error) {
	bs, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	content := html.NewTokenizer(strings.NewReader(string(bs)))
	var tags []string

	for {
		tokenType := content.Next()
		if tokenType == html.ErrorToken {
			break
		}

		// We are interested for now only on open tags
		if tokenType == html.StartTagToken {
			token := content.Token()
			tags = append(tags, token.Data)
		}
	}
	return tags, nil
}

func CheckVueFileForComponent(file *VueFile, component *VueFile) {
	for _, tag := range file.Tags {
		// NOTE: in vue a component like `ComponentA` can be accessed as <ComponentA /> or <component-a></component-a>
		if tag == strings.ToLower(component.Name) || tag == component.KebabCaseName {
			if file == component {
				// in this case we don't need to keep checking since we already marked it as a recursive component
				file.Recursive = true
				return
			}
			// keep a distinct list of vue files
			if !StringSliceContains(component.UsedBy, file.Name) {
				component.UsedBy = append(component.UsedBy, file.Name)
			}
			component.Counter += 1
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
	sort.SliceStable(components, func(i, j int) bool {
		return components[i].Counter > components[j].Counter
	})

	for _, component := range components {
		if component.Counter > 0 {
			fmt.Printf("%v.vue: is used %v ", component.Name, component.Counter)
			if component.Counter > 1 {
				fmt.Printf("times")
			} else {
				fmt.Printf("time")
			}
			if component.Recursive {
				fmt.Printf(" **recursive component**")
			}
			fmt.Printf("\n")
		} else {
			fmt.Printf("%v.vue: not in use\n", component.Name)
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
