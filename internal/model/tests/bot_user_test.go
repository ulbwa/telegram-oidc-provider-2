package tests

import (
	"errors"
	"testing"
	"time"

	"github.com/ulbwa/telegram-oidc-provider/internal/model"
	"github.com/ulbwa/telegram-oidc-provider/pkg/utils"
)

func TestNewUserInfoSuccess(t *testing.T) {
	t.Parallel()

	info, err := model.NewUserInfo(
		"Ivan",
		utils.Ptr("Petrov"),
		utils.Ptr("ivan_petrov"),
		utils.Ptr("https://cdn.example.com/photo.jpg"),
		utils.Ptr(true),
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if info.FirstName != "Ivan" {
		t.Fatalf("expected first name Ivan, got %q", info.FirstName)
	}

	if info.Username == nil || *info.Username != "ivan_petrov" {
		t.Fatalf("expected username ivan_petrov, got %v", info.Username)
	}
}

func TestNewUserInfoValidation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		firstName string
		lastName  *string
		username  *string
		photoURL  *string
		expected  error
	}{
		{
			name:      "empty first name",
			firstName: "",
			expected:  model.ErrUserInvalidFirstName,
		},
		{
			name:      "username invalid",
			firstName: "Ivan",
			username:  utils.Ptr("!bad"),
			expected:  model.ErrUserInvalidUsername,
		},
		{
			name:      "username only spaces is invalid",
			firstName: "Ivan",
			username:  utils.Ptr("   "),
			expected:  model.ErrUserInvalidUsername,
		},
		{
			name:      "photo url invalid",
			firstName: "Ivan",
			photoURL:  utils.Ptr("ftp://example.com/photo.jpg"),
			expected:  model.ErrUserInvalidPhotoURL,
		},
		{
			name:      "last name has spaces",
			firstName: "Ivan",
			lastName:  utils.Ptr(" Petrov "),
			expected:  model.ErrUserInvalidLastName,
		},
		{
			name:      "last name only spaces is invalid",
			firstName: "Ivan",
			lastName:  utils.Ptr("   "),
			expected:  model.ErrUserInvalidLastName,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			_, err := model.NewUserInfo(testCase.firstName, testCase.lastName, testCase.username, testCase.photoURL, nil)
			if !errors.Is(err, testCase.expected) {
				t.Fatalf("expected error %v, got %v", testCase.expected, err)
			}
		})
	}
}

func TestNewBotUserSuccess(t *testing.T) {
	t.Parallel()

	info, err := model.NewUserInfo("Ivan", nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	botUser, err := model.NewBotUser(100, 200, info, "127.0.0.1", utils.Ptr("Mozilla/5.0"), utils.Ptr("ru-RU"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if botUser.BotID != 100 || botUser.UserID != 200 {
		t.Fatalf("unexpected ids: bot=%d user=%d", botUser.BotID, botUser.UserID)
	}

	if botUser.LastLoginIP != "127.0.0.1" {
		t.Fatalf("expected login ip 127.0.0.1, got %q", botUser.LastLoginIP)
	}

	if botUser.UpdatedAt != nil {
		t.Fatalf("expected nil updatedAt for new entity")
	}

	if botUser.LastLoginAt.IsZero() {
		t.Fatalf("expected non-zero last login at")
	}
}

func TestBotUserUpdateInfoNoChanges(t *testing.T) {
	t.Parallel()

	info, err := model.NewUserInfo("Ivan", nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	botUser, err := model.NewBotUser(100, 200, info, "127.0.0.1", nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = botUser.UpdateInfo(info)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if botUser.UpdatedAt != nil {
		t.Fatalf("expected updatedAt to remain nil on no-op")
	}
}

func TestBotUserUpdateInfoChanged(t *testing.T) {
	t.Parallel()

	info, err := model.NewUserInfo("Ivan", nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	botUser, err := model.NewBotUser(100, 200, info, "127.0.0.1", nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	newInfo, err := model.NewUserInfo("Ivan", utils.Ptr("Petrov"), utils.Ptr("ivan_petrov"), nil, utils.Ptr(true))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = botUser.UpdateInfo(newInfo)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if botUser.UpdatedAt == nil {
		t.Fatalf("expected non-nil updatedAt after update")
	}

	if botUser.Info.Username == nil || *botUser.Info.Username != "ivan_petrov" {
		t.Fatalf("expected username ivan_petrov, got %v", botUser.Info.Username)
	}
}

func TestBotUserRecordLogin(t *testing.T) {
	t.Parallel()

	info, err := model.NewUserInfo("Ivan", nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	botUser, err := model.NewBotUser(100, 200, info, "127.0.0.1", nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	initialLastLogin := botUser.LastLoginAt
	time.Sleep(2 * time.Millisecond)

	err = botUser.RecordLogin("10.0.0.1", utils.Ptr("Mozilla/5.0"), utils.Ptr("en-US"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if botUser.LastLoginIP != "10.0.0.1" {
		t.Fatalf("expected login ip 10.0.0.1, got %q", botUser.LastLoginIP)
	}

	if !botUser.LastLoginAt.After(initialLastLogin) {
		t.Fatalf("expected last login time to be updated")
	}

	if botUser.UpdatedAt == nil {
		t.Fatalf("expected non-nil updatedAt after login record")
	}
}

func TestBotUserRecordLoginValidation(t *testing.T) {
	t.Parallel()

	info, err := model.NewUserInfo("Ivan", nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	botUser, err := model.NewBotUser(100, 200, info, "127.0.0.1", nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = botUser.RecordLogin("bad-ip", nil, utils.Ptr("bad_lang"))
	if !errors.Is(err, model.ErrBotUserInvalidIP) {
		t.Fatalf("expected error %v, got %v", model.ErrBotUserInvalidIP, err)
	}
}

func TestBotUserLastModifiedAt(t *testing.T) {
	t.Parallel()

	info, err := model.NewUserInfo("Ivan", nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	botUser, err := model.NewBotUser(100, 200, info, "127.0.0.1", nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !botUser.LastModifiedAt().Equal(botUser.CreatedAt) {
		t.Fatalf("expected last modified to equal createdAt")
	}

	time.Sleep(2 * time.Millisecond)
	err = botUser.RecordLogin("10.0.0.1", nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if botUser.UpdatedAt == nil {
		t.Fatalf("expected non-nil updatedAt")
	}

	if !botUser.LastModifiedAt().Equal(*botUser.UpdatedAt) {
		t.Fatalf("expected last modified to equal updatedAt")
	}
}
