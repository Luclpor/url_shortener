package service

import "github.com/Luclpor/url_shortener.git/internal/model"

func TryCreateShortURL(longURL string) (string, bool) {
	key := GenerateRandomString(5)
	urls := model.GetUrls()
	_, ok := urls[key]
	if !ok && containsValue(urls, longURL) {
		urls[key] = longURL
		return urls[key], true
	}
	return urls[key], false
}

func containsValue(urls map[string]string, code string) bool {
	for _, v := range urls {
		if v == code {
			return false
		}
	}
	return true
}
