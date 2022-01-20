package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/deckarep/golang-set"
)

type Translation struct {
	pre   string
	key   string
	value string
	lc    int
}

type StringsFile struct {
	path         string
	language     string
	translations []Translation
}

const (
	PRE = iota
	KEY
	EQUALS
	VALUE
)

func (this StringsFile) Translations(key string) []Translation {
	ret := []Translation{}
	for _, translation := range this.translations {
		if translation.key == key {
			ret = append(ret, translation)
		}
	}
	return ret
}

func glob(dir string, ext string) ([]string, error) {

	files := []string{}
	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if filepath.Ext(path) == ext {
			files = append(files, path)
			// fmt.Printf("found %s (%s)\n", path, filepath.Ext(path))
		}
		return nil
	})

	return files, err
}

func strings_keys(dirs []string, yield func(path string, lc int, key string, value string, pre string, language string)) {
	r, _ := regexp.Compile("([^\\/]*)\\.lproj")

	for _, dir := range dirs {
		// files, _ := filepath.Glob(dir + "/**/*.lproj/*.strings")
		files, _ := glob(dir, ".strings")
		for _, file := range files {

			match := r.FindStringSubmatch(file)
			if len(match) >= 2 {
				language := match[1]

				fd, err := os.Open(file)
				if err != nil {
					log.Fatal(err)
				}
				defer fd.Close()

				scanner := bufio.NewScanner(fd)

				var buffer bytes.Buffer
				pre := ""
				key := ""
				value := ""
				lc := 0
				for scanner.Scan() {
					line := scanner.Text()
					lc += 1

					// fmt.Printf("LINE: %s\n", line)

					if line == "" {
						continue
					} else {
						line = line + "\n"
					}

					state := PRE
					buffer.Reset()

					for _, c := range line {
						b := buffer.Bytes()
						switch state {
						case PRE:
							switch {
							case bytes.HasSuffix(buffer.Bytes(), []byte("\"")):
								pre = string(bytes.TrimSuffix(b, []byte("\"")))
								// fmt.Printf("PRE: %s\n", pre)
								buffer.Reset()
								buffer.WriteRune(c)
								state = KEY
							default:
								buffer.WriteRune(c)
							}
						case KEY:
							switch {
							case bytes.HasSuffix(b, []byte("\"")):
								key = string(bytes.TrimSuffix(b, []byte("\"")))
								// fmt.Printf("KEY: %s\n", key)
								buffer.Reset()
								buffer.WriteRune(c)
								state = EQUALS
							default:
								buffer.WriteRune(c)
							}
						case EQUALS:
							switch {
							case c == ' ':
							case bytes.HasSuffix(b, []byte("=\"")):
								// fmt.Printf("EQUALS:\n")
								buffer.Reset()
								buffer.WriteRune(c)
								state = VALUE
							default:
								buffer.WriteRune(c)
							}
						case VALUE:
							switch {
							case c == '\n' && bytes.HasSuffix(b, []byte("\";")):
								value = string(bytes.TrimSuffix(b, []byte("\";")))
								// fmt.Printf("VALUE: %s\n", value)
								buffer.Reset()
								buffer.WriteRune(c)
								state = PRE

								yield(file, lc, key, value, pre, language)

								pre = ""
								key = ""
								value = ""
							default:
								buffer.WriteRune(c)
							}
						}
					}
				}

				if err := scanner.Err(); err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}

// func xib_keys(dirs []string, yield func(path string, key string)) {

//   for _, dir := range dirs {
//     // files, _ := filepath.Glob(dir + "/**/*.xib")
//     files, _ := glob(dir, ".xib")
//     for _, file := range files {
//       // fmt.Printf("reading %s\n", file)
//       content, _ := ioutil.ReadFile(file)

//       doc, _ := gokogiri.ParseXml(content)
//       defer doc.Free()

//       nodes_title, _ := doc.Root().Search(xpath.Compile("//string[@key=\"NSTitle\"]"))
//       for _, node := range nodes_title {
//         key := node.Content()
//         if key != "" {
//           yield(file, key)
//         }
//       }

//       nodes_responder, _ := doc.Root().Search(xpath.Compile("//label|//button|//textField"))
//       for _, node := range nodes_responder {

//         label := node.Attr("userLabel")
//         if label == "" {
//           nodes_userlabel, _ := node.Search(xpath.Compile("ancestor-or-self::*[@userLabel]/@userLabel"))
//           if len(nodes_userlabel) > 0 {
//             label = nodes_userlabel[0].Content()
//           }
//         }

//         if label != "File's Owner" {
//           for _, attr := range []string{ "text", "title", "placeholder" } {
//             nodes_attributes, _ := node.Search(xpath.Compile(fmt.Sprintf(".//*[@%s]|.", attr)))
//             for _, n := range nodes_attributes {
//               key := n.Attr(attr)
//               if key != "" {
//                 // fmt.Printf(" %s = '%s' (%s)\n", attr, key, label)
//                 yield(file, key)
//               }
//             }
//           }
//         }
//       }
//     }
//   }
// }

func code_keys(dirs []string, yield func(path string, lc int, key string)) {

	// r, _ := regexp.Compile("NSLocalizedString\\(@\"(.*?)\",")
	r, _ := regexp.Compile("\"(.*?)\".localized")

	for _, dir := range dirs {
		// files, _ := glob(dir, ".m")
		files, _ := glob(dir, ".swift")

		for _, file := range files {

			fd, err := os.Open(file)
			if err != nil {
				log.Fatal(err)
			}
			defer fd.Close()

			scanner := bufio.NewScanner(fd)

			lc := 0
			for scanner.Scan() {
				line := scanner.Text()
				lc += 1

				if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") {
					continue
				}

				matches := r.FindAllStringSubmatch(line, -1)
				for _, match := range matches {
					yield(file, lc, match[1])
				}
			}
		}
	}
}

func show(dirs []string) int {
	ret := 0

	keys_available := mapset.NewThreadUnsafeSet()

	strings_map := make(map[string]*StringsFile)

	// gather keys from strings file
	strings_keys(dirs, func(path string, lc int, key string, value string, pre string, language string) {
		keys_available.Add(key)

		strings_file, exists := strings_map[path]
		if !exists {
			strings_file = &StringsFile{path, language, []Translation{}}
			strings_map[path] = strings_file
		}

		translation := Translation{pre, key, value, lc}

		strings_file.translations = append(strings_file.translations, translation)
	})

	// print the number of translations per langugage
	for _, strings_file := range strings_map {
		fmt.Printf("translations: %s = %d\n", strings_file.path, len(strings_file.translations))
	}
	fmt.Printf("--\n")

	// print translation for each key
	for key := range keys_available.Iter() {
		for _, strings_file := range strings_map {
			for _, translation := range strings_file.translations {
				if translation.key == key {
					fmt.Printf("%s: %s = %s\n", strings_file.language, key, translation.value)
				}
			}
		}
		fmt.Printf("--\n")
	}

	return ret
}

func generate(dirs []string) int {

	ret := 0

	keys_available := mapset.NewThreadUnsafeSet()

	strings_map := make(map[string]*StringsFile)

	// gather keys from strings file
	strings_keys(dirs, func(path string, lc int, key string, value string, pre string, language string) {
		keys_available.Add(key)

		strings_file, exists := strings_map[path]
		if !exists {
			strings_file = &StringsFile{path, language, []Translation{}}
			strings_map[path] = strings_file
		}

		translation := Translation{pre, key, value, lc}

		strings_file.translations = append(strings_file.translations, translation)
	})

	keys_unused := keys_available.Clone()
	keys_missing := mapset.NewThreadUnsafeSet()

	// check keys in the code base
	code_keys(dirs, func(path string, lc int, key string) {
		// fmt.Printf("code: %s\n", key)
		if keys_available.Contains(key) {
			keys_unused.Remove(key)
		} else {
			keys_missing.Add(key)
			fmt.Printf("\"%s\" = \"%s\";\n", key, key)
			// fmt.Printf("\"%s\" = \"%s\"; # %s:%d\n", key, key, path, lc)
			ret = 1
		}
	})

	return ret
}

func check(dirs []string) int {

	ret := 0

	keys_available := mapset.NewThreadUnsafeSet()

	strings_map := make(map[string]*StringsFile)

	// gather keys from strings file
	strings_keys(dirs, func(path string, lc int, key string, value string, pre string, language string) {
		keys_available.Add(key)

		strings_file, exists := strings_map[path]
		if !exists {
			strings_file = &StringsFile{path, language, []Translation{}}
			strings_map[path] = strings_file
		}

		translation := Translation{pre, key, value, lc}

		strings_file.translations = append(strings_file.translations, translation)
	})

	keys_unused := keys_available.Clone()
	keys_missing := mapset.NewThreadUnsafeSet()

	// check keys in the code base
	code_keys(dirs, func(path string, lc int, key string) {
		// fmt.Printf("code: %s\n", key)
		if keys_available.Contains(key) {
			keys_unused.Remove(key)
		} else {
			keys_missing.Add(key)
			fmt.Printf("%s:%d: error: code uses missing key '%s'\n", path, lc, key)
			ret = 1
		}
	})

	// xib_keys(dirs, func(path string, key string){
	//   if !strings.HasPrefix(key, "!") {
	//     // fmt.Printf("xib: %s\n", key)
	//     if keys_available.Contains(key) {
	//       keys_unused.Remove(key)
	//     } else {
	//       keys_missing.Add(key)
	//       fmt.Printf("%s:%d: error: xib uses missing key '%s'\n", path, 0, key)
	//       ret = 1
	//     }
	//   }
	// })

	// unused
	for key_unused := range keys_unused.Iter() {
		for _, strings_file := range strings_map {
			for _, translation := range strings_file.Translations(key_unused.(string)) {
				if !strings.HasPrefix(translation.key, "_") {
					fmt.Printf("%s:%d: warning: '%s' has unused key '%s'\n", strings_file.path, translation.lc, strings_file.language, translation.key)
				}
				ret = 1
			}
		}
	}

	// duplicates
	keys_needed := keys_available.Union(keys_missing)
	for key_needed := range keys_needed.Iter() {
		for _, strings_file := range strings_map {
			key := key_needed.(string)
			translations := strings_file.Translations(key)
			switch len(translations) {
			case 0:
				fmt.Printf("%s:%d: error: (%s) missing key '%s'\n", strings_file.path, 0, strings_file.language, key)
				ret = 1
			case 1:
				for _, translation := range translations {
					if strings.TrimSpace(translation.value) == "" {
						fmt.Printf("%s:%d: warning: (%s) empty key '%s'\n", strings_file.path, translation.lc, strings_file.language, key)
					}
				}
			default:
				for _, translation := range translations {
					fmt.Printf("%s:%d: error: (%s) duplicate key '%s'\n", strings_file.path, translation.lc, strings_file.language, key)
				}
				ret = 1
			}
		}
	}

	return ret

}

func path() string {
	env := os.Getenv("PROJECT_DIR")
	xcode := env != ""

	if xcode {
		// fmt.Println("running via Xcode")
		return env
	} else {
		// fmt.Println("running via command line")
		if len(os.Args) == 2 {
			return os.Args[1]
		} else {
			return "."
		}
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func main() {

	p_generate := flag.Bool("generate", false, "output keys")

	flag.Parse()

	if len(flag.Args()) > 2 {
		log.Fatal("too many args")
	}

	dir := path()

	content, _ := ioutil.ReadFile(dir + "/.stringsignore")
	lines := strings.Split(string(content), "\n")

	files_filtered := []string{}
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		basename := filepath.Base(path)

		if strings.HasPrefix(basename, ".") && len(basename) > 1 {
			// println("skipping hidden ", path)
			if info.IsDir() {
				return filepath.SkipDir
			} else {
				return nil
			}
		}

		if contains(lines, basename) {
			// println("skipping ignored ", path)
			if info.IsDir() {
				return filepath.SkipDir
			} else {
				return nil
			}
		}

		if info.Mode().IsRegular() {
			files_filtered = append(files_filtered, path)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	// files, _ := filepath.Glob(dir + "/[^\\.]*")
	// files_filtered := []string{}
	// for _, file := range files {
	// 	if !strings.HasPrefix(file, ".") {
	// 		if contains(lines, filepath.Base(file)) {
	// 			// fmt.Printf("ignoring '%s'\n", file)
	// 		} else {
	// 			files_filtered = append(files_filtered, file)
	// 		}
	// 	}
	// }

	if *p_generate {
		os.Exit(generate(files_filtered))
	} else {
		os.Exit(check(files_filtered))
	}
}
