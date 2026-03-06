package tests

import (
	"errors"
	"testing"
	"time"

	apperrors "github.com/ulbwa/telegram-oidc-provider/internal/errors"
	"github.com/ulbwa/telegram-oidc-provider/internal/model"
)

func TestNewBotSuccess(t *testing.T) {
	t.Parallel()

	bot, err := model.NewBot(123456, "OIDC Login Bot", "oidc_login_bot", "123:ABC")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if bot.ID != 123456 {
		t.Fatalf("expected id 123456, got %d", bot.ID)
	}

	if bot.Name != "OIDC Login Bot" {
		t.Fatalf("expected name %q, got %q", "OIDC Login Bot", bot.Name)
	}

	if bot.Username != "oidc_login_bot" {
		t.Fatalf("expected username %q, got %v", "oidc_login_bot", bot.Username)
	}

	if bot.Token != "123:ABC" {
		t.Fatalf("expected token %q, got %q", "123:ABC", bot.Token)
	}

	if bot.OAuthClientID != nil {
		t.Fatalf("expected nil oauth client id for a new bot")
	}

	if bot.CreatedAt.IsZero() {
		t.Fatalf("expected non-zero createdAt")
	}

	if bot.UpdatedAt != nil {
		t.Fatalf("expected nil updatedAt for a new bot")
	}
}

func TestNewBotValidation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		id       int64
		botName  string
		username string
		token    string
		expected error
	}{
		{
			name:     "invalid id",
			id:       0,
			botName:  "OIDC Bot",
			username: "oidc_login_bot",
			token:    "123:ABC",
			expected: apperrors.ErrBotInvalidID,
		},
		{
			name:     "empty name",
			id:       1,
			botName:  "   ",
			username: "oidc_login_bot",
			token:    "123:ABC",
			expected: apperrors.ErrBotNameInvalid,
		},
		{
			name:     "name with trailing space",
			id:       1,
			botName:  "OIDC Bot ",
			username: "oidc_login_bot",
			token:    "123:ABC",
			expected: apperrors.ErrBotNameInvalid,
		},
		{
			name:     "invalid username format",
			id:       1,
			botName:  "OIDC Bot",
			username: "bad-name",
			token:    "123:ABC",
			expected: apperrors.ErrBotUsernameInvalid,
		},
		{
			name:     "username too short",
			id:       1,
			botName:  "OIDC Bot",
			username: "ab",
			token:    "123:ABC",
			expected: apperrors.ErrBotUsernameInvalid,
		},
		{
			name:     "username without bot suffix is allowed",
			id:       1,
			botName:  "OIDC Bot",
			username: "oidc_login_user",
			token:    "123:ABC",
			expected: nil,
		},
		{
			name:     "empty username",
			id:       1,
			botName:  "OIDC Bot",
			username: "",
			token:    "123:ABC",
			expected: apperrors.ErrBotUsernameInvalid,
		},
		{
			name:     "username with trailing space",
			id:       1,
			botName:  "OIDC Bot",
			username: "oidc_login_bot ",
			token:    "123:ABC",
			expected: apperrors.ErrBotUsernameInvalid,
		},
		{
			name:     "empty token",
			id:       1,
			botName:  "OIDC Bot",
			username: "oidc_login_bot",
			token:    "",
			expected: apperrors.ErrBotTokenInvalid,
		},
		{
			name:     "token with trailing space",
			id:       1,
			botName:  "OIDC Bot",
			username: "oidc_login_bot",
			token:    "123:ABC ",
			expected: apperrors.ErrBotTokenInvalid,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			_, err := model.NewBot(testCase.id, testCase.botName, testCase.username, testCase.token)
			if testCase.expected == nil {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}

				return
			}

			if !errors.Is(err, testCase.expected) {
				t.Fatalf("expected error %v, got %v", testCase.expected, err)
			}
		})
	}
}

