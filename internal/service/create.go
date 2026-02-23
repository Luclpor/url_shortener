package service

import "github.com/Luclpor/url_shortener.git/internal/model"

func TryCreateShortURL(longURL string) (string, bool) {
	urls := model.GetUrls()
	k, b := containsValue(urls, longURL)
	if b {
		return k, false
	}
	key := getUniqueKey(urls)
	urls[key] = longURL
	return key, true
}

func containsValue(urls map[string]string, code string) (string, bool) {
	for k, v := range urls {
		if v == code {
			return k, true
		}
	}
	return "", false
}

func getUniqueKey(urls model.URL) string {
	key := GenerateRandomString(5)
	_, ok := urls[key]
	if ok {
		getUniqueKey(urls)
	}
	return key
}
