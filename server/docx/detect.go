package docx

import (
	"bytes"
	"errors"
	"regexp"
	"sort"

	"github.com/unidoc/unioffice/document"
)

var rxVar = regexp.MustCompile(`\{\{[a-zA-Z0-9_.]+\}\}`)

// DetectFields reads a .docx (bytes) and returns unique placeholders without {{ }}
func DetectFields(docBytes []byte) ([]string, error) {
	doc, err := document.Read(bytes.NewReader(docBytes), int64(len(docBytes)))
	if err != nil {
		return nil, errors.New("unable to read .docx (ensure it is a valid Word file)")
	}
	set := map[string]struct{}{}
	for _, p := range doc.Paragraphs() {
		for _, r := range p.Runs() {
			text := r.Text()
			m := rxVar.FindAllString(text, -1)
			for _, v := range m {
				// v is like {{client_name}} → strip braces
				key := v[2 : len(v)-2]
				set[key] = struct{}{}
			}
		}
	}
	fields := make([]string, 0, len(set))
	for k := range set {
		fields = append(fields, k)
	}
	sort.Strings(fields)
	if len(fields) == 0 {
		return fields, nil
	}
	return fields, nil
}
