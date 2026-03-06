package model

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ulbwa/telegram-oidc-provider/pkg/utils"
)

var (
	ErrBotInvalidData          = errors.New("bot data is invalid")
	ErrBotInvalidID            = fmt.Errorf("%w: ID is invalid", ErrBotInvalidData)
	ErrBotInvalidName          = fmt.Errorf("%w: name is invalid", ErrBotInvalidData)
	ErrBotInvalidUsername      = fmt.Errorf("%w: username is invalid", ErrBotInvalidData)
	ErrBotInvalidToken         = fmt.Errorf("%w: token is invalid", ErrBotInvalidData)
	ErrBotInvalidOAuthClientID = fmt.Errorf("%w: OAuth client ID is invalid", ErrBotInvalidData)
	ErrBotInvalidCreatedAt     = fmt.Errorf("%w: created at is invalid", ErrBotInvalidData)
	ErrBotInvalidUpdatedAt     = fmt.Errorf("%w: updated at is invalid", ErrBotInvalidData)
	ErrBotUpdatedBeforeCreate  = fmt.Errorf("%w: updated at cannot be before created at", ErrBotInvalidUpdatedAt)
)

var (
	botUsernamePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]{2,31}$`)
	botNamePattern     = regexp.MustCompile(`^[\p{L}\p{N}_\- ]{1,128}$`)
)

const (
	maskedTokenVisibleTail = 4
	maskedTokenMinHidden   = 8
	maskedTokenPrefix      = "************"
)

// Bot is a rich domain entity for Telegram bot credentials used by OIDC flows.
//
// Fields are public for explicit model visibility, while domain methods enforce
// invariants without implicit normalization.
type Bot struct {
	ID            int64
	Name          string
	Username      string
	Token         string
	OAuthClientID *string
	CreatedAt     time.Time
	UpdatedAt     *time.Time
}

// NewBot creates a new Bot with immutable creation time.
// UpdatedAt is nil until first business update.
func NewBot(id int64, name, username, token string) (*Bot, error) {
	if err := validateBotID(id); err != nil {
		return nil, err
	}

	if err := validateBotName(name); err != nil {
		return nil, err
	}

	if err := validateBotUsername(username); err != nil {
		return nil, err
	}

	if err := validateBotToken(token); err != nil {
		return nil, err
	}

	return &Bot{
		ID:            id,
		Name:          name,
		Username:      username,
		Token:         token,
		OAuthClientID: nil,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     nil,
	}, nil
}

// RestoreBot restores Bot from persistence while validating all invariants.
func RestoreBot(
	id int64,
	name string,
	username string,
	token string,
	oauthClientID *string,
	createdAt time.Time,
	updatedAt *time.Time,
) (*Bot, error) {
	if err := validateBotID(id); err != nil {
		return nil, err
	}

	if err := validateBotName(name); err != nil {
		return nil, err
	}

	if err := validateBotUsername(username); err != nil {
		return nil, err
	}

	if err := validateBotToken(token); err != nil {
		return nil, err
	}

	if err := validateOAuthClientID(oauthClientID); err != nil {
		return nil, err
	}

	if createdAt.IsZero() {
		return nil, ErrBotInvalidCreatedAt
	}

	if err := validateUpdatedAt(updatedAt, createdAt); err != nil {
		return nil, err
	}

	return &Bot{
		ID:            id,
		Name:          name,
		Username:      username,
		Token:         token,
		OAuthClientID: utils.PtrClone(oauthClientID),
		CreatedAt:     createdAt,
		UpdatedAt:     utils.PtrClone(updatedAt),
	}, nil
}

// SetName sets bot display name and validates it.
func (b *Bot) SetName(name string) error {
	if b.Name == name {
		return nil
	}

	if err := validateBotName(name); err != nil {
		return err
	}

	b.Name = name

	if err := b.touch(); err != nil {
		return err
	}

	return nil
}

// SetUsername sets Telegram username and validates format.
func (b *Bot) SetUsername(username string) error {
	if b.Username == username {
		return nil
	}

	if err := validateBotUsername(username); err != nil {
		return err
	}

	b.Username = username

	if err := b.touch(); err != nil {
		return err
	}

	return nil
}

// SetToken sets bot token and validates it.
func (b *Bot) SetToken(token string) error {
	if b.Token == token {
		return nil
	}

	if err := validateBotToken(token); err != nil {
		return err
	}

	b.Token = token

	if err := b.touch(); err != nil {
		return err
	}

	return nil
}

// BindOAuthClient binds external OAuth client identifier (for example, Ory Hydra client ID).
func (b *Bot) BindOAuthClient(clientID string) error {
	if b.OAuthClientID != nil && *b.OAuthClientID == clientID {
		return nil
	}

	if err := validateOAuthClientID(utils.Ptr(clientID)); err != nil {
		return err
	}

	b.OAuthClientID = utils.Ptr(clientID)

	if err := b.touch(); err != nil {
		return err
	}

	return nil
}

// UnbindOAuthClient removes external OAuth client binding from bot.
func (b *Bot) UnbindOAuthClient() error {
	if b.OAuthClientID == nil {
		return nil
	}

	b.OAuthClientID = nil

	if err := b.touch(); err != nil {
		return err
	}

	return nil
}

// MaskedToken keeps bot ID prefix before ':' and masks the secret part,
// preserving only a very small trailing part of the secret.
// Bot ID before ':' is shown only when it is a valid positive int64.
func (b *Bot) MaskedToken() string {
	parts := strings.SplitN(b.Token, ":", 2)
	if len(parts) != 2 {
		return maskTokenPart(b.Token)
	}

	botID, ok := parseValidTokenBotID(parts[0])
	if !ok {
		return maskTokenPart(b.Token)
	}

	return botID + ":" + maskTokenPart(parts[1])
}

func parseValidTokenBotID(raw string) (string, bool) {
	if raw == "" {
		return "", false
	}

	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return "", false
	}

	return raw, true
}

func maskTokenPart(secret string) string {
	if len(secret) <= maskedTokenVisibleTail+maskedTokenMinHidden {
		return maskedTokenPrefix
	}

	return maskedTokenPrefix + secret[len(secret)-maskedTokenVisibleTail:]
}

// LastModifiedAt returns the latest change timestamp for this entity.
func (b *Bot) LastModifiedAt() time.Time {
	if b.UpdatedAt == nil {
		return b.CreatedAt
	}

	if b.UpdatedAt.After(b.CreatedAt) {
		return *b.UpdatedAt
	}

	return b.CreatedAt
}

// UpdateProfile updates bot profile fields and update timestamp atomically.
func (b *Bot) UpdateProfile(name, username string) error {
	if b.Name == name && b.Username == username {
		return nil
	}

	if b.Name != name {
		if err := validateBotName(name); err != nil {
			return err
		}
	}

	if b.Username != username {
		if err := validateBotUsername(username); err != nil {
			return err
		}
	}

	b.Name = name
	b.Username = username

	if err := b.touch(); err != nil {
		return err
	}

	return nil
}

func (b *Bot) touch() error {
	now := time.Now().UTC()
	if now.Before(b.CreatedAt) {
		return ErrBotUpdatedBeforeCreate
	}

	b.UpdatedAt = &now

	return nil
}

func validateBotID(id int64) error {
	if id <= 0 {
		return ErrBotInvalidID
	}

	return nil
}

func validateBotName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" || name != trimmed {
		return ErrBotInvalidName
	}

	if !botNamePattern.MatchString(name) {
		return ErrBotInvalidName
	}

	return nil
}

func validateBotUsername(username string) error {
	trimmed := strings.TrimSpace(username)
	if trimmed == "" || username != trimmed {
		return ErrBotInvalidUsername
	}

	if !botUsernamePattern.MatchString(username) {
		return ErrBotInvalidUsername
	}

	return nil
}

func validateBotToken(token string) error {
	trimmed := strings.TrimSpace(token)
	if trimmed == "" || token != trimmed {
		return ErrBotInvalidToken
	}

	return nil
}

func validateOAuthClientID(clientID *string) error {
	if clientID == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*clientID)
	if trimmed == "" || *clientID != trimmed {
		return ErrBotInvalidOAuthClientID
	}

	return nil
}

func validateUpdatedAt(updatedAt *time.Time, createdAt time.Time) error {
	if updatedAt == nil {
		return nil
	}

	if updatedAt.IsZero() {
		return ErrBotInvalidUpdatedAt
	}

	if updatedAt.Before(createdAt) {
		return ErrBotUpdatedBeforeCreate
	}

	return nil
}
