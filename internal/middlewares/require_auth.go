package middlewares

import (
	"net/http"
	"slices"
)

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		appCtx := GetAppContext(r)
		if appCtx == nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		_, ok := appCtx.SessionManager.GetUser(appCtx)
		if !ok {
			appCtx.Logger.Warn("session not found")
			appCtx.SetJSONError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func RequireAdminAndAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		appCtx := GetAppContext(r)
		if appCtx == nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		user, ok := appCtx.SessionManager.GetUser(appCtx)
		if !ok {
			appCtx.SetJSONError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
			return
		}

		if !slices.Contains(user.Groups, appCtx.Config.Features.MTLSManagement.AdminGroup) {
			appCtx.SetJSONError(http.StatusForbidden, http.StatusText(http.StatusForbidden))
			return
		}

		next.ServeHTTP(w, r)
	})
}
