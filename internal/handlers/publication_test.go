package handlers

import (
	"context"
	"testing"
	"time"

	"github.com/stormlightlabs/noteleaf/internal/services"
	"github.com/stormlightlabs/noteleaf/internal/store"
)

func TestPublicationHandler(t *testing.T) {
	t.Run("sessionFromConfig", func(t *testing.T) {
		t.Run("returns error when DID is missing", func(t *testing.T) {
			config := &store.Config{
				ATProtoDID:        "",
				ATProtoHandle:     "test.bsky.social",
				ATProtoAccessJWT:  "access_token",
				ATProtoRefreshJWT: "refresh_token",
			}

			_, err := sessionFromConfig(config)
			if err == nil {
				t.Error("Expected error when DID is missing")
			}
		})

		t.Run("returns error when AccessJWT is missing", func(t *testing.T) {
			config := &store.Config{
				ATProtoDID:        "did:plc:test123",
				ATProtoHandle:     "test.bsky.social",
				ATProtoAccessJWT:  "",
				ATProtoRefreshJWT: "refresh_token",
			}

			_, err := sessionFromConfig(config)
			if err == nil {
				t.Error("Expected error when AccessJWT is missing")
			}
		})

		t.Run("returns error when RefreshJWT is missing", func(t *testing.T) {
			config := &store.Config{
				ATProtoDID:        "did:plc:test123",
				ATProtoHandle:     "test.bsky.social",
				ATProtoAccessJWT:  "access_token",
				ATProtoRefreshJWT: "",
			}

			_, err := sessionFromConfig(config)
			if err == nil {
				t.Error("Expected error when RefreshJWT is missing")
			}
		})

		t.Run("successfully creates session from complete config", func(t *testing.T) {
			expiresAt := time.Now().Add(2 * time.Hour)
			config := &store.Config{
				ATProtoDID:        "did:plc:test123",
				ATProtoHandle:     "test.bsky.social",
				ATProtoAccessJWT:  "access_token",
				ATProtoRefreshJWT: "refresh_token",
				ATProtoPDSURL:     "https://bsky.social",
				ATProtoExpiresAt:  expiresAt.Format("2006-01-02T15:04:05Z07:00"),
			}

			session, err := sessionFromConfig(config)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if session.DID != config.ATProtoDID {
				t.Errorf("Expected DID '%s', got '%s'", config.ATProtoDID, session.DID)
			}
			if session.Handle != config.ATProtoHandle {
				t.Errorf("Expected Handle '%s', got '%s'", config.ATProtoHandle, session.Handle)
			}
			if session.AccessJWT != config.ATProtoAccessJWT {
				t.Errorf("Expected AccessJWT '%s', got '%s'", config.ATProtoAccessJWT, session.AccessJWT)
			}
			if session.RefreshJWT != config.ATProtoRefreshJWT {
				t.Errorf("Expected RefreshJWT '%s', got '%s'", config.ATProtoRefreshJWT, session.RefreshJWT)
			}
			if session.PDSURL != config.ATProtoPDSURL {
				t.Errorf("Expected PDSURL '%s', got '%s'", config.ATProtoPDSURL, session.PDSURL)
			}
			if !session.Authenticated {
				t.Error("Expected session to be authenticated")
			}
		})

		t.Run("handles missing expiry time gracefully", func(t *testing.T) {
			config := &store.Config{
				ATProtoDID:        "did:plc:test123",
				ATProtoHandle:     "test.bsky.social",
				ATProtoAccessJWT:  "access_token",
				ATProtoRefreshJWT: "refresh_token",
				ATProtoExpiresAt:  "",
			}

			session, err := sessionFromConfig(config)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if session.ExpiresAt.After(time.Now()) {
				t.Error("Expected ExpiresAt to be in the past when not provided")
			}
		})

		t.Run("handles invalid expiry time format gracefully", func(t *testing.T) {
			config := &store.Config{
				ATProtoDID:        "did:plc:test123",
				ATProtoHandle:     "test.bsky.social",
				ATProtoAccessJWT:  "access_token",
				ATProtoRefreshJWT: "refresh_token",
				ATProtoExpiresAt:  "invalid-timestamp",
			}

			session, err := sessionFromConfig(config)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if session.ExpiresAt.After(time.Now()) {
				t.Error("Expected ExpiresAt to be in the past when parse fails")
			}
		})
	})

	t.Run("Auth", func(t *testing.T) {
		t.Run("validates required parameters", func(t *testing.T) {
			_ = NewHandlerTestSuite(t)
			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			err := handler.Auth(ctx, "", "password")
			if err == nil {
				t.Error("Expected error when handle is empty")
			}

			err = handler.Auth(ctx, "handle", "")
			if err == nil {
				t.Error("Expected error when password is empty")
			}
		})

	})

	t.Run("GetAuthStatus", func(t *testing.T) {
		t.Run("returns not authenticated when no session", func(t *testing.T) {
			_ = NewHandlerTestSuite(t)
			handler := CreateHandler(t, NewPublicationHandler)

			status := handler.GetAuthStatus()
			if status != "Not authenticated" {
				t.Errorf("Expected 'Not authenticated', got '%s'", status)
			}
		})

		t.Run("returns authenticated status with session", func(t *testing.T) {
			_ = NewHandlerTestSuite(t)
			handler := CreateHandler(t, NewPublicationHandler)

			session := &services.Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				PDSURL:        "https://bsky.social",
				ExpiresAt:     time.Now().Add(2 * time.Hour),
				Authenticated: true,
			}

			err := handler.atproto.RestoreSession(session)
			if err != nil {
				t.Fatalf("Failed to restore session: %v", err)
			}

			status := handler.GetAuthStatus()
			expectedStatus := "Authenticated as test.bsky.social"
			if status != expectedStatus {
				t.Errorf("Expected '%s', got '%s'", expectedStatus, status)
			}
		})
	})

	t.Run("NewPublicationHandler", func(t *testing.T) {
		t.Run("creates handler successfully", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler, err := NewPublicationHandler()
			if err != nil {
				t.Fatalf("Expected no error creating handler, got %v", err)
			}
			defer handler.Close()

			if handler.db == nil {
				t.Error("Expected database to be initialized")
			}
			if handler.config == nil {
				t.Error("Expected config to be initialized")
			}
			if handler.repos == nil {
				t.Error("Expected repos to be initialized")
			}
			if handler.atproto == nil {
				t.Error("Expected atproto service to be initialized")
			}
		})

		t.Run("restores session from config when available", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			config, err := store.LoadConfig()
			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}

			config.ATProtoDID = "did:plc:test123"
			config.ATProtoHandle = "test.bsky.social"
			config.ATProtoAccessJWT = "access_token"
			config.ATProtoRefreshJWT = "refresh_token"
			config.ATProtoPDSURL = "https://bsky.social"
			config.ATProtoExpiresAt = time.Now().Add(2 * time.Hour).Format("2006-01-02T15:04:05Z07:00")

			err = store.SaveConfig(config)
			if err != nil {
				t.Fatalf("Failed to save config: %v", err)
			}

			handler, err := NewPublicationHandler()
			if err != nil {
				t.Fatalf("Expected no error creating handler, got %v", err)
			}
			defer handler.Close()

			if !handler.atproto.IsAuthenticated() {
				t.Error("Expected handler to be authenticated after restoring from config")
			}

			session, err := handler.atproto.GetSession()
			if err != nil {
				t.Errorf("Expected to get session, got error: %v", err)
			}
			if session.DID != config.ATProtoDID {
				t.Errorf("Expected DID '%s', got '%s'", config.ATProtoDID, session.DID)
			}
		})

		t.Run("handles empty config gracefully", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler, err := NewPublicationHandler()
			if err != nil {
				t.Fatalf("Expected no error creating handler, got %v", err)
			}
			defer handler.Close()

			if handler.atproto.IsAuthenticated() {
				t.Error("Expected handler to not be authenticated with empty config")
			}
		})
	})

	t.Run("Close", func(t *testing.T) {
		t.Run("cleans up resources properly", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler, err := NewPublicationHandler()
			if err != nil {
				t.Fatalf("Expected no error creating handler, got %v", err)
			}

			err = handler.Close()
			if err != nil {
				t.Errorf("Expected no error on close, got %v", err)
			}
		})
	})
}
