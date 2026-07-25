package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Luclpor/url_shortener.git/internal/config"
	"github.com/Luclpor/url_shortener.git/internal/logger"
	"github.com/Luclpor/url_shortener.git/internal/model"
	modelapi "github.com/Luclpor/url_shortener.git/internal/model/api"
	"github.com/Luclpor/url_shortener.git/internal/service"
	"github.com/Luclpor/url_shortener.git/internal/service/audit"
	storageMock "github.com/Luclpor/url_shortener.git/internal/storage/mock"
	"github.com/Luclpor/url_shortener.git/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type testUserAuth struct {
	user *model.User
	err  error
}

func (a testUserAuth) CreateEncryptedUser() (*model.User, string, error) {
	return a.user, "", a.err
}

func (a testUserAuth) DecryptUser(string) (*model.User, error) {
	return a.user, a.err
}

func (a testUserAuth) GetUserFromContext(context.Context) (*model.User, error) {
	return a.user, a.err
}

func (a testUserAuth) SetUserOnContext(ctx context.Context, _ *model.User) context.Context {
	return ctx
}

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
				code:        http.StatusConflict,
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
			mockRep := storageMock.NewMockURLRepository(ctrl)
			user := &model.User{ID: uuid.New()}

			var requestBody modelapi.CreateShortenReq
			err := json.Unmarshal([]byte(tt.body), &requestBody)
			require.NoError(t, err)

			if tt.name == "already exist" {
				gomock.InOrder(
					mockRep.EXPECT().
						FindByShortURL(gomock.Any(), gomock.Any()).
						DoAndReturn(func(_ any, shortURL string) (*model.ShortenURL, bool) {
							return nil, false
						}),
					mockRep.EXPECT().
						Save(gomock.Any(), gomock.Any(), "https://www.google.com", user.ID).
						DoAndReturn(func(_ any, shortURL string, fullURL string, _ any) (*model.ShortenURL, error) {
							return &model.ShortenURL{
								ShortURL:    "gle",
								OriginalURL: "https://www.google.com",
							}, errors.ErrAlreadyExists
						}))
			} else {
				var generatedShortURL string
				mockRep.EXPECT().
					FindByShortURL(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ any, shortURL string) (*model.ShortenURL, bool) {
						generatedShortURL = shortURL
						return nil, false
					})
				mockRep.EXPECT().
					Save(gomock.Any(), gomock.Any(), requestBody.URL, user.ID).
					DoAndReturn(func(_ any, shortURL string, fullURL string, _ any) (*model.ShortenURL, error) {
						assert.Equal(t, generatedShortURL, shortURL)
						return &model.ShortenURL{
							ShortURL:    shortURL,
							OriginalURL: fullURL,
						}, nil
					})

				tt.want.response = cfg.BaseAddressShort + "/"
			}
			appLog, _ := logger.InitLogger(config.ProdEnv)
			manager := service.NewURLManager(mockRep, appLog)
			pub := new(audit.Event)
			handler := NewHandler(cfg, nil, manager, testUserAuth{user: user}, pub, appLog)
			createHandler := handler.NewCreateShortenUlrJSONHandler()

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

