package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/CruGlobal/cru-iap/cruiap"
)

type emailKey struct{}

// requireIAP is the sign-in gate. On Cloud Run, Google IAP signs people in with
// Okta before a request gets here, and the platform sets IAP_AUDIENCE: then
// every route but /up needs IAP's signed assertion, and anything else is a 401.
// Without IAP_AUDIENCE (ECS, or local dev) there is no gate;
// CRU_IAP_DEV_BYPASS_EMAIL=you@cru.org gives you a signed-in email locally.
// Sign out with /?gcp-iap-mode=CLEAR_LOGIN_COOKIE (cruiap.LogoutURL).
func requireIAP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// /up is the health check, and the load balancer lets it skip IAP.
		if r.URL.Path == "/up" {
			next.ServeHTTP(w, r)
			return
		}

		// Gate on deploy config, never on the header being absent.
		if os.Getenv("IAP_AUDIENCE") == "" {
			if dev, ok := cruiap.DevBypass(); ok {
				r = r.WithContext(context.WithValue(r.Context(), emailKey{}, dev.Email))
			}
			next.ServeHTTP(w, r)
			return
		}

		result := cruiap.VerifyRequest(r.Context(), r, cruiap.WithLogger(slog.Default()))
		if !result.OK {
			// Fail closed: never fall back to a dev identity here.
			slog.Warn("iap_rejected", "reason", result.Reason, "path", r.URL.Path)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), emailKey{}, result.Email)))
	})
}

// signedInEmail is the email requireIAP verified, or "" when there is no gate.
func signedInEmail(r *http.Request) string {
	email, _ := r.Context().Value(emailKey{}).(string)
	return email
}
