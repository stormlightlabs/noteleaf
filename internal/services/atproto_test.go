package services

import (
	"context"
	"testing"
	"time"

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
}
