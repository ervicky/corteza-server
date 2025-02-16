package renderer

import (
	"io"
	"io/ioutil"
	"text/template"
)

func preprocPlainTemplate(tpl io.Reader, pp map[string]io.Reader) (*template.Template, error) {
	bb, err := ioutil.ReadAll(tpl)
	if err != nil {
		return nil, err
	}

	gtpl := template.New("text/plain_render")

	// Prep the original template
	t, err := gtpl.Parse(string(bb))
	if err != nil {
		return nil, err
	}

	// Prep partials
	for _, p := range pp {
		bb, err = ioutil.ReadAll(p)
		if err != nil {
			return nil, err
		}

		t, err = gtpl.Parse(string(bb))
		if err != nil {
			return nil, err
		}
	}

	return t, nil
}
