package model

import (
	"net/netip"
	"net/url"
	"regexp"
	"strings"
	"time"

	errs "github.com/ulbwa/telegram-oidc-provider/internal/errors"
	"github.com/ulbwa/telegram-oidc-provider/pkg/utils"
)

var (
	botUserNamePattern     = regexp.MustCompile(`^[\p{L}\p{N}_\- ]{1,128}$`)
	botUserUsernamePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]{2,31}$`)
	bcp47Pattern           = regexp.MustCompile(`^[A-Za-z]{2,3}(-[A-Za-z0-9]{2,8})*$`)
)

// UserInfo stores Telegram user profile fields.
type UserInfo struct {
	FirstName string
	LastName  *string
	Username  *string
	PhotoURL  *string
	IsPremium *bool
}

func NewUserInfo(firstName string, lastName, username, photoURL *string, isPremium *bool) (UserInfo, error) {
	if err := validateUserFirstName(firstName); err != nil {
		return UserInfo{}, err
	}

	if err := validateUserLastName(lastName); err != nil {
		return UserInfo{}, err
	}

	if err := validateUserUsername(username); err != nil {
		return UserInfo{}, err
	}

	if err := validateUserPhotoURL(photoURL); err != nil {
		return UserInfo{}, err
	}

	return UserInfo{
		FirstName: firstName,
		LastName:  utils.PtrClone(lastName),
		Username:  utils.PtrClone(username),
		PhotoURL:  utils.PtrClone(photoURL),
		IsPremium: utils.PtrClone(isPremium),
	}, nil
}

func (i UserInfo) Equal(other UserInfo) bool {
	return i.FirstName == other.FirstName &&
		utils.PtrEqual(i.LastName, other.LastName) &&
		utils.PtrEqual(i.Username, other.Username) &&
		utils.PtrEqual(i.PhotoURL, other.PhotoURL) &&
		utils.PtrEqual(i.IsPremium, other.IsPremium)
}

func (i UserInfo) clone() UserInfo {
	return UserInfo{
		FirstName: i.FirstName,
		LastName:  utils.PtrClone(i.LastName),
		Username:  utils.PtrClone(i.Username),
		PhotoURL:  utils.PtrClone(i.PhotoURL),
		IsPremium: utils.PtrClone(i.IsPremium),
	}
}

// BotUser stores user data scoped by a concrete Telegram bot.
type BotUser struct {
	BotID              int64
	UserID             int64
	Info               UserInfo
	LastLoginIP        string
	LastLoginUserAgent *string
	LastLoginLanguage  *string
	LastLoginAt        time.Time
	CreatedAt          time.Time
	UpdatedAt          *time.Time
}

func NewBotUser(botID, userID int64, info UserInfo, loginIP string, userAgent, language *string) (*BotUser, error) {
	if err := validateBotUserIDs(botID, userID); err != nil {
		return nil, err
	}

	if err := validateUserInfo(info); err != nil {
		return nil, err
	}

	if err := validateLoginMetadata(loginIP, userAgent, language); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	if now.IsZero() {
		return nil, errs.ErrBotUserCreatedAtInvalid
	}

	return &BotUser{
		BotID:              botID,
		UserID:             userID,
		Info:               info.clone(),
		LastLoginIP:        loginIP,
		LastLoginUserAgent: utils.PtrClone(userAgent),
		LastLoginLanguage:  utils.PtrClone(language),
		LastLoginAt:        now,
		CreatedAt:          now,
		UpdatedAt:          nil,
	}, nil
}

func (u *BotUser) UpdateInfo(info UserInfo) error {
	if u.Info.Equal(info) {
		return nil
	}

	if err := validateUserInfo(info); err != nil {
		return err
	}

	u.Info = info.clone()

	return u.touch()
}

func (u *BotUser) RecordLogin(loginIP string, userAgent, language *string) error {
	if err := validateLoginMetadata(loginIP, userAgent, language); err != nil {
		return err
	}

	now := time.Now().UTC()
	if now.Before(u.CreatedAt) {
		return errs.ErrBotUserLastLoginBeforeCreate
	}

	u.LastLoginIP = loginIP
	u.LastLoginUserAgent = utils.PtrClone(userAgent)
	u.LastLoginLanguage = utils.PtrClone(language)
	u.LastLoginAt = now

	return u.touch()
}

func (u *BotUser) LastModifiedAt() time.Time {
	if u.UpdatedAt == nil {
		return u.CreatedAt
	}

	if u.UpdatedAt.After(u.CreatedAt) {
		return *u.UpdatedAt
	}

	return u.CreatedAt
}

func (u *BotUser) touch() error {
	now := time.Now().UTC()
	if now.Before(u.CreatedAt) {
		return errs.ErrBotUserUpdatedBeforeCreate
	}

	u.UpdatedAt = &now

	return nil
}

func validateBotUserIDs(botID, userID int64) error {
	if botID <= 0 {
		return errs.ErrBotUserInvalidBotID
	}

	if userID <= 0 {
		return errs.ErrBotUserInvalidUserID
	}

	return nil
}

func validateUserInfo(info UserInfo) error {
	if err := validateUserFirstName(info.FirstName); err != nil {
		return err
	}

	if err := validateUserLastName(info.LastName); err != nil {
		return err
	}

	if err := validateUserUsername(info.Username); err != nil {
		return err
	}

	if err := validateUserPhotoURL(info.PhotoURL); err != nil {
		return err
	}

	return nil
}

func validateUserFirstName(firstName string) error {
	if strings.TrimSpace(firstName) == "" {
		return errs.ErrUserFirstNameRequired
	}

	if firstName != strings.TrimSpace(firstName) {
		return errs.ErrUserFirstNameHasOuterSpaces
	}

	if !botUserNamePattern.MatchString(firstName) {
		return errs.ErrUserFirstNameInvalid
	}

	return nil
}

func validateUserLastName(lastName *string) error {
	if lastName == nil {
		return nil
	}

	if strings.TrimSpace(*lastName) == "" || *lastName != strings.TrimSpace(*lastName) {
		return errs.ErrUserLastNameHasOuterSpaces
	}

	if !botUserNamePattern.MatchString(*lastName) {
		return errs.ErrUserLastNameInvalid
	}

	return nil
}

func validateUserUsername(username *string) error {
	if username == nil {
		return nil
	}

	if strings.TrimSpace(*username) == "" || *username != strings.TrimSpace(*username) {
		return errs.ErrUserUsernameHasOuterSpaces
	}

	if !botUserUsernamePattern.MatchString(*username) {
		return errs.ErrUserUsernameInvalid
	}

	return nil
}

func validateUserPhotoURL(photoURL *string) error {
	if photoURL == nil {
		return nil
	}

	if strings.TrimSpace(*photoURL) == "" || *photoURL != strings.TrimSpace(*photoURL) {
		return errs.ErrUserPhotoURLInvalid
	}

	parsedURL, err := url.ParseRequestURI(*photoURL)
	if err != nil {
		return errs.ErrUserPhotoURLInvalid
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return errs.ErrUserPhotoURLInvalid
	}

	return nil
}

func validateLoginMetadata(loginIP string, userAgent, language *string) error {
	if strings.TrimSpace(loginIP) == "" {
		return errs.ErrBotUserLoginIPRequired
	}

	if loginIP != strings.TrimSpace(loginIP) {
		return errs.ErrBotUserLoginIPInvalid
	}

	if _, err := netip.ParseAddr(loginIP); err != nil {
		return errs.ErrBotUserLoginIPInvalid
	}

	if err := validateBotUserUserAgent(userAgent); err != nil {
		return err
	}

	if err := validateBotUserLanguage(language); err != nil {
		return err
	}

	return nil
}

func validateBotUserUserAgent(userAgent *string) error {
	if userAgent == nil {
		return nil
	}

	if strings.TrimSpace(*userAgent) == "" {
		return errs.ErrBotUserUserAgentInvalid
	}

	if *userAgent != strings.TrimSpace(*userAgent) {
		return errs.ErrBotUserUserAgentHasOuterSpaces
	}

	if len(*userAgent) > 2048 {
		return errs.ErrBotUserUserAgentInvalid
	}

	return nil
}

func validateBotUserLanguage(language *string) error {
	if language == nil {
		return nil
	}

	if strings.TrimSpace(*language) == "" {
		return errs.ErrBotUserLanguageInvalid
	}

	if *language != strings.TrimSpace(*language) {
		return errs.ErrBotUserLanguageHasOuterSpaces
	}

	if !bcp47Pattern.MatchString(*language) {
		return errs.ErrBotUserLanguageInvalid
	}

	return nil
}