func TestBotUpdateProfileAndSetToken(t *testing.T) {
	t.Parallel()

	bot, err := model.NewBot(777, "Old Bot", "old_login_bot", "111:OLD")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = bot.UpdateProfile("New Bot", "new_login_bot")
	if err != nil {
		t.Fatalf("expected no error while updating profile, got %v", err)
	}

	if bot.Name != "New Bot" {
		t.Fatalf("expected name %q, got %q", "New Bot", bot.Name)
	}

	if bot.Username != "new_login_bot" {
		t.Fatalf("expected username %q, got %v", "new_login_bot", bot.Username)
	}

	if bot.UpdatedAt == nil {
		t.Fatalf("expected non-nil updatedAt after profile update")
	}

	profileUpdatedAt := *bot.UpdatedAt

	time.Sleep(2 * time.Millisecond)
	err = bot.SetToken("222:NEW")
	if err != nil {
		t.Fatalf("expected no error while setting token, got %v", err)
	}

	if bot.Token != "222:NEW" {
		t.Fatalf("expected token %q, got %q", "222:NEW", bot.Token)
	}

	if bot.UpdatedAt == nil {
		t.Fatalf("expected non-nil updatedAt after token rotation")
	}

	if !bot.UpdatedAt.After(profileUpdatedAt) {
		t.Fatalf("expected updatedAt %v to be after %v", *bot.UpdatedAt, profileUpdatedAt)
	}
}

func TestBotLastModifiedAt(t *testing.T) {
	t.Parallel()

	bot, err := model.NewBot(777, "OIDC Bot", "oidc_login_user", "111:OLD")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !bot.LastModifiedAt().Equal(bot.CreatedAt) {
		t.Fatalf("expected last modified to equal createdAt")
	}

	time.Sleep(2 * time.Millisecond)
	err = bot.SetToken("222:NEW")
	if err != nil {
		t.Fatalf("expected no error while setting token, got %v", err)
	}

	if bot.UpdatedAt == nil {
		t.Fatalf("expected non-nil updatedAt after token change")
	}

	if !bot.LastModifiedAt().Equal(*bot.UpdatedAt) {
		t.Fatalf("expected last modified to equal updatedAt")
	}
}

func TestBotMaskedToken(t *testing.T) {
	t.Parallel()

	bot, err := model.NewBot(777, "OIDC Bot", "oidc_login_user", "123456:ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if bot.MaskedToken() != "123456:************WXYZ" {
		t.Fatalf("unexpected masked token: %q", bot.MaskedToken())
	}
}

func TestBotMaskedTokenShortSecret(t *testing.T) {
	t.Parallel()

	bot, err := model.NewBot(777, "OIDC Bot", "oidc_login_user", "123456:ABC")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if bot.MaskedToken() != "123456:************" {
		t.Fatalf("unexpected masked token: %q", bot.MaskedToken())
	}
}

func TestBotMaskedTokenTooShortToRevealTail(t *testing.T) {
	t.Parallel()

	bot, err := model.NewBot(777, "OIDC Bot", "oidc_login_user", "123456:ABCDE")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if bot.MaskedToken() != "123456:************" {
		t.Fatalf("unexpected masked token: %q", bot.MaskedToken())
	}
}

func TestBotMaskedTokenBorderlineLengthDoesNotRevealTail(t *testing.T) {
	t.Parallel()

	bot, err := model.NewBot(777, "OIDC Bot", "oidc_login_user", "123456:ABCDEFGHIJKL")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if bot.MaskedToken() != "123456:************" {
		t.Fatalf("unexpected masked token: %q", bot.MaskedToken())
	}
}

func TestBotMaskedTokenInvalidTokenBotID(t *testing.T) {
	t.Parallel()

	bot, err := model.NewBot(777, "OIDC Bot", "oidc_login_user", "111:OLD")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = bot.SetToken("abc:ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	if err != nil {
		t.Fatalf("expected no error while setting token, got %v", err)
	}

	if bot.MaskedToken() != "************WXYZ" {
		t.Fatalf("unexpected masked token: %q", bot.MaskedToken())
	}
}

func TestBotUpdateProfileNoChangesReturnsNilAndDoesNotTouch(t *testing.T) {
	t.Parallel()

	bot, err := model.NewBot(777, "OIDC Bot", "oidc_login_bot", "111:OLD")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = bot.UpdateProfile("OIDC Bot", "oidc_login_bot")
	if err != nil {
		t.Fatalf("expected no error on no-op profile update, got %v", err)
	}

	if bot.UpdatedAt != nil {
		t.Fatalf("expected updatedAt to stay nil on no-op update")
	}
}

