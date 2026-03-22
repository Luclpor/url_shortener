package service

import (
	"context"
	"reflect"
	"testing"

	"github.com/Luclpor/url_shortener.git/internal/model"
	mocks "github.com/Luclpor/url_shortener.git/internal/repository/mock"
	"github.com/Luclpor/url_shortener.git/internal/service/mock"
	"go.uber.org/mock/gomock"
)

func TestURLManager_GetURL(t *testing.T) {
	type args struct {
		shortURL string
	}
	tests := []struct {
		name    string
		args    args
		want    *model.URL
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				shortURL: "gle",
			},
			want: &model.URL{
				ShortURL: "gle",
				FullURL:  "https://google.com",
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

			var res *model.URL
			if tt.want != nil {
				res = &model.URL{
					ShortURL: tt.want.ShortURL,
					FullURL:  tt.want.FullURL,
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
	type fields struct {
		repo URLRepository
	}
	type args struct {
		longURL string
	}

	mockRep := mocks.NewURLRepoMock()
	_, _ = mockRep.Save(context.Background(), "gle", "https://google.com")
	_, _ = mockRep.Save(context.Background(), "ya", "http://test2.com")

	tests := []struct {
		name        string
		fields      fields
		args        args
		want        string
		wantCreated bool
		wantErr     bool
	}{
		{
			name: "already exists - returns existing short",
			fields: fields{
				repo: mockRep,
			},
			args: args{
				longURL: "https://google.com",
			},
			want:        "gle",
			wantCreated: false,
			wantErr:     false,
		},
		{
			name: "new url - creates new short",
			fields: fields{
				repo: mockRep,
			},
			args: args{
				longURL: "https://example.com/new",
			},
			want:        "",
			wantCreated: true,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &URLManager{
				repo: tt.fields.repo,
			}

			got, created, err := m.TryCreateShortURL(context.Background(), tt.args.longURL)

			if (err != nil) != tt.wantErr {
				t.Fatalf("TryCreateShortURL() error = %v, wantErr %v", err, tt.wantErr)
			}

			if created != tt.wantCreated {
				t.Fatalf("TryCreateShortURL() created = %v, want %v", created, tt.wantCreated)
			}

			if !tt.wantCreated {
				if got.Result != tt.want {
					t.Fatalf("TryCreateShortURL() got = %q, want %q", got, tt.want)
				}
				return
			}

			if got == nil {
				t.Fatalf("TryCreateShortURL() got is empty, expected generated short")
			}
			if len(got.Result) != 5 {
				t.Fatalf("TryCreateShortURL() got length = %d, want %d (generated key length)", len(got.Result), 5)
			}

			saved, ok := mockRep.FindByShortURL(context.Background(), got.Result)
			if !ok || saved == nil {
				t.Fatalf("expected repo to contain saved short %q", got)
			}
			if saved.FullURL != tt.args.longURL {
				t.Fatalf("saved.FullURL = %q, want %q", saved.FullURL, tt.args.longURL)
			}
		})
	}
}
