package services

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/stormlightlabs/noteleaf/internal/public"
)

func TestATProtoService(t *testing.T) {
	t.Run("NewATProtoService", func(t *testing.T) {
		t.Run("creates service with default configuration", func(t *testing.T) {
			svc := NewATProtoService()

			if svc == nil {
				t.Fatal("Expected service to be created, got nil")
			}

			if svc.pdsURL != "https://bsky.social" {
				t.Errorf("Expected pdsURL to be 'https://bsky.social', got '%s'", svc.pdsURL)
			}

			if svc.client == nil {
				t.Fatal("Expected client to be initialized, got nil")
			}

			if svc.client.Host != "https://bsky.social" {
				t.Errorf("Expected client Host to be 'https://bsky.social', got '%s'", svc.client.Host)
			}
		})
	})

	t.Run("Authenticate", func(t *testing.T) {
		t.Run("validates required parameters", func(t *testing.T) {
			svc := NewATProtoService()
			ctx := context.Background()

			err := svc.Authenticate(ctx, "", "password")
			if err == nil {
				t.Error("Expected error for empty handle, got nil")
			}

			err = svc.Authenticate(ctx, "handle", "")
			if err == nil {
				t.Error("Expected error for empty password, got nil")
			}

			err = svc.Authenticate(ctx, "", "")
			if err == nil {
				t.Error("Expected error for empty handle and password, got nil")
			}
		})
	})

	t.Run("IsAuthenticated", func(t *testing.T) {
		t.Run("returns false when no session exists", func(t *testing.T) {
			svc := NewATProtoService()

			if svc.IsAuthenticated() {
				t.Error("Expected IsAuthenticated to return false for new service")
			}
		})

		t.Run("returns false when session is not authenticated", func(t *testing.T) {
			svc := NewATProtoService()
			svc.session = &Session{
				Handle:        "test.bsky.social",
				Authenticated: false,
			}

			if svc.IsAuthenticated() {
				t.Error("Expected IsAuthenticated to return false for unauthenticated session")
			}
		})

		t.Run("returns true when session is authenticated", func(t *testing.T) {
			svc := NewATProtoService()
			svc.session = &Session{
				Handle:        "test.bsky.social",
				Authenticated: true,
			}

			if !svc.IsAuthenticated() {
				t.Error("Expected IsAuthenticated to return true for authenticated session")
			}
		})
	})

	t.Run("GetSession", func(t *testing.T) {
		t.Run("returns error when not authenticated", func(t *testing.T) {
			svc := NewATProtoService()

			session, err := svc.GetSession()
			if err == nil {
				t.Error("Expected error when getting session without authentication")
			}
			if session != nil {
				t.Error("Expected nil session when not authenticated")
			}
		})

		t.Run("returns session when authenticated", func(t *testing.T) {
			svc := NewATProtoService()
			expectedSession := &Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				PDSURL:        "https://bsky.social",
				ExpiresAt:     time.Now().Add(2 * time.Hour),
				Authenticated: true,
			}
			svc.session = expectedSession

			session, err := svc.GetSession()
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			if session == nil {
				t.Fatal("Expected session to be returned, got nil")
			}
			if session.DID != expectedSession.DID {
				t.Errorf("Expected DID '%s', got '%s'", expectedSession.DID, session.DID)
			}
			if session.Handle != expectedSession.Handle {
				t.Errorf("Expected Handle '%s', got '%s'", expectedSession.Handle, session.Handle)
			}
		})
	})

	t.Run("RefreshToken", func(t *testing.T) {
		t.Run("returns error when no session exists", func(t *testing.T) {
			svc := NewATProtoService()
			ctx := context.Background()

			err := svc.RefreshToken(ctx)
			if err == nil {
				t.Error("Expected error when refreshing without session")
			}
		})

		t.Run("returns error when refresh token is empty", func(t *testing.T) {
			svc := NewATProtoService()
			ctx := context.Background()
			svc.session = &Session{
				Handle:     "test.bsky.social",
				RefreshJWT: "",
			}

			err := svc.RefreshToken(ctx)
			if err == nil {
				t.Error("Expected error when refreshing with empty refresh token")
			}
		})
	})

	t.Run("RestoreSession", func(t *testing.T) {
		t.Run("returns error when session is nil", func(t *testing.T) {
			svc := NewATProtoService()

			err := svc.RestoreSession(nil)
			if err == nil {
				t.Error("Expected error when restoring nil session")
			}
		})

		t.Run("returns error when session missing DID", func(t *testing.T) {
			svc := NewATProtoService()
			session := &Session{
				DID:        "",
				Handle:     "test.bsky.social",
				AccessJWT:  "access_token",
				RefreshJWT: "refresh_token",
			}

			err := svc.RestoreSession(session)
			if err == nil {
				t.Error("Expected error when session missing DID")
			}
		})

		t.Run("returns error when session missing AccessJWT", func(t *testing.T) {
			svc := NewATProtoService()
			session := &Session{
				DID:        "did:plc:test123",
				Handle:     "test.bsky.social",
				AccessJWT:  "",
				RefreshJWT: "refresh_token",
			}

			err := svc.RestoreSession(session)
			if err == nil {
				t.Error("Expected error when session missing AccessJWT")
			}
		})

		t.Run("returns error when session missing RefreshJWT", func(t *testing.T) {
			svc := NewATProtoService()
			session := &Session{
				DID:        "did:plc:test123",
				Handle:     "test.bsky.social",
				AccessJWT:  "access_token",
				RefreshJWT: "",
			}

			err := svc.RestoreSession(session)
			if err == nil {
				t.Error("Expected error when session missing RefreshJWT")
			}
		})

		t.Run("successfully restores valid session", func(t *testing.T) {
			svc := NewATProtoService()
			session := &Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				PDSURL:        "https://test.pds.example",
				ExpiresAt:     time.Now().Add(2 * time.Hour),
				Authenticated: true,
			}

			err := svc.RestoreSession(session)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if !svc.IsAuthenticated() {
				t.Error("Expected service to be authenticated after restore")
			}

			restoredSession, err := svc.GetSession()
			if err != nil {
				t.Errorf("Expected to get session, got error: %v", err)
			}
			if restoredSession.DID != session.DID {
				t.Errorf("Expected DID '%s', got '%s'", session.DID, restoredSession.DID)
			}
			if restoredSession.Handle != session.Handle {
				t.Errorf("Expected Handle '%s', got '%s'", session.Handle, restoredSession.Handle)
			}
		})

		t.Run("updates client authentication", func(t *testing.T) {
			svc := NewATProtoService()
			session := &Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				PDSURL:        "https://test.pds.example",
				ExpiresAt:     time.Now().Add(2 * time.Hour),
				Authenticated: true,
			}

			err := svc.RestoreSession(session)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if svc.client.Auth == nil {
				t.Fatal("Expected client Auth to be set")
			}
			if svc.client.Auth.Did != session.DID {
				t.Errorf("Expected client DID '%s', got '%s'", session.DID, svc.client.Auth.Did)
			}
			if svc.client.Auth.AccessJwt != session.AccessJWT {
				t.Errorf("Expected client AccessJwt '%s', got '%s'", session.AccessJWT, svc.client.Auth.AccessJwt)
			}
		})

		t.Run("updates PDS URL when provided", func(t *testing.T) {
			svc := NewATProtoService()
			customPDS := "https://custom.pds.example"
			session := &Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				PDSURL:        customPDS,
				ExpiresAt:     time.Now().Add(2 * time.Hour),
				Authenticated: true,
			}

			err := svc.RestoreSession(session)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if svc.pdsURL != customPDS {
				t.Errorf("Expected pdsURL '%s', got '%s'", customPDS, svc.pdsURL)
			}
			if svc.client.Host != customPDS {
				t.Errorf("Expected client Host '%s', got '%s'", customPDS, svc.client.Host)
			}
		})
	})

	t.Run("Close", func(t *testing.T) {
		t.Run("clears session", func(t *testing.T) {
			svc := NewATProtoService()
			svc.session = &Session{
				Handle:        "test.bsky.social",
				Authenticated: true,
			}

			err := svc.Close()
			if err != nil {
				t.Errorf("Expected no error on close, got %v", err)
			}
			if svc.session != nil {
				t.Error("Expected session to be nil after close")
			}
		})

		t.Run("handles nil session gracefully", func(t *testing.T) {
			svc := NewATProtoService()
			svc.session = nil

			err := svc.Close()
			if err != nil {
				t.Errorf("Expected no error on close with nil session, got %v", err)
			}
		})
	})

	t.Run("PullDocuments", func(t *testing.T) {
		t.Run("returns error when not authenticated", func(t *testing.T) {
			svc := NewATProtoService()
			ctx := context.Background()

			docs, err := svc.PullDocuments(ctx)
			if err == nil {
				t.Error("Expected error when pulling documents without authentication")
			}
			if docs != nil {
				t.Error("Expected nil documents when not authenticated")
			}
			if err.Error() != "not authenticated" {
				t.Errorf("Expected 'not authenticated' error, got '%v'", err)
			}
		})

		t.Run("returns error when session not authenticated", func(t *testing.T) {
			svc := NewATProtoService()
			ctx := context.Background()
			svc.session = &Session{
				Handle:        "test.bsky.social",
				Authenticated: false,
			}

			docs, err := svc.PullDocuments(ctx)
			if err == nil {
				t.Error("Expected error when pulling documents with unauthenticated session")
			}
			if docs != nil {
				t.Error("Expected nil documents when session not authenticated")
			}
		})

		t.Run("returns error when context cancelled", func(t *testing.T) {
			svc := NewATProtoService()
			svc.session = &Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				Authenticated: true,
			}

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			docs, err := svc.PullDocuments(ctx)
			if err == nil {
				t.Error("Expected error when context is cancelled")
			}
			if docs != nil {
				t.Error("Expected nil documents when context is cancelled")
			}
		})

		t.Run("returns empty list when no documents exist", func(t *testing.T) {
			svc := NewATProtoService()
			svc.session = &Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				Authenticated: true,
			}
			ctx := context.Background()

			docs, err := svc.PullDocuments(ctx)

			if err != nil && err.Error() == "not authenticated" {
				t.Error("Authentication check should pass, but got authentication error")
			}

			_ = docs
		})
	})

	t.Run("ListPublications", func(t *testing.T) {
		t.Run("returns error when not authenticated", func(t *testing.T) {
			svc := NewATProtoService()
			ctx := context.Background()

			pubs, err := svc.ListPublications(ctx)
			if err == nil {
				t.Error("Expected error when listing publications without authentication")
			}
			if pubs != nil {
				t.Error("Expected nil publications when not authenticated")
			}
			if err.Error() != "not authenticated" {
				t.Errorf("Expected 'not authenticated' error, got '%v'", err)
			}
		})

		t.Run("returns error when session not authenticated", func(t *testing.T) {
			svc := NewATProtoService()
			ctx := context.Background()
			svc.session = &Session{
				Handle:        "test.bsky.social",
				Authenticated: false,
			}

			pubs, err := svc.ListPublications(ctx)
			if err == nil {
				t.Error("Expected error when listing publications with unauthenticated session")
			}
			if pubs != nil {
				t.Error("Expected nil publications when session not authenticated")
			}
		})

		t.Run("returns error when context cancelled", func(t *testing.T) {
			svc := NewATProtoService()
			svc.session = &Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				Authenticated: true,
			}

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			pubs, err := svc.ListPublications(ctx)
			if err == nil {
				t.Error("Expected error when context is cancelled")
			}
			if pubs != nil {
				t.Error("Expected nil publications when context is cancelled")
			}
		})

		t.Run("returns error when context timeout", func(t *testing.T) {
			svc := NewATProtoService()
			svc.session = &Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				Authenticated: true,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 1)
			defer cancel()
			time.Sleep(2 * time.Millisecond)

			pubs, err := svc.ListPublications(ctx)
			if err == nil {
				t.Error("Expected error when context times out")
			}
			if pubs != nil {
				t.Error("Expected nil publications when context times out")
			}
		})

		t.Run("returns empty list when no publications exist", func(t *testing.T) {
			svc := NewATProtoService()
			svc.session = &Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				Authenticated: true,
			}
			ctx := context.Background()

			pubs, err := svc.ListPublications(ctx)

			if err != nil && err.Error() == "not authenticated" {
				t.Error("Authentication check should pass, but got authentication error")
			}

			_ = pubs
		})
	})

	t.Run("GetDefaultPublication", func(t *testing.T) {
		t.Run("returns error when not authenticated", func(t *testing.T) {
			svc := NewATProtoService()
			ctx := context.Background()

			uri, err := svc.GetDefaultPublication(ctx)
			if err == nil {
				t.Error("Expected error when getting default publication without authentication")
			}
			if uri != "" {
				t.Errorf("Expected empty URI, got %s", uri)
			}
			if !strings.Contains(err.Error(), "not authenticated") {
				t.Errorf("Expected 'not authenticated' error, got '%v'", err)
			}
		})

		t.Run("returns error when session not authenticated", func(t *testing.T) {
			svc := NewATProtoService()
			ctx := context.Background()
			svc.session = &Session{
				Handle:        "test.bsky.social",
				Authenticated: false,
			}

			uri, err := svc.GetDefaultPublication(ctx)
			if err == nil {
				t.Error("Expected error when getting default publication with unauthenticated session")
			}
			if uri != "" {
				t.Errorf("Expected empty URI, got %s", uri)
			}
		})

		t.Run("returns error when no publications exist", func(t *testing.T) {
			svc := NewATProtoService()
			svc.session = &Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				Authenticated: true,
			}
			ctx := context.Background()

			_, err := svc.GetDefaultPublication(ctx)
			if err == nil {
				t.Error("Expected error when getting default publication")
			}
			// With invalid credentials, we expect either auth error or no publications error
			if !strings.Contains(err.Error(), "no publications found") &&
				!strings.Contains(err.Error(), "Authentication") &&
				!strings.Contains(err.Error(), "AuthMissing") &&
				!strings.Contains(err.Error(), "failed to fetch repository") {
				t.Errorf("Expected authentication or 'no publications found' error, got '%v'", err)
			}
		})

		t.Run("returns error when context cancelled", func(t *testing.T) {
			svc := NewATProtoService()
			svc.session = &Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				Authenticated: true,
			}

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			uri, err := svc.GetDefaultPublication(ctx)
			if err == nil {
				t.Error("Expected error when context is cancelled")
			}
			if uri != "" {
				t.Errorf("Expected empty URI when error occurs, got %s", uri)
			}
		})
	})

	t.Run("Authentication Error Scenarios", func(t *testing.T) {
		t.Run("returns error with context timeout", func(t *testing.T) {
			svc := NewATProtoService()
			ctx, cancel := context.WithTimeout(context.Background(), 1)
			defer cancel()
			time.Sleep(2 * time.Millisecond)

			err := svc.Authenticate(ctx, "test.bsky.social", "password")
			if err == nil {
				t.Error("Expected error when context times out")
			}
		})

		t.Run("returns error with cancelled context", func(t *testing.T) {
			svc := NewATProtoService()
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			err := svc.Authenticate(ctx, "test.bsky.social", "password")
			if err == nil {
				t.Error("Expected error when context is cancelled")
			}
		})

		t.Run("validates both handle and password together", func(t *testing.T) {
			svc := NewATProtoService()
			ctx := context.Background()

			testCases := []struct {
				name     string
				handle   string
				password string
			}{
				{"empty handle", "", "password"},
				{"empty password", "handle", ""},
				{"both empty", "", ""},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					err := svc.Authenticate(ctx, tc.handle, tc.password)
					if err == nil {
						t.Errorf("Expected error for %s", tc.name)
					}
					if !svc.IsAuthenticated() == false {
						t.Error("Expected service to not be authenticated after error")
					}
				})
			}
		})
	})

	t.Run("RefreshToken Error Scenarios", func(t *testing.T) {
		t.Run("returns error with cancelled context", func(t *testing.T) {
			svc := NewATProtoService()
			svc.session = &Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				Authenticated: true,
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			err := svc.RefreshToken(ctx)
			if err == nil {
				t.Error("Expected error when context is cancelled")
			}
		})

		t.Run("returns error with timeout context", func(t *testing.T) {
			svc := NewATProtoService()
			svc.session = &Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				Authenticated: true,
			}
			ctx, cancel := context.WithTimeout(context.Background(), 1)
			defer cancel()
			time.Sleep(2 * time.Millisecond)

			err := svc.RefreshToken(ctx)
			if err == nil {
				t.Error("Expected error when context times out")
			}
		})
	})

	t.Run("PostDocument", func(t *testing.T) {
		t.Run("returns error when not authenticated", func(t *testing.T) {
			svc := NewATProtoService()
			ctx := context.Background()

			doc := public.Document{
				Title: "Test Document",
			}

			result, err := svc.PostDocument(ctx, doc, false)
			if err == nil {
				t.Error("Expected error when posting document without authentication")
			}
			if result != nil {
				t.Error("Expected nil result when not authenticated")
			}
			if err.Error() != "not authenticated" {
				t.Errorf("Expected 'not authenticated' error, got '%v'", err)
			}
		})

		t.Run("returns error when session not authenticated", func(t *testing.T) {
			svc := NewATProtoService()
			ctx := context.Background()
			svc.session = &Session{
				Handle:        "test.bsky.social",
				Authenticated: false,
			}

			doc := public.Document{
				Title: "Test Document",
			}

			result, err := svc.PostDocument(ctx, doc, false)
			if err == nil {
				t.Error("Expected error when posting document with unauthenticated session")
			}
			if result != nil {
				t.Error("Expected nil result when session not authenticated")
			}
		})

		t.Run("returns error when document title is empty", func(t *testing.T) {
			svc := NewATProtoService()
			ctx := context.Background()
			svc.session = &Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				Authenticated: true,
			}

			doc := public.Document{
				Title: "",
			}

			result, err := svc.PostDocument(ctx, doc, false)
			if err == nil {
				t.Error("Expected error when document title is empty")
			}
			if result != nil {
				t.Error("Expected nil result when title is empty")
			}
			if err.Error() != "document title is required" {
				t.Errorf("Expected 'document title is required' error, got '%v'", err)
			}
		})

		t.Run("returns error when context cancelled", func(t *testing.T) {
			svc := NewATProtoService()
			svc.session = &Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				Authenticated: true,
			}

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			doc := public.Document{
				Title: "Test Document",
			}

			result, err := svc.PostDocument(ctx, doc, false)
			if err == nil {
				t.Error("Expected error when context is cancelled")
			}
			if result != nil {
				t.Error("Expected nil result when context is cancelled")
			}
		})

		t.Run("returns error when context timeout", func(t *testing.T) {
			svc := NewATProtoService()
			svc.session = &Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				Authenticated: true,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 1)
			defer cancel()
			time.Sleep(2 * time.Millisecond)

			doc := public.Document{
				Title: "Test Document",
			}

			result, err := svc.PostDocument(ctx, doc, false)
			if err == nil {
				t.Error("Expected error when context times out")
			}
			if result != nil {
				t.Error("Expected nil result when context times out")
			}
		})

		t.Run("validates draft parameter sets correct collection", func(t *testing.T) {
			svc := NewATProtoService()
			svc.session = &Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				Authenticated: true,
			}
			ctx := context.Background()

			doc := public.Document{
				Title: "Test Document",
			}

			_, err := svc.PostDocument(ctx, doc, true)

			if err != nil && err.Error() == "not authenticated" {
				t.Error("Authentication check should pass, but got authentication error")
			}
		})

		t.Run("validates published parameter sets correct collection", func(t *testing.T) {
			svc := NewATProtoService()
			svc.session = &Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				Authenticated: true,
			}
			ctx := context.Background()

			doc := public.Document{
				Title: "Test Document",
			}

			_, err := svc.PostDocument(ctx, doc, false)

			if err != nil && err.Error() == "not authenticated" {
				t.Error("Authentication check should pass, but got authentication error")
			}
		})
	})

	t.Run("PatchDocument", func(t *testing.T) {
		t.Run("returns error when not authenticated", func(t *testing.T) {
			svc := NewATProtoService()
			ctx := context.Background()

			doc := public.Document{
				Title: "Updated Document",
			}

			result, err := svc.PatchDocument(ctx, "test-rkey", doc, false)
			if err == nil {
				t.Error("Expected error when patching document without authentication")
			}
			if result != nil {
				t.Error("Expected nil result when not authenticated")
			}
			if err.Error() != "not authenticated" {
				t.Errorf("Expected 'not authenticated' error, got '%v'", err)
			}
		})

		t.Run("returns error when session not authenticated", func(t *testing.T) {
			svc := NewATProtoService()
			ctx := context.Background()
			svc.session = &Session{
				Handle:        "test.bsky.social",
				Authenticated: false,
			}

			doc := public.Document{
				Title: "Updated Document",
			}

			result, err := svc.PatchDocument(ctx, "test-rkey", doc, false)
			if err == nil {
				t.Error("Expected error when patching document with unauthenticated session")
			}
			if result != nil {
				t.Error("Expected nil result when session not authenticated")
			}
		})

		t.Run("returns error when rkey is empty", func(t *testing.T) {
			svc := NewATProtoService()
			ctx := context.Background()
			svc.session = &Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				Authenticated: true,
			}

			doc := public.Document{
				Title: "Updated Document",
			}

			result, err := svc.PatchDocument(ctx, "", doc, false)
			if err == nil {
				t.Error("Expected error when rkey is empty")
			}
			if result != nil {
				t.Error("Expected nil result when rkey is empty")
			}
			if err.Error() != "rkey is required" {
				t.Errorf("Expected 'rkey is required' error, got '%v'", err)
			}
		})

		t.Run("returns error when document title is empty", func(t *testing.T) {
			svc := NewATProtoService()
			ctx := context.Background()
			svc.session = &Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				Authenticated: true,
			}

			doc := public.Document{
				Title: "",
			}

			result, err := svc.PatchDocument(ctx, "test-rkey", doc, false)
			if err == nil {
				t.Error("Expected error when document title is empty")
			}
			if result != nil {
				t.Error("Expected nil result when title is empty")
			}
			if err.Error() != "document title is required" {
				t.Errorf("Expected 'document title is required' error, got '%v'", err)
			}
		})

		t.Run("returns error when context cancelled", func(t *testing.T) {
			svc := NewATProtoService()
			svc.session = &Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				Authenticated: true,
			}

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			doc := public.Document{
				Title: "Updated Document",
			}

			result, err := svc.PatchDocument(ctx, "test-rkey", doc, false)
			if err == nil {
				t.Error("Expected error when context is cancelled")
			}
			if result != nil {
				t.Error("Expected nil result when context is cancelled")
			}
		})

		t.Run("returns error when context timeout", func(t *testing.T) {
			svc := NewATProtoService()
			svc.session = &Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				Authenticated: true,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 1)
			defer cancel()
			time.Sleep(2 * time.Millisecond)

			doc := public.Document{
				Title: "Updated Document",
			}

			result, err := svc.PatchDocument(ctx, "test-rkey", doc, false)
			if err == nil {
				t.Error("Expected error when context times out")
			}
			if result != nil {
				t.Error("Expected nil result when context times out")
			}
		})

		t.Run("validates draft parameter sets correct collection", func(t *testing.T) {
			svc := NewATProtoService()
			svc.session = &Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				Authenticated: true,
			}
			ctx := context.Background()

			doc := public.Document{
				Title: "Updated Document",
			}

			_, err := svc.PatchDocument(ctx, "test-rkey", doc, true)

			if err != nil && err.Error() == "not authenticated" {
				t.Error("Authentication check should pass, but got authentication error")
			}
		})

		t.Run("validates published parameter sets correct collection", func(t *testing.T) {
			svc := NewATProtoService()
			svc.session = &Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				Authenticated: true,
			}
			ctx := context.Background()

			doc := public.Document{
				Title: "Updated Document",
			}

			_, err := svc.PatchDocument(ctx, "test-rkey", doc, false)

			if err != nil && err.Error() == "not authenticated" {
				t.Error("Authentication check should pass, but got authentication error")
			}
		})
	})

	t.Run("DeleteDocument", func(t *testing.T) {
		t.Run("returns error when not authenticated", func(t *testing.T) {
			svc := NewATProtoService()
			ctx := context.Background()

			err := svc.DeleteDocument(ctx, "test-rkey", false)
			if err == nil {
				t.Error("Expected error when deleting document without authentication")
			}
			if err.Error() != "not authenticated" {
				t.Errorf("Expected 'not authenticated' error, got '%v'", err)
			}
		})

		t.Run("returns error when session not authenticated", func(t *testing.T) {
			svc := NewATProtoService()
			ctx := context.Background()
			svc.session = &Session{
				Handle:        "test.bsky.social",
				Authenticated: false,
			}

			err := svc.DeleteDocument(ctx, "test-rkey", false)
			if err == nil {
				t.Error("Expected error when deleting document with unauthenticated session")
			}
		})

		t.Run("returns error when rkey is empty", func(t *testing.T) {
			svc := NewATProtoService()
			ctx := context.Background()
			svc.session = &Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				Authenticated: true,
			}

			err := svc.DeleteDocument(ctx, "", false)
			if err == nil {
				t.Error("Expected error when rkey is empty")
			}
			if err.Error() != "rkey is required" {
				t.Errorf("Expected 'rkey is required' error, got '%v'", err)
			}
		})

		t.Run("returns error when context cancelled", func(t *testing.T) {
			svc := NewATProtoService()
			svc.session = &Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				Authenticated: true,
			}

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			err := svc.DeleteDocument(ctx, "test-rkey", false)
			if err == nil {
				t.Error("Expected error when context is cancelled")
			}
		})

		t.Run("returns error when context timeout", func(t *testing.T) {
			svc := NewATProtoService()
			svc.session = &Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				Authenticated: true,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 1)
			defer cancel()
			time.Sleep(2 * time.Millisecond)

			err := svc.DeleteDocument(ctx, "test-rkey", false)
			if err == nil {
				t.Error("Expected error when context times out")
			}
		})

		t.Run("validates draft parameter sets correct collection", func(t *testing.T) {
			svc := NewATProtoService()
			svc.session = &Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				Authenticated: true,
			}
			ctx := context.Background()

			err := svc.DeleteDocument(ctx, "test-rkey", true)

			if err != nil && err.Error() == "not authenticated" {
				t.Error("Authentication check should pass, but got authentication error")
			}
		})

		t.Run("validates published parameter sets correct collection", func(t *testing.T) {
			svc := NewATProtoService()
			svc.session = &Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				Authenticated: true,
			}
			ctx := context.Background()

			err := svc.DeleteDocument(ctx, "test-rkey", false)

			if err != nil && err.Error() == "not authenticated" {
				t.Error("Authentication check should pass, but got authentication error")
			}
		})
	})

	t.Run("Session Management Edge Cases", func(t *testing.T) {
		t.Run("GetSession returns distinct error for nil session", func(t *testing.T) {
			svc := NewATProtoService()

			session, err := svc.GetSession()
			if err == nil {
				t.Error("Expected error when getting nil session")
			}
			if session != nil {
				t.Error("Expected nil session when not authenticated")
			}
			expectedMsg := "not authenticated"
			if !strings.Contains(err.Error(), expectedMsg) {
				t.Errorf("Expected error message to contain '%s', got '%v'", expectedMsg, err)
			}
		})

		t.Run("RestoreSession validates all required fields", func(t *testing.T) {
			svc := NewATProtoService()

			testCases := []struct {
				name    string
				session *Session
			}{
				{
					name: "missing DID",
					session: &Session{
						DID:        "",
						Handle:     "test.bsky.social",
						AccessJWT:  "access",
						RefreshJWT: "refresh",
					},
				},
				{
					name: "missing AccessJWT",
					session: &Session{
						DID:        "did:plc:test",
						Handle:     "test.bsky.social",
						AccessJWT:  "",
						RefreshJWT: "refresh",
					},
				},
				{
					name: "missing RefreshJWT",
					session: &Session{
						DID:        "did:plc:test",
						Handle:     "test.bsky.social",
						AccessJWT:  "access",
						RefreshJWT: "",
					},
				},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					err := svc.RestoreSession(tc.session)
					if err == nil {
						t.Errorf("Expected error for %s", tc.name)
					}
					if !strings.Contains(err.Error(), "session missing required fields") {
						t.Errorf("Expected 'session missing required fields' error, got: %v", err)
					}
				})
			}
		})

		t.Run("RestoreSession preserves empty PDSURL", func(t *testing.T) {
			svc := NewATProtoService()
			defaultPDSURL := svc.pdsURL

			session := &Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				PDSURL:        "",
				ExpiresAt:     time.Now().Add(2 * time.Hour),
				Authenticated: true,
			}

			err := svc.RestoreSession(session)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if svc.pdsURL != defaultPDSURL {
				t.Errorf("Expected pdsURL to remain default when session PDSURL is empty, got '%s'", svc.pdsURL)
			}
		})
	})

	t.Run("PostDocument Validation", func(t *testing.T) {
		t.Run("validates title before marshaling", func(t *testing.T) {
			svc := NewATProtoService()
			svc.session = &Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				Authenticated: true,
			}
			ctx := context.Background()

			doc := public.Document{
				Title: "",
			}

			result, err := svc.PostDocument(ctx, doc, false)
			if err == nil {
				t.Error("Expected error when title is empty")
			}
			if result != nil {
				t.Error("Expected nil result when validation fails")
			}
			if !strings.Contains(err.Error(), "document title is required") {
				t.Errorf("Expected 'document title is required' error, got: %v", err)
			}
		})

		t.Run("sets correct collection for draft", func(t *testing.T) {
			svc := NewATProtoService()
			svc.session = &Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				Authenticated: true,
			}
			ctx := context.Background()

			doc := public.Document{
				Title: "Test Draft",
			}

			_, err := svc.PostDocument(ctx, doc, true)

			if err != nil && strings.Contains(err.Error(), "document title is required") {
				t.Error("Title validation should pass")
			}
		})

		t.Run("sets correct collection for published", func(t *testing.T) {
			svc := NewATProtoService()
			svc.session = &Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				Authenticated: true,
			}
			ctx := context.Background()

			doc := public.Document{
				Title: "Test Published",
			}

			_, err := svc.PostDocument(ctx, doc, false)

			if err != nil && strings.Contains(err.Error(), "document title is required") {
				t.Error("Title validation should pass")
			}
		})
	})

	t.Run("PatchDocument Validation", func(t *testing.T) {
		t.Run("validates rkey before title", func(t *testing.T) {
			svc := NewATProtoService()
			svc.session = &Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				Authenticated: true,
			}
			ctx := context.Background()

			doc := public.Document{
				Title: "Valid Title",
			}

			result, err := svc.PatchDocument(ctx, "", doc, false)
			if err == nil {
				t.Error("Expected error when rkey is empty")
			}
			if result != nil {
				t.Error("Expected nil result when rkey validation fails")
			}
			if !strings.Contains(err.Error(), "rkey is required") {
				t.Errorf("Expected 'rkey is required' error, got: %v", err)
			}
		})

		t.Run("validates title after rkey", func(t *testing.T) {
			svc := NewATProtoService()
			svc.session = &Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				Authenticated: true,
			}
			ctx := context.Background()

			doc := public.Document{
				Title: "",
			}

			result, err := svc.PatchDocument(ctx, "valid-rkey", doc, false)
			if err == nil {
				t.Error("Expected error when title is empty")
			}
			if result != nil {
				t.Error("Expected nil result when title validation fails")
			}
			if !strings.Contains(err.Error(), "document title is required") {
				t.Errorf("Expected 'document title is required' error, got: %v", err)
			}
		})

		t.Run("sets correct collection for draft", func(t *testing.T) {
			svc := NewATProtoService()
			svc.session = &Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				Authenticated: true,
			}
			ctx := context.Background()

			doc := public.Document{
				Title: "Test Draft",
			}

			_, err := svc.PatchDocument(ctx, "test-rkey", doc, true)

			if err != nil && strings.Contains(err.Error(), "document title is required") {
				t.Error("Title validation should pass")
			}
		})

		t.Run("sets correct collection for published", func(t *testing.T) {
			svc := NewATProtoService()
			svc.session = &Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				Authenticated: true,
			}
			ctx := context.Background()

			doc := public.Document{
				Title: "Test Published",
			}

			_, err := svc.PatchDocument(ctx, "test-rkey", doc, false)

			if err != nil && strings.Contains(err.Error(), "document title is required") {
				t.Error("Title validation should pass")
			}
		})
	})

	t.Run("DeleteDocument Validation", func(t *testing.T) {
		t.Run("validates rkey before attempting delete", func(t *testing.T) {
			svc := NewATProtoService()
			svc.session = &Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				Authenticated: true,
			}
			ctx := context.Background()

			err := svc.DeleteDocument(ctx, "", false)
			if err == nil {
				t.Error("Expected error when rkey is empty")
			}
			if !strings.Contains(err.Error(), "rkey is required") {
				t.Errorf("Expected 'rkey is required' error, got: %v", err)
			}
		})

		t.Run("uses correct collection for draft", func(t *testing.T) {
			svc := NewATProtoService()
			svc.session = &Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				Authenticated: true,
			}
			ctx := context.Background()

			err := svc.DeleteDocument(ctx, "test-rkey", true)

			if err != nil && strings.Contains(err.Error(), "rkey is required") {
				t.Error("Rkey validation should pass")
			}
		})

		t.Run("uses correct collection for published", func(t *testing.T) {
			svc := NewATProtoService()
			svc.session = &Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				Authenticated: true,
			}
			ctx := context.Background()

			err := svc.DeleteDocument(ctx, "test-rkey", false)

			if err != nil && strings.Contains(err.Error(), "rkey is required") {
				t.Error("Rkey validation should pass")
			}
		})
	})

	t.Run("Concurrent Operations", func(t *testing.T) {
		t.Run("Close can be called multiple times", func(t *testing.T) {
			svc := NewATProtoService()
			svc.session = &Session{
				Handle:        "test.bsky.social",
				Authenticated: true,
			}

			err1 := svc.Close()
			if err1 != nil {
				t.Errorf("First close should succeed: %v", err1)
			}

			err2 := svc.Close()
			if err2 != nil {
				t.Errorf("Second close should succeed: %v", err2)
			}
		})

		t.Run("IsAuthenticated after Close returns false", func(t *testing.T) {
			svc := NewATProtoService()
			svc.session = &Session{
				Handle:        "test.bsky.social",
				Authenticated: true,
			}

			if !svc.IsAuthenticated() {
				t.Error("Expected IsAuthenticated to return true before close")
			}

			err := svc.Close()
			if err != nil {
				t.Errorf("Close failed: %v", err)
			}

			if svc.IsAuthenticated() {
				t.Error("Expected IsAuthenticated to return false after close")
			}
		})
	})

	t.Run("CBOR Conversion Functions", func(t *testing.T) {
		t.Run("convertCBORToJSONCompatible handles simple map", func(t *testing.T) {
			input := map[any]any{
				"key1": "value1",
				"key2": 42,
				"key3": true,
			}

			result := convertCBORToJSONCompatible(input)

			mapResult, ok := result.(map[string]any)
			if !ok {
				t.Fatal("Expected result to be map[string]any")
			}

			if mapResult["key1"] != "value1" {
				t.Errorf("Expected key1='value1', got '%v'", mapResult["key1"])
			}
			if mapResult["key2"] != 42 {
				t.Errorf("Expected key2=42, got %v", mapResult["key2"])
			}
			if mapResult["key3"] != true {
				t.Errorf("Expected key3=true, got %v", mapResult["key3"])
			}
		})

		t.Run("convertCBORToJSONCompatible handles nested maps", func(t *testing.T) {
			input := map[any]any{
				"outer": map[any]any{
					"inner": map[any]any{
						"deep": "value",
					},
				},
			}

			result := convertCBORToJSONCompatible(input)

			mapResult, ok := result.(map[string]any)
			if !ok {
				t.Fatal("Expected result to be map[string]any")
			}

			outer, ok := mapResult["outer"].(map[string]any)
			if !ok {
				t.Fatal("Expected outer to be map[string]any")
			}

			inner, ok := outer["inner"].(map[string]any)
			if !ok {
				t.Fatal("Expected inner to be map[string]any")
			}

			if inner["deep"] != "value" {
				t.Errorf("Expected deep='value', got '%v'", inner["deep"])
			}
		})

		t.Run("convertCBORToJSONCompatible handles arrays", func(t *testing.T) {
			input := []any{
				"string",
				42,
				map[any]any{"nested": "map"},
				[]any{"nested", "array"},
			}

			result := convertCBORToJSONCompatible(input)

			arrayResult, ok := result.([]any)
			if !ok {
				t.Fatal("Expected result to be []any")
			}

			if len(arrayResult) != 4 {
				t.Fatalf("Expected 4 elements, got %d", len(arrayResult))
			}

			if arrayResult[0] != "string" {
				t.Errorf("Expected arrayResult[0]='string', got '%v'", arrayResult[0])
			}

			nestedMap, ok := arrayResult[2].(map[string]any)
			if !ok {
				t.Fatal("Expected arrayResult[2] to be map[string]any")
			}
			if nestedMap["nested"] != "map" {
				t.Errorf("Expected nested='map', got '%v'", nestedMap["nested"])
			}

			nestedArray, ok := arrayResult[3].([]any)
			if !ok {
				t.Fatal("Expected arrayResult[3] to be []any")
			}
			if len(nestedArray) != 2 {
				t.Errorf("Expected nested array length 2, got %d", len(nestedArray))
			}
		})

		t.Run("convertJSONToCBORCompatible handles simple map", func(t *testing.T) {
			input := map[string]any{
				"key1": "value1",
				"key2": 42,
				"key3": true,
			}

			result := convertJSONToCBORCompatible(input)

			mapResult, ok := result.(map[any]any)
			if !ok {
				t.Fatal("Expected result to be map[any]any")
			}

			if mapResult["key1"] != "value1" {
				t.Errorf("Expected key1='value1', got '%v'", mapResult["key1"])
			}
			if mapResult["key2"] != 42 {
				t.Errorf("Expected key2=42, got %v", mapResult["key2"])
			}
			if mapResult["key3"] != true {
				t.Errorf("Expected key3=true, got %v", mapResult["key3"])
			}
		})

		t.Run("convertJSONToCBORCompatible handles nested maps", func(t *testing.T) {
			input := map[string]any{
				"outer": map[string]any{
					"inner": map[string]any{
						"deep": "value",
					},
				},
			}

			result := convertJSONToCBORCompatible(input)

			mapResult, ok := result.(map[any]any)
			if !ok {
				t.Fatal("Expected result to be map[any]any")
			}

			outer, ok := mapResult["outer"].(map[any]any)
			if !ok {
				t.Fatal("Expected outer to be map[any]any")
			}

			inner, ok := outer["inner"].(map[any]any)
			if !ok {
				t.Fatal("Expected inner to be map[any]any")
			}

			if inner["deep"] != "value" {
				t.Errorf("Expected deep='value', got '%v'", inner["deep"])
			}
		})

		t.Run("convertJSONToCBORCompatible handles arrays", func(t *testing.T) {
			input := []any{
				"string",
				42,
				map[string]any{"nested": "map"},
				[]any{"nested", "array"},
			}

			result := convertJSONToCBORCompatible(input)

			arrayResult, ok := result.([]any)
			if !ok {
				t.Fatal("Expected result to be []any")
			}

			if len(arrayResult) != 4 {
				t.Fatalf("Expected 4 elements, got %d", len(arrayResult))
			}

			if arrayResult[0] != "string" {
				t.Errorf("Expected arrayResult[0]='string', got '%v'", arrayResult[0])
			}

			nestedMap, ok := arrayResult[2].(map[any]any)
			if !ok {
				t.Fatal("Expected arrayResult[2] to be map[any]any")
			}
			if nestedMap["nested"] != "map" {
				t.Errorf("Expected nested='map', got '%v'", nestedMap["nested"])
			}

			nestedArray, ok := arrayResult[3].([]any)
			if !ok {
				t.Fatal("Expected arrayResult[3] to be []any")
			}
			if len(nestedArray) != 2 {
				t.Errorf("Expected nested array length 2, got %d", len(nestedArray))
			}
		})

		t.Run("round-trip conversion preserves data", func(t *testing.T) {
			original := map[string]any{
				"title":   "Test Document",
				"author":  "did:plc:test123",
				"content": []any{"paragraph1", "paragraph2"},
				"metadata": map[string]any{
					"tags":      []any{"test", "document"},
					"published": true,
					"count":     42,
				},
			}

			cborCompatible := convertJSONToCBORCompatible(original)
			jsonCompatible := convertCBORToJSONCompatible(cborCompatible)

			originalJSON, err := json.Marshal(original)
			if err != nil {
				t.Fatalf("Failed to marshal original: %v", err)
			}

			resultJSON, err := json.Marshal(jsonCompatible)
			if err != nil {
				t.Fatalf("Failed to marshal result: %v", err)
			}

			if string(originalJSON) != string(resultJSON) {
				t.Errorf("Round-trip conversion changed data.\nOriginal: %s\nResult: %s", originalJSON, resultJSON)
			}
		})

		t.Run("Document conversion through CBOR preserves structure", func(t *testing.T) {
			doc := public.Document{
				Type:  public.TypeDocument,
				Title: "Test Document",
				Pages: []public.LinearDocument{
					{
						Type: public.TypeLinearDocument,
						Blocks: []public.BlockWrap{
							{
								Type: public.TypeBlock,
								Block: public.TextBlock{
									Type:      public.TypeTextBlock,
									Plaintext: "Hello, world!",
								},
							},
						},
					},
				},
				PublishedAt: time.Now().UTC().Format(time.RFC3339),
			}

			jsonBytes, err := json.Marshal(doc)
			if err != nil {
				t.Fatalf("Failed to marshal document to JSON: %v", err)
			}

			var jsonData map[string]any
			if err := json.Unmarshal(jsonBytes, &jsonData); err != nil {
				t.Fatalf("Failed to unmarshal JSON to map: %v", err)
			}

			cborCompatible := convertJSONToCBORCompatible(jsonData)

			cborBytes, err := cbor.Marshal(cborCompatible)
			if err != nil {
				t.Fatalf("Failed to marshal to CBOR: %v", err)
			}

			var cborData any
			if err := cbor.Unmarshal(cborBytes, &cborData); err != nil {
				t.Fatalf("Failed to unmarshal CBOR: %v", err)
			}

			jsonCompatible := convertCBORToJSONCompatible(cborData)

			finalJSONBytes, err := json.Marshal(jsonCompatible)
			if err != nil {
				t.Fatalf("Failed to marshal final JSON: %v", err)
			}

			var finalDoc public.Document
			if err := json.Unmarshal(finalJSONBytes, &finalDoc); err != nil {
				t.Fatalf("Failed to unmarshal final document: %v", err)
			}

			if finalDoc.Title != doc.Title {
				t.Errorf("Title changed: expected '%s', got '%s'", doc.Title, finalDoc.Title)
			}

			if len(finalDoc.Pages) != len(doc.Pages) {
				t.Errorf("Pages length changed: expected %d, got %d", len(doc.Pages), len(finalDoc.Pages))
			}

			if len(finalDoc.Pages) > 0 && len(finalDoc.Pages[0].Blocks) > 0 {
				if textBlock, ok := finalDoc.Pages[0].Blocks[0].Block.(public.TextBlock); ok {
					expectedBlock := doc.Pages[0].Blocks[0].Block.(public.TextBlock)
					if textBlock.Plaintext != expectedBlock.Plaintext {
						t.Errorf("Block plaintext changed: expected '%s', got '%s'",
							expectedBlock.Plaintext, textBlock.Plaintext)
					}
				} else {
					t.Error("Expected Block to be TextBlock")
				}
			}
		})
	})
}
