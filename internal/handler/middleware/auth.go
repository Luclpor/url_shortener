package middleware

import (
	"net/http"

	"github.com/Luclpor/url_shortener.git/internal/logger"
	"github.com/Luclpor/url_shortener.git/internal/model"
	"github.com/Luclpor/url_shortener.git/internal/service/auth"
)

func Auth(userAuth auth.UserAuthentication) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var user *model.User

			for _, cookie := range r.Cookies() {
				logger.SugarLogger.Infow("cookie",
					"name", cookie.Name,
					"value", cookie.Value,
				)
			}
			cookie, err := r.Cookie("user_id")
			if err != nil {
				logger.SugarLogger.Info("cookie not found")
				newUser, encryptedUserID, err := userAuth.CreateEncryptedUser()
				if err != nil {
					http.Error(w, "failed to create user", http.StatusInternalServerError)
					return
				}

				user = newUser

				http.SetCookie(w, &http.Cookie{
					Name:     "user_id",
					Value:    encryptedUserID,
					Path:     "/",
					HttpOnly: true,
					SameSite: http.SameSiteLaxMode,
				})
			} else {
				logger.SugarLogger.Info("cookie found")
				existingUser, err := userAuth.DecryptUser(cookie.Value)
				if err != nil {
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