func TestRestoreBotWithNullableUpdatedAt(t *testing.T) {
	t.Parallel()

	createdAt := time.Now().UTC().Add(-time.Minute)

	bot, err := model.RestoreBot(1, "OIDC Bot", "oidc_login_bot", "123:ABC", nil, createdAt, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if bot.UpdatedAt != nil {
		t.Fatalf("expected nil updatedAt for restored bot")
	}
}

func TestRestoreBotUpdatedAtBeforeCreatedAt(t *testing.T) {
	t.Parallel()

	createdAt := time.Now().UTC()
	updatedAt := createdAt.Add(-time.Second)

	_, err := model.RestoreBot(1, "OIDC Bot", "oidc_login_bot", "123:ABC", nil, createdAt, &updatedAt)
	if !errors.Is(err, apperrors.ErrBotUpdatedBeforeCreate) {
		t.Fatalf("expected error %v, got %v", apperrors.ErrBotUpdatedBeforeCreate, err)
	}
}

func TestRestoreBotClonesUpdatedAtPointer(t *testing.T) {
	t.Parallel()

	createdAt := time.Now().UTC().Add(-time.Minute)
	originalUpdatedAt := createdAt.Add(10 * time.Second)

	bot, err := model.RestoreBot(1, "OIDC Bot", "oidc_login_bot", "123:ABC", nil, createdAt, &originalUpdatedAt)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if bot.UpdatedAt == nil {
		t.Fatalf("expected non-nil updatedAt")
	}

	if !bot.UpdatedAt.Equal(originalUpdatedAt) {
		t.Fatalf("expected updatedAt %v, got %v", originalUpdatedAt, *bot.UpdatedAt)
	}

	mutated := originalUpdatedAt.Add(2 * time.Minute)
	originalUpdatedAt = mutated

	if bot.UpdatedAt.Equal(originalUpdatedAt) {
		t.Fatalf("expected bot updatedAt to be isolated from external pointer mutation")
	}
}

func TestBotSetUsernameEmpty(t *testing.T) {
	t.Parallel()

	bot, err := model.NewBot(777, "OIDC Bot", "oidc_login_bot", "111:OLD")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = bot.SetUsername("")
	if !errors.Is(err, apperrors.ErrBotUsernameInvalid) {
		t.Fatalf("expected error %v, got %v", apperrors.ErrBotUsernameInvalid, err)
	}
}

func TestBotBindAndUnbindOAuthClient(t *testing.T) {
	t.Parallel()

	bot, err := model.NewBot(777, "OIDC Bot", "oidc_login_bot", "111:OLD")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = bot.BindOAuthClient("hydra-client-1")
	if err != nil {
		t.Fatalf("expected no error while binding oauth client, got %v", err)
	}

	if bot.OAuthClientID == nil || *bot.OAuthClientID != "hydra-client-1" {
		t.Fatalf("expected oauth client id hydra-client-1, got %v", bot.OAuthClientID)
	}

	if bot.UpdatedAt == nil {
		t.Fatalf("expected non-nil updatedAt after binding oauth client")
	}

	updatedAtAfterBind := *bot.UpdatedAt
	time.Sleep(2 * time.Millisecond)

	err = bot.UnbindOAuthClient()
	if err != nil {
		t.Fatalf("expected no error while unbinding oauth client, got %v", err)
	}

	if bot.OAuthClientID != nil {
		t.Fatalf("expected nil oauth client id after unbind")
	}

	if bot.UpdatedAt == nil || !bot.UpdatedAt.After(updatedAtAfterBind) {
		t.Fatalf("expected updatedAt to be updated after unbind")
	}
}

func TestBotBindOAuthClientValidation(t *testing.T) {
	t.Parallel()

	bot, err := model.NewBot(777, "OIDC Bot", "oidc_login_bot", "111:OLD")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = bot.BindOAuthClient("   ")
	if !errors.Is(err, apperrors.ErrBotOAuthClientIDInvalid) {
		t.Fatalf("expected error %v, got %v", apperrors.ErrBotOAuthClientIDInvalid, err)
	}

	err = bot.BindOAuthClient(" hydra-client ")
	if !errors.Is(err, apperrors.ErrBotOAuthClientIDInvalid) {
		t.Fatalf("expected error %v, got %v", apperrors.ErrBotOAuthClientIDInvalid, err)
	}
}

func TestRestoreBotWithOAuthClientID(t *testing.T) {
	t.Parallel()

	createdAt := time.Now().UTC().Add(-time.Minute)
	oauthClientID := "hydra-client-1"

	bot, err := model.RestoreBot(
		1,
		"OIDC Bot",
		"oidc_login_bot",
		"123:ABC",
		&oauthClientID,
		createdAt,
		nil,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if bot.OAuthClientID == nil || *bot.OAuthClientID != "hydra-client-1" {
		t.Fatalf("expected oauth client id hydra-client-1, got %v", bot.OAuthClientID)
	}
}
