package main

import (
	"bufio"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"net/http"
	"os"
	"sort"
)

const (
	fullURLFile = "https://raw.githubusercontent.com/joscha82/wattpilot/da08c3fb387b06497e007bef1ff88f0112a080ea/src/wattpilot/ressources/wattpilot.yaml"
	output      = "wattpilot_mapping_gen.go"
)

func downloadWattpilotYaml() ([]byte, error) {

	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}
	// Put content on file
	resp, err := client.Get(fullURLFile)
	if err != nil {
		print(err)
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func main() {
	s, _ := downloadWattpilotYaml()
	a := make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(s), &a); err != nil {
		print(err)
		return
	}
	propertyMap := make(map[string]string)
	for _, v := range a["properties"].([]interface{}) {
		key := ""
		alias := ""
		data := v.(map[interface{}]interface{})
		for x, y := range data {

			switch x.(string) {
			case "key":
				key = y.(string)
			case "alias":
				alias = y.(string)
			}
		}
		if key != "" && alias != "" {
			propertyMap[alias] = key
		}
	}

	f, _ := os.Create(output)
	defer f.Close()

	w := bufio.NewWriter(f)
	if _, err := w.WriteString("package wattpilot\nvar propertyMap = map[string]string {\n"); err != nil {
		return
	}
	mk := make([]string, len(propertyMap))
	i := 0
	for k := range propertyMap {
		mk[i] = k
		i++
	}
	sort.Strings(mk)
	for _, i := range mk {
		s := propertyMap[i]
		if _, err := w.WriteString(fmt.Sprintf("\"%s\": \"%s\",\n", i, s)); err != nil {
			return
		}
	}
	if _, err := w.WriteString("}\n"); err != nil {
		return
	}
	w.Flush()

}
