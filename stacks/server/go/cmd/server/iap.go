package main

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/CruGlobal/cru-iap/cruiap"
)

type emailKey struct{}

// requireIAP is the sign-in gate. Google IAP signs people in with Okta before a
// request gets here, and the platform sets IAP_AUDIENCE. Every route but /up
// needs IAP's signed assertion; anything else is a 401, so an app with no IAP in
// front (ECS) must remove or replace this gate. Locally, CRU_IAP_DEV_BYPASS_EMAIL
// (from .env.development) signs you in as that email; cru-iap ignores it when
// IAP_AUDIENCE is set or on Cloud Run.
// Sign out with /?gcp-iap-mode=CLEAR_LOGIN_COOKIE (cruiap.LogoutURL).
func requireIAP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// /up is the health check, and the load balancer lets it skip IAP.
		if r.URL.Path == "/up" {
			next.ServeHTTP(w, r)
			return
		}

		result, bypassed := cruiap.DevBypass()
		if !bypassed {
			result = cruiap.VerifyRequest(r.Context(), r, cruiap.WithLogger(slog.Default()))
		}
		if !result.OK {
			slog.Warn("iap_rejected", "reason", result.Reason, "path", r.URL.Path)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), emailKey{}, result.Email)))
	})
}

// signedInEmail is the email requireIAP verified, or "" on /up.
func signedInEmail(r *http.Request) string {
	email, _ := r.Context().Value(emailKey{}).(string)
	return email
}
