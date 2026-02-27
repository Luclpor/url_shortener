package handler

import (
	"fmt"
	"strings"

	"github.com/Luclpor/url_shortener.git/internal/config"
	"github.com/Luclpor/url_shortener.git/internal/repository"
	"github.com/Luclpor/url_shortener.git/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func init() {
	_ = config.InitConfig()
}

func TestCreatedShortURL(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name string
		body string
		want want
	}{
		{
			name: "positive test created short url",
			body: "https://practicum.yandex.ru/",
			want: want{
				code:        http.StatusCreated,
				contentType: "text/plain",
				response:    `{"http://localhost:8080/"}`,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "http://localhost:8080/", strings.NewReader("https://www.google.com"))
			w := httptest.NewRecorder()
			repo := repository.NewRepository()
			manger := service.NewURLManager(repo)
			createHandler := NewCreateHandler(config.InitConfig(), manger)
			createHandler(w, request)
			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.NotEmpty(t, resBody)
			fmt.Println(string(resBody))
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func TestGetShortURL(t *testing.T) {
	type want struct {
		code     int
		Header   http.Header
		response string
	}
	tests := []struct {
		name    string
		longURL string
		want    want
	}{
		{
			name:    "positive test created short url",
			longURL: "https://www.google.com",
			want: want{
				code:     http.StatusTemporaryRedirect,
				response: "https://www.google.com",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mux := http.NewServeMux()
			repo := repository.NewRepository()
			manger := service.NewURLManager(repo)
			createHandler := NewCreateHandler(config.InitConfig(), manger)
			getHandler := NewGetterHandler(manger)
			mux.HandleFunc("POST /", createHandler)
			mux.HandleFunc("GET /{id}", getHandler)

			postReq := httptest.NewRequest(http.MethodPost, "http://localhost:8080/", strings.NewReader(test.longURL))
			postRec := httptest.NewRecorder()

			mux.ServeHTTP(postRec, postReq)

			postRes := postRec.Result()
			defer postRes.Body.Close()
			result, err := io.ReadAll(postRes.Body)
			fmt.Println("Created short ulr: ", string(result))
			require.NoError(t, err)

			getReq := httptest.NewRequest(http.MethodGet, string(result), nil)

			getRec := httptest.NewRecorder()
			mux.ServeHTTP(getRec, getReq)

			getRes := getRec.Result()
			assert.Equal(t, test.want.code, getRes.StatusCode)
			defer getRes.Body.Close()
			resBody, err := io.ReadAll(getRes.Body)
			fmt.Println("Get full url:, ", test.want.response)
			require.NoError(t, err)
			assert.Empty(t, resBody)
			assert.Equal(t, test.want.response, getRes.Header.Get("Location"))
		})
	}
}
