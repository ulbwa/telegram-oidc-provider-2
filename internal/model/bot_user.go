package model

import (
	"errors"
	"fmt"
	"net/netip"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/ulbwa/telegram-oidc-provider/pkg/utils"
)

var (
	ErrUserInvalidData      = errors.New("user is invalid")
	ErrUserInvalidID        = fmt.Errorf("%w: ID is invalid", ErrUserInvalidData)
	ErrUserInvalidFirstName = fmt.Errorf("%w: first name is invalid", ErrUserInvalidData)
	ErrUserInvalidLastName  = fmt.Errorf("%w: last name is invalid", ErrUserInvalidData)
	ErrUserInvalidUsername  = fmt.Errorf("%w: username is invalid", ErrUserInvalidData)
	ErrUserInvalidPhotoURL  = fmt.Errorf("%w: photo URL is invalid", ErrUserInvalidData)

	ErrBotUserInvalidData         = errors.New("bot user is invalid")
	ErrBotUserInvalidBotID        = fmt.Errorf("%w: bot ID is invalid", ErrBotUserInvalidData)
	ErrBotUserInvalidUserID       = fmt.Errorf("%w: user ID is invalid", ErrBotUserInvalidData)
	ErrBotUserInvalidIP           = fmt.Errorf("%w: login IP is invalid", ErrBotUserInvalidData)
	ErrBotUserInvalidUserAgent    = fmt.Errorf("%w: login user agent is invalid", ErrBotUserInvalidData)
	ErrBotUserInvalidLanguage     = fmt.Errorf("%w: login language is invalid", ErrBotUserInvalidData)
	ErrBotUserInvalidCreatedAt    = fmt.Errorf("%w: created at is invalid", ErrBotUserInvalidData)
	ErrBotUserInvalidUpdatedAt    = fmt.Errorf("%w: updated at is invalid", ErrBotUserInvalidData)
	ErrBotUserUpdatedBeforeCreate = fmt.Errorf("%w: updated at cannot be before created at", ErrBotUserInvalidUpdatedAt)
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

	if err := validateBotUserLoginMetadata(loginIP, userAgent, language); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	if now.IsZero() {
		return nil, ErrBotUserInvalidCreatedAt
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

func RestoreBotUser(
	botID, userID int64,
	info UserInfo,
	loginIP string,
	userAgent, language *string,
	lastLoginAt time.Time,
	createdAt time.Time,
	updatedAt *time.Time,
) (*BotUser, error) {
	if err := validateBotUserIDs(botID, userID); err != nil {
		return nil, err
	}

	if err := validateUserInfo(info); err != nil {
		return nil, err
	}

	if err := validateBotUserLoginMetadata(loginIP, userAgent, language); err != nil {
		return nil, err
	}

	if lastLoginAt.IsZero() {
		return nil, ErrBotUserInvalidCreatedAt
	}

	if createdAt.IsZero() {
		return nil, ErrBotUserInvalidCreatedAt
	}

	if err := validateBotUserUpdatedAt(updatedAt, createdAt); err != nil {
		return nil, err
	}

	return &BotUser{
		BotID:              botID,
		UserID:             userID,
		Info:               info.clone(),
		LastLoginIP:        loginIP,
		LastLoginUserAgent: utils.PtrClone(userAgent),
		LastLoginLanguage:  utils.PtrClone(language),
		LastLoginAt:        lastLoginAt,
		CreatedAt:          createdAt,
		UpdatedAt:          utils.PtrClone(updatedAt),
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
	if err := validateBotUserLoginMetadata(loginIP, userAgent, language); err != nil {
		return err
	}

	now := time.Now().UTC()
	if now.Before(u.CreatedAt) {
		return ErrBotUserUpdatedBeforeCreate
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
		return ErrBotUserUpdatedBeforeCreate
	}

	u.UpdatedAt = &now

	return nil
}

func validateBotUserIDs(botID, userID int64) error {
	if err := validateBotID(botID); err != nil {
		return ErrBotUserInvalidBotID
	}

	if err := validateUserID(userID); err != nil {
		return ErrBotUserInvalidUserID
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

func validateUserID(id int64) error {
	if id <= 0 {
		return ErrUserInvalidID
	}

	return nil
}

func validateUserFirstName(firstName string) error {
	trimmed := strings.TrimSpace(firstName)
	if trimmed == "" || firstName != trimmed {
		return ErrUserInvalidFirstName
	}

	if !botUserNamePattern.MatchString(firstName) {
		return ErrUserInvalidFirstName
	}

	return nil
}

func validateUserLastName(lastName *string) error {
	if lastName == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*lastName)
	if trimmed == "" || *lastName != trimmed {
		return ErrUserInvalidLastName
	}

	if !botUserNamePattern.MatchString(*lastName) {
		return ErrUserInvalidLastName
	}

	return nil
}

func validateUserUsername(username *string) error {
	if username == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*username)
	if trimmed == "" || *username != trimmed {
		return ErrUserInvalidUsername
	}

	if !botUserUsernamePattern.MatchString(*username) {
		return ErrUserInvalidUsername
	}

	return nil
}

func validateUserPhotoURL(photoURL *string) error {
	if photoURL == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*photoURL)
	if trimmed == "" || *photoURL != trimmed {
		return ErrUserInvalidPhotoURL
	}

	parsedURL, err := url.ParseRequestURI(*photoURL)
	if err != nil {
		return ErrUserInvalidPhotoURL
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return ErrUserInvalidPhotoURL
	}

	return nil
}

func validateBotUserLoginMetadata(loginIP string, userAgent, language *string) error {
	if err := validateBotUserIP(loginIP); err != nil {
		return err
	}

	if err := validateBotUserUserAgent(userAgent); err != nil {
		return err
	}

	if err := validateBotUserLanguage(language); err != nil {
		return err
	}

	return nil
}

func validateBotUserIP(loginIP string) error {
	trimmed := strings.TrimSpace(loginIP)
	if trimmed == "" || loginIP != trimmed {
		return ErrBotUserInvalidIP
	}

	if _, err := netip.ParseAddr(loginIP); err != nil {
		return ErrBotUserInvalidIP
	}

	return nil
}

func validateBotUserUserAgent(userAgent *string) error {
	if userAgent == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*userAgent)
	if trimmed == "" || *userAgent != trimmed {
		return ErrBotUserInvalidUserAgent
	}

	if len(*userAgent) > 2048 {
		return ErrBotUserInvalidUserAgent
	}

	return nil
}

func validateBotUserLanguage(language *string) error {
	if language == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*language)
	if trimmed == "" || *language != trimmed {
		return ErrBotUserInvalidLanguage
	}

	if !bcp47Pattern.MatchString(*language) {
		return ErrBotUserInvalidLanguage
	}

	return nil
}

func validateBotUserUpdatedAt(updatedAt *time.Time, createdAt time.Time) error {
	if updatedAt == nil {
		return nil
	}

	if updatedAt.IsZero() {
		return ErrBotUserInvalidUpdatedAt
	}

	if updatedAt.Before(createdAt) {
		return ErrBotUserUpdatedBeforeCreate
	}

	return nil
}
