package services

import (
	"context"
	"testing"
	"time"
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
	})
}
