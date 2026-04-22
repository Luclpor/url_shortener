package service

import (
	"context"
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
		wantErr error
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
			wantErr: nil,
		},
		{
			name: "not found",
			args: args{
				shortURL: "ru",
			},
			want:    nil,
			wantErr: errors2.ErrNotFound,
		},
		{
			name: "deleted url",
			args: args{
				shortURL: "gone",
			},
			want:    nil,
			wantErr: errors2.ErrURLWasDeleted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRep := mock.NewMockURLRepository(ctrl)

			switch tt.name {
			case "success":
				mockRep.EXPECT().FindByShortURL(gomock.Any(), tt.args.shortURL).
					Return(&model.ShortenURL{
						ShortURL:    tt.want.ShortURL,
						OriginalURL: tt.want.OriginalURL,
					}, true)
			case "deleted url":
				mockRep.EXPECT().FindByShortURL(gomock.Any(), tt.args.shortURL).
					Return(&model.ShortenURL{
						ShortURL:  tt.args.shortURL,
						IsDeleted: true,
					}, true)
			default:
				mockRep.EXPECT().FindByShortURL(gomock.Any(), tt.args.shortURL).
					Return(nil, false)
			}

			manager := &URLManager{
				repo: mockRep,
			}
			got, err := manager.GetURL(context.Background(), tt.args.shortURL)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestURLManager_CreateShortURL(t *testing.T) {
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
			want:         nil,
			wantErr:      nil,
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
				Times(1)

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
					saveResult := &model.ShortenURL{
						ShortURL:    shortURL,
						OriginalURL: tt.args.longURL,
					}
					if tt.alreadyExist {
						saveResult.ShortURL = tt.want.Result
					}
					return saveResult, saveErr
				})

			got, err := m.CreateShortURL(context.Background(), tt.args.longURL, user)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
			if tt.alreadyExist {
				assert.Equal(t, tt.want, got)
				return
			}
			assert.NotNil(t, got)
			assert.Len(t, got.Result, 5)
		})
	}
}
