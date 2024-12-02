package botmother

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/samber/lo"
	"gopkg.in/telebot.v3"
)

type (
	Meta struct {
		PageID    string  `json:"p,omitempty"`
		BtnUnique string  `json:"b,omitempty"`
		Token     string  `json:"t,omitempty"`
		Account   string  `json:"a,omitempty"`
		Amount    float64 `json:"m,omitempty"`
		MessageID int     `json:"mid,omitempty"`
		ChatID    int64   `json:"cid,omitempty"`
	}
)

const (
	META_PREFIX = "tg://meta/"
)

func GetMetaText(ctx telebot.Context, reply bool) (string, bool) {
	m := ctx.Message()
	if reply {
		m = m.ReplyTo
	}

	entity, ok := lo.Find(m.Entities, func(v telebot.MessageEntity) bool {
		return v.Type == telebot.EntityTextLink && strings.HasPrefix(v.URL, META_PREFIX)
	})
	if !ok {
		return "", false
	}

	return strings.TrimPrefix(entity.URL, META_PREFIX), true
}

func GetMetaData(ctx telebot.Context, reply bool) (*Meta, error) {
	metaText, ok := GetMetaText(ctx, reply)
	if !ok {
		return nil, ErrNotFound
	}
	data, err := base64.StdEncoding.DecodeString(metaText)
	if err != nil {
		return nil, err
	}
	var meta Meta
	if err = json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}
