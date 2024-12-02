package botmother

import (
	"bytes"
	"text/template"

	"gopkg.in/telebot.v3"
)

func Sprintf(text string, data interface{}) string {
	if data == nil {
		return text
	}
	t := template.Must(template.New("t").Parse(text))
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return text
	}
	return buf.String()
}

func CreateMarkup(buttons [][]Button, data interface{}) *telebot.ReplyMarkup {
	markup := &telebot.ReplyMarkup{ResizeKeyboard: true}
	rows := make([]telebot.Row, 0)
	for _, row := range buttons {
		btns := make([]telebot.Btn, 0)
		for _, b := range row {
			text := Sprintf(b.Text, data)
			if len(b.Url) > 0 {
				btns = append(btns, markup.URL(text, Sprintf(b.Url, data)))
			} else {
				if len(b.Data) > 0 {
					btns = append(btns, markup.Data(text, b.Unique, Sprintf(b.Data, data)))
				} else {
					btns = append(btns, markup.Data(text, b.Unique))
				}
			}
		}
		rows = append(rows, markup.Row(btns...))
	}
	markup.Inline(rows...)
	return markup
}

func (bm *BotMother) CreatePage(locale string, pageID string, data interface{}) (string, *telebot.ReplyMarkup, error) {
	layout := bm.GetLayout(locale)
	page := layout.Pages[pageID]

	t := template.Must(template.New("t").Parse(page.Message))
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", nil, err
	}

	return buf.String(), CreateMarkup(page.Items, data), nil
}
