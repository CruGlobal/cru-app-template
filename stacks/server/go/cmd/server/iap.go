package main

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/CruGlobal/cru-iap/cruiap"
)

type emailKey struct{}

// requireIAP is the IAP sign-in gate: 401 for all but /up. ECS has no IAP in front,
// so an ECS app must remove or replace it.
func requireIAP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// /up is the health probe: it calls the container directly, with no IAP assertion.
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

// signedInEmail is the email requireIAP verified.
func signedInEmail(r *http.Request) string {
	email, _ := r.Context().Value(emailKey{}).(string)
	return email
}
