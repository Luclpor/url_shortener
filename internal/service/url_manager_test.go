package service

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/Luclpor/url_shortener.git/internal/model"
	"github.com/Luclpor/url_shortener.git/internal/model/api"
	"github.com/Luclpor/url_shortener.git/internal/service/auth"
	"github.com/Luclpor/url_shortener.git/internal/storage/mock"
	errors2 "github.com/Luclpor/url_shortener.git/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestURLManager_GetURL(t *testing.T) {
	type args struct {
		shortURL string
	}
	tests := []struct {
		name    string
		args    args
		want    *model.ShortenURL
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				shortURL: "gle",
			},
			want: &model.ShortenURL{
				ShortURL:    "gle",
				OriginalURL: "https://google.com",
			},
			wantErr: false,
		},
		{
			name: "failure",
			args: args{
				shortURL: "ru",
			},
			want:    nil,
			wantErr: true,
		},
	}

	ctx := auth.WithUser(context.Background(), &model.User{ID: uuid.New()})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRep := mock.NewMockURLRepository(ctrl)

			var res *model.ShortenURL
			if tt.want != nil {
				res = &model.ShortenURL{
					ShortURL:    tt.want.ShortURL,
					OriginalURL: tt.want.OriginalURL,
				}
			}

			var b = false
			if res != nil {
				b = true
			}

			mockRep.EXPECT().FindByShortURL(gomock.Any(), tt.args.shortURL, gomock.Any()).
				Return(res, b)

			manager := &URLManager{
				repo: mockRep,
			}
			got, err := manager.GetURL(ctx, tt.args.shortURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetURL() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestURLManager_TryCreateShortURL(t *testing.T) {
	type fields struct {
		repo URLRepository
	}
	type args struct {
		shortURL string
		longURL  string
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRep := mock.NewMockURLRepository(ctrl)

	//mockRep := mocks.NewURLRepoMock()
	//_, _ = mockRep.Save(context.Background(), "gle", "https://google.com")
	//_, _ = mockRep.Save(context.Background(), "ya", "http://test2.com")

	tests := []struct {
		name         string
		fields       fields
		alreadyExist bool
		args         args
		want         *api.ShortenResp
		wantErr      error
	}{
		{
			name: "already exists - returns existing short",
			fields: fields{
				repo: mockRep,
			},
			args: args{
				longURL: "https://google.com",
			},
			alreadyExist: true,
			want: &api.ShortenResp{
				Result: "gle",
			},
			wantErr: errors2.ErrAlreadyExists,
		},
		{
			name: "new url - creates new short",
			fields: fields{
				repo: mockRep,
			},

			args: args{
				longURL: "https://example.com/new",
			},
			alreadyExist: false,
			want: &api.ShortenResp{
				Result: "shortik,",
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &URLManager{
				repo: tt.fields.repo,
			}
			mockRep.EXPECT().
				FindByShortURL(gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, shortURL string, _ any) (*model.ShortenURL, bool) {
					if shortURL == tt.want.Result {
						return &model.ShortenURL{
							ShortURL:    tt.want.Result,
							OriginalURL: tt.args.longURL,
						}, true
					}
					return nil, false
				}).
				AnyTimes()

			if tt.alreadyExist {
				mockRep.EXPECT().
					Save(gomock.Any(), gomock.Any(), tt.args.longURL, gomock.Any()).
					Return(&model.ShortenURL{
						ShortURL:    tt.want.Result,
						OriginalURL: tt.args.longURL,
					}, errors2.ErrAlreadyExists)
			} else {
				mockRep.EXPECT().
					Save(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, shortURL string, longURL string, _ any) (*model.ShortenURL, error) {
						if longURL == tt.args.longURL {
							return &model.ShortenURL{
								ShortURL:    tt.want.Result,
								OriginalURL: tt.args.longURL,
							}, nil
						}
						return nil, fmt.Errorf("error")
					})
			}

			got, err := m.CreateShortURL(context.Background(), tt.args.longURL)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, got, tt.want)
		})
	}
}
