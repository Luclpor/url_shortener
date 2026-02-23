package service

import (
	"github.com/Luclpor/url_shortener.git/internal/model"
)

func GetShortURL(key string) string {
	urls := model.GetUrls()
	v, ok := urls[key]
	if !ok {
		return ""
	}
	return v
}
