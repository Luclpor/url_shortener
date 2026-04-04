package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Luclpor/url_shortener.git/internal/config"
	"github.com/Luclpor/url_shortener.git/internal/model"
	modelapi "github.com/Luclpor/url_shortener.git/internal/model/api"
	"github.com/Luclpor/url_shortener.git/internal/service"
	serviceMock "github.com/Luclpor/url_shortener.git/internal/service/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCreateShortenURLJSONHandler(t *testing.T) {
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
			body: `{"url":"https://www.google.com"}`,
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
				response:    "http://localhost:8080/gle",
			},
		},
		{
			name: "successful creation",
			body: `{"url":"https://www.practicum.com"}`,
			want: want{
				code:        http.StatusCreated,
				contentType: "application/json",
			},
		},
	}

	cfg := &config.Config{
		BaseAddressShort: "http://localhost:8080",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockRep := serviceMock.NewMockURLRepository(ctrl)

			var requestBody modelapi.CreateShortenReq
			err := json.Unmarshal([]byte(tt.body), &requestBody)
			require.NoError(t, err)

			if tt.name == "already exist" {
				mockRep.EXPECT().
					FindByLongURL(gomock.Any(), requestBody.URL).
					Return(&model.URL{
						ShortURL: "gle",
						FullUrl:  requestBody.URL,
					}, true)
			} else {
				var generatedShortURL string

				mockRep.EXPECT().
					FindByLongURL(gomock.Any(), requestBody.URL).
					Return(nil, false)
				mockRep.EXPECT().
					FindByShortURL(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ any, shortURL string) (*model.URL, bool) {
						generatedShortURL = shortURL
						return nil, false
					})
				mockRep.EXPECT().
					Save(gomock.Any(), gomock.Any(), requestBody.URL).
					DoAndReturn(func(_ any, shortURL string, fullURL string) (*model.URL, error) {
						assert.Equal(t, generatedShortURL, shortURL)
						return &model.URL{
							ShortURL: shortURL,
							FullUrl:  fullURL,
						}, nil
					})

				tt.want.response = cfg.BaseAddressShort + "/"
			}

			manager := service.NewURLManager(mockRep)
			createHandler := NewCreateShortenUlrJSONHandler(cfg, manager)

			req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			createHandler(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want.code, res.StatusCode)

			ct := res.Header.Get("Content-Type")
			require.NotEmpty(t, ct)
			assert.True(t, strings.HasPrefix(ct, tt.want.contentType), "Content-Type=%q", ct)

			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			var got modelapi.ShortenResp
			err = json.Unmarshal(body, &got)
			require.NoError(t, err)

			if tt.name == "successful creation" {
				assert.True(t, strings.HasPrefix(got.Result, tt.want.response), "result=%q", got.Result)
				assert.Len(t, strings.TrimPrefix(got.Result, tt.want.response), 5)
				return
			}

			assert.Equal(t, tt.want.response, got.Result)
		})
	}
}

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
				response:    "http://localhost:8080/gle",
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

	cfg := &config.Config{
		BaseAddressShort: "http://localhost:8080",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockRep := serviceMock.NewMockURLRepository(ctrl)

			if tt.name == "already exist" {
				mockRep.EXPECT().
					FindByLongURL(gomock.Any(), tt.body).
					Return(&model.URL{
						ShortURL: "gle",
						FullUrl:  tt.body,
					}, true)
			} else {
				var generatedShortURL string

				mockRep.EXPECT().
					FindByLongURL(gomock.Any(), tt.body).
					Return(nil, false)
				mockRep.EXPECT().
					FindByShortURL(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ any, shortURL string) (*model.URL, bool) {
						generatedShortURL = shortURL
						return nil, false
					})
				mockRep.EXPECT().
					Save(gomock.Any(), gomock.Any(), tt.body).
					DoAndReturn(func(_ any, shortURL string, fullURL string) (*model.URL, error) {
						assert.Equal(t, generatedShortURL, shortURL)
						return &model.URL{
							ShortURL: shortURL,
							FullUrl:  fullURL,
						}, nil
					})

				tt.want.response = cfg.BaseAddressShort + "/"
			}

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
			responseBody := strings.TrimSpace(string(b))

			if tt.name == "successful creation" {
				assert.True(t, strings.HasPrefix(responseBody, tt.want.response), "response=%q", responseBody)
				assert.Len(t, strings.TrimPrefix(responseBody, tt.want.response), 5)
				return
			}

			assert.Equal(t, tt.want.response, responseBody)
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
			ctrl := gomock.NewController(t)
			mockRep := serviceMock.NewMockURLRepository(ctrl)

			if tt.want.location != "" {
				mockRep.EXPECT().
					FindByShortURL(gomock.Any(), tt.shortURL).
					Return(&model.URL{
						ShortURL: tt.shortURL,
						FullUrl:  tt.want.location,
					}, true)
			} else {
				mockRep.EXPECT().
					FindByShortURL(gomock.Any(), tt.shortURL).
					Return(nil, false)
			}

			manager := service.NewURLManager(mockRep)
			getHandler := NewGetterHandler(manager)

			mux := http.NewServeMux()
			mux.HandleFunc("GET /{id}", getHandler)

			req := httptest.NewRequest(http.MethodGet, "/"+tt.shortURL, nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want.code, res.StatusCode)

			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Empty(t, body)

			if tt.want.location != "" {
				assert.Equal(t, tt.want.location, res.Header.Get("Location"))
			}
		})
	}
}
