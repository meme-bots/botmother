package botmother

import (
	"bytes"
	"os"
	"text/template"

	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
)

type (
	Command struct {
		Text        string `yaml:"text,omitempty"`
		Description string `yaml:"description,omitempty"`
	}

	Button struct {
		Unique string `yaml:"unique,omitempty"`
		Text   string `yaml:"text"`
		Url    string `yaml:"url"`
		Data   string `yaml:"data"`
	}

	Page struct {
		ID          string     `yaml:"id,omitempty"`
		Message     string     `yaml:"message"`
		Alternative string     `yaml:"alternative"`
		Items       [][]Button `yaml:"items"`
	}

	Layout_ struct {
		Commands []Command `yaml:"commands,omitempty"`
		Pages    []Page    `yaml:"pages,omitempty"`
	}

	Layout struct {
		locale   string
		Commands []Command
		Pages    map[string]*Page
		Texts    map[string]string
	}
)

func NewLayout(locale, layoutFile, localeFile string) (*Layout, error) {
	var texts map[string]string
	data, err := os.ReadFile(localeFile)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, &texts); err != nil {
		return nil, err
	}

	var layout_ Layout_
	content, err := os.ReadFile(layoutFile)
	if err != nil {
		return nil, err
	}

	tmpl := template.Must(template.New(locale).Delims("<<", ">>").Parse(string(content)))
	var buf bytes.Buffer
	tmpl.Execute(&buf, texts)

	err = yaml.Unmarshal(buf.Bytes(), &layout_)
	if err != nil {
		return nil, err
	}

	layout := &Layout{
		locale:   locale,
		Commands: layout_.Commands,
		Pages:    make(map[string]*Page),
		Texts:    texts,
	}

	pages := lo.Map(layout_.Pages, func(p Page, _ int) *Page {
		np := &Page{
			ID:          p.ID,
			Message:     p.Message,
			Alternative: p.Alternative,
		}
		np.Items = lo.Map(p.Items, func(bb []Button, _ int) []Button {
			return lo.Map(bb, func(b Button, _ int) Button {
				nb := Button{
					Unique: b.Unique,
					Url:    b.Url,
					Data:   b.Data,
				}
				if t, ok := texts[b.Unique]; ok {
					nb.Text = t
				}
				return nb
			})
		})
		return np
	})
	layout.Pages = lo.KeyBy(pages, func(p *Page) string { return p.ID })

	return layout, err
}

func (lo *Layout) GetPage(pageID string) *Page {
	return lo.Pages[pageID]
}

func (lo *Layout) GetText(text string) string {
	return lo.Texts[text]
}
