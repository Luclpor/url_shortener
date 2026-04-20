package service

import (
	"context"
	"reflect"
	"testing"

	"github.com/Luclpor/url_shortener.git/internal/model"
	"github.com/Luclpor/url_shortener.git/internal/model/api"
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

			mockRep.EXPECT().FindByShortURL(gomock.Any(), tt.args.shortURL).
				Return(res, b)

			manager := &URLManager{
				repo: mockRep,
			}
			got, err := manager.GetURL(context.Background(), tt.args.shortURL)
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
	type args struct {
		longURL string
	}

	tests := []struct {
		name         string
		alreadyExist bool
		args         args
		want         *api.ShortenResp
		wantErr      error
	}{
		{
			name: "already exists - returns existing short",
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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRep := mock.NewMockURLRepository(ctrl)
			user := &model.User{ID: uuid.New()}

			m := NewURLManager(mockRep)

			mockRep.EXPECT().
				FindByShortURL(gomock.Any(), gomock.Any()).
				Return(nil, false).
				AnyTimes()

			saveResult := &model.ShortenURL{
				ShortURL:    tt.want.Result,
				OriginalURL: tt.args.longURL,
			}
			saveErr := tt.wantErr
			if !tt.alreadyExist {
				saveErr = nil
			}

			mockRep.EXPECT().
				Save(gomock.Any(), gomock.Any(), tt.args.longURL, user.ID).
				DoAndReturn(func(_ context.Context, shortURL string, longURL string, userID uuid.UUID) (*model.ShortenURL, error) {
					assert.Len(t, shortURL, 5)
					assert.Equal(t, tt.args.longURL, longURL)
					assert.Equal(t, user.ID, userID)
					return saveResult, saveErr
				})

			got, err := m.CreateShortURL(context.Background(), tt.args.longURL, user)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
