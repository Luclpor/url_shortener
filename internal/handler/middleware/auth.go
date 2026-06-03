package middleware

import (
	"net/http"

	"github.com/Luclpor/url_shortener.git/internal/model"
	"github.com/Luclpor/url_shortener.git/internal/service/auth"
	"go.uber.org/zap"
)

func Auth(userAuth auth.UserAuthentication, appLogger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var user *model.User

			appLogger.Info("incoming cookies",
				zap.String("cookie_header", r.Header.Get("Cookie")),
				zap.Int("cookies_count", len(r.Cookies())))

			for _, cookie := range r.Cookies() {
				appLogger.Info("cookie",
					zap.String("name", cookie.Name),
					zap.String("value", cookie.Value))
			}
			cookie, err := r.Cookie("user_id")
			if err != nil {
				appLogger.Info("cookie not found in request")
				newUser, encryptedUserID, err := userAuth.CreateEncryptedUser()
				if err != nil {
					appLogger.Error("error creating new user", zap.Error(err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				user = newUser

				http.SetCookie(w, &http.Cookie{
					Name:  "user_id",
					Value: encryptedUserID,
				})
			} else {
				appLogger.Info("cookie found in request")
				existingUser, err := userAuth.DecryptUser(cookie.Value)
				if err != nil {
					appLogger.Error("error decrypting user", zap.Error(err))
					http.Error(w, "invalid user cookie", http.StatusUnauthorized)
					return
				}

				user = existingUser
			}

			ctx := userAuth.SetUserOnContext(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
