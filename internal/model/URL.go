package model

type URL map[string]string

var urls URL

func GetUrls() URL {
	if urls == nil {
		urls = make(URL)
	}
	return urls
}
