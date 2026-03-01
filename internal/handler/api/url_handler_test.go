package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Luclpor/url_shortener.git/internal/config"
	mocks "github.com/Luclpor/url_shortener.git/internal/repository/mock"
	"github.com/Luclpor/url_shortener.git/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreatedShortURL(t *testing.T) {
	type want struct {
		code        int
		contentType string
		response    string
	}

	tests := []struct {
		name string
		body string
		want want
	}{
		{
			name: "already exist",
			body: "https://www.google.com",
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain",
				response:    "localhost:8080/gle",
			},
		},

		{
			name: "successful creation",
			body: "https://www.practicum.com",
			want: want{
				code:        http.StatusCreated,
				contentType: "text/plain",
			},
		},
	}

	cfg := config.InitConfig()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRep := mocks.NewURLRepoMock()
			_, _ = mockRep.Save("gle", "https://www.google.com")

			manager := service.NewURLManager(mockRep)
			createHandler := NewCreateHandler(cfg, manager)

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			createHandler(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want.code, res.StatusCode)

			ct := res.Header.Get("Content-Type")
			require.NotEmpty(t, ct)
			assert.True(t, strings.HasPrefix(ct, tt.want.contentType), "Content-Type=%q", ct)

			b, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			if tt.want.response != "" {
				assert.Equal(t, tt.want.response, strings.TrimSpace(string(b)))
			}
		})
	}
}

func TestGetShortURL(t *testing.T) {
	type want struct {
		code     int
		location string
	}

	tests := []struct {
		name     string
		shortURL string
		want     want
	}{
		{
			name:     "redirects to full url",
			shortURL: "gle",
			want: want{
				code:     http.StatusTemporaryRedirect,
				location: "https://www.google.com",
			},
		},
		{
			name:     "not found",
			shortURL: "notExist",
			want: want{
				code: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// полностью детерминируем: кладём запись заранее, чтобы не зависеть от генератора
			mockRep := mocks.NewURLRepoMock()
			_, _ = mockRep.Save("gle", "https://www.google.com")

			manager := service.NewURLManager(mockRep)
			getHandler := NewGetterHandler(manager)

			mux := http.NewServeMux()
			mux.HandleFunc("GET /{id}", getHandler)

			// важно: путь должен быть "/gle", чтобы mux положил PathValue("id") = "gle"
			req := httptest.NewRequest(http.MethodGet, "/"+tt.shortURL, nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want.code, res.StatusCode)

			// у редиректа тело обычно пустое
			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Empty(t, body)

			if tt.want.location != "" {
				assert.Equal(t, tt.want.location, res.Header.Get("Location"))
			}
		})
	}
}