func TestCreateShortenURLJSONHandlerWritesAuditEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRep := storageMock.NewMockURLRepository(ctrl)
	user := &model.User{ID: uuid.New()}
	originalURL := "https://www.practicum.com"
	cfg := &config.Config{
		BaseAddressShort: "http://localhost:8080",
	}

	mockRep.EXPECT().
		FindByShortURL(gomock.Any(), gomock.Any()).
		Return(nil, false)
	mockRep.EXPECT().
		Save(gomock.Any(), gomock.Any(), originalURL, user.ID).
		DoAndReturn(func(_ any, shortURL string, fullURL string, _ any) (*model.ShortenURL, error) {
			return &model.ShortenURL{
				ShortURL:    shortURL,
				OriginalURL: fullURL,
			}, nil
		})

	appLog, _ := logger.InitLogger(config.ProdEnv)
	auditPath := filepath.Join(t.TempDir(), "audit.log")
	auditObserver, closeAudit, err := audit.NewStorageAuditor(auditPath)
	require.NoError(t, err)
	require.NotNil(t, closeAudit)
	t.Cleanup(func() {
		require.NoError(t, closeAudit())
	})
	pub := audit.NewEvent(appLog)
	pub.Register(auditObserver)

	manager := service.NewURLManager(mockRep, appLog)
	handler := NewHandler(cfg, nil, manager, testUserAuth{user: user}, pub, appLog)
	createHandler := handler.NewCreateShortenUlrJSONHandler()
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{"url":"`+originalURL+`"}`))
	rec := httptest.NewRecorder()

	createHandler(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	require.Equal(t, http.StatusCreated, res.StatusCode)

	content, err := os.ReadFile(auditPath)
	require.NoError(t, err)
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	require.Len(t, lines, 1)

	var got audit.EventAudit
	require.NoError(t, json.Unmarshal([]byte(lines[0]), &got))
	assert.Equal(t, "shorten", got.Action)
	assert.Equal(t, user.ID, got.UserID)
	assert.Equal(t, originalURL, got.URL)
	assert.NotZero(t, got.Timestamp)
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
				code:        http.StatusConflict,
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
			mockRep := storageMock.NewMockURLRepository(ctrl)
			user := &model.User{ID: uuid.New()}

			if tt.name == "already exist" {

				gomock.InOrder(
					mockRep.EXPECT().
						FindByShortURL(gomock.Any(), gomock.Any()).
						DoAndReturn(func(_ any, shortURL string) (*model.ShortenURL, bool) {
							return nil, false
						}),
					mockRep.EXPECT().
						Save(gomock.Any(), gomock.Any(), tt.body, user.ID).
						DoAndReturn(func(_ any, shortURL string, fullURL string, _ any) (*model.ShortenURL, error) {
							return &model.ShortenURL{
								ShortURL:    "gle",
								OriginalURL: tt.body,
							}, errors.ErrAlreadyExists
						}))
			} else {
				var generatedShortURL string
				mockRep.EXPECT().
					FindByShortURL(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ any, shortURL string) (*model.ShortenURL, bool) {
						generatedShortURL = shortURL
						return nil, false
					})
				mockRep.EXPECT().
					Save(gomock.Any(), gomock.Any(), tt.body, user.ID).
					DoAndReturn(func(_ any, shortURL string, fullURL string, _ any) (*model.ShortenURL, error) {
						assert.Equal(t, generatedShortURL, shortURL)
						return &model.ShortenURL{
							ShortURL:    shortURL,
							OriginalURL: fullURL,
						}, nil
					})

				tt.want.response = cfg.BaseAddressShort + "/"
			}
			appLog, _ := logger.InitLogger(config.ProdEnv)
			manager := service.NewURLManager(mockRep, appLog)
			pub := new(audit.Event)
			handler := NewHandler(cfg, nil, manager, testUserAuth{user: user}, pub, appLog)
			createHandler := handler.NewCreateHandler()

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
				code: http.StatusNoContent,
			},
		},
		{
			name:     "deleted url",
			shortURL: "gone",
			want: want{
				code: http.StatusGone,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &model.User{ID: uuid.New()}

			ctrl := gomock.NewController(t)
			mockRep := storageMock.NewMockURLRepository(ctrl)

			if tt.want.location != "" {
				mockRep.EXPECT().
					FindByShortURL(gomock.Any(), tt.shortURL).
					Return(&model.ShortenURL{
						ShortURL:    tt.shortURL,
						OriginalURL: tt.want.location,
					}, true)
			} else if tt.name == "deleted url" {
				mockRep.EXPECT().
					FindByShortURL(gomock.Any(), tt.shortURL).
					Return(&model.ShortenURL{
						ShortURL:  tt.shortURL,
						IsDeleted: true,
					}, true)
			} else {
				mockRep.EXPECT().
					FindByShortURL(gomock.Any(), tt.shortURL).
					Return(nil, false)
			}

			appLog, _ := logger.InitLogger(config.ProdEnv)
			manager := service.NewURLManager(mockRep, appLog)
			pub := new(audit.Event)
			handler := NewHandler(nil, nil, manager, testUserAuth{user: user}, pub, appLog)
			getHandler := handler.NewGetterHandler()

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

func TestStatsHandler(t *testing.T) {
	tests := []struct {
		name          string
		trustedSubnet string
		realIP        string
		wantCode      int
		wantStats     *modelapi.StatsResp
	}{
		{
			name:          "allowed trusted ip",
			trustedSubnet: "192.168.1.0/24",
			realIP:        "192.168.1.42",
			wantCode:      http.StatusOK,
			wantStats: &modelapi.StatsResp{
				URLs:  7,
				Users: 3,
			},
		},
		{
			name:          "empty trusted subnet forbids",
			trustedSubnet: "",
			realIP:        "192.168.1.42",
			wantCode:      http.StatusForbidden,
		},
		{
			name:          "outside trusted subnet forbids",
			trustedSubnet: "192.168.1.0/24",
			realIP:        "192.168.2.42",
			wantCode:      http.StatusForbidden,
		},
		{
			name:          "invalid real ip forbids",
			trustedSubnet: "192.168.1.0/24",
			realIP:        "not-ip",
			wantCode:      http.StatusForbidden,
		},
		{
			name:          "invalid trusted subnet forbids",
			trustedSubnet: "not-cidr",
			realIP:        "192.168.1.42",
			wantCode:      http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockRep := storageMock.NewMockURLRepository(ctrl)
			if tt.wantStats != nil {
				mockRep.EXPECT().
					GetStats(gomock.Any()).
					Return(tt.wantStats.URLs, tt.wantStats.Users, nil)
			}

			appLog, _ := logger.InitLogger(config.ProdEnv)
			manager := service.NewURLManager(mockRep, appLog)
			handler := NewHandler(&config.Config{
				HTTPServer: config.HTTPServer{
					TrustedSubnet: tt.trustedSubnet,
				},
			}, nil, manager, testUserAuth{}, new(audit.Event), appLog)
			statsHandler := handler.NewStatsHandler()

			req := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
			req.Header.Set("X-Real-IP", tt.realIP)
			rec := httptest.NewRecorder()

			statsHandler(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.wantCode, res.StatusCode)
			if tt.wantStats == nil {
				return
			}

			ct := res.Header.Get("Content-Type")
			require.NotEmpty(t, ct)
			assert.True(t, strings.HasPrefix(ct, "application/json"), "Content-Type=%q", ct)

			var got modelapi.StatsResp
			require.NoError(t, json.NewDecoder(res.Body).Decode(&got))
			assert.Equal(t, *tt.wantStats, got)
		})
	}
}
