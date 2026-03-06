package model

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	errs "github.com/ulbwa/telegram-oidc-provider/internal/errors"
	"github.com/ulbwa/telegram-oidc-provider/pkg/utils"
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
func NewBot(id int64, name string, username string, token string) (*Bot, error) {
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

	now := time.Now().UTC()
	if now.IsZero() {
		return nil, errs.ErrBotCreatedAtInvalid
	}

	return &Bot{
		ID:            id,
		Name:          name,
		Username:      username,
		Token:         token,
		OAuthClientID: nil,
		CreatedAt:     now,
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
		return nil, errs.ErrBotCreatedAtInvalid
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
		UpdatedAt:     updatedAt,
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
func (b *Bot) UpdateProfile(name string, username string) error {
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
		return errs.ErrBotUpdatedBeforeCreate
	}

	b.UpdatedAt = &now

	return nil
}

func validateBotID(id int64) error {
	if id <= 0 {
		return errs.ErrBotInvalidID
	}

	return nil
}

func validateBotName(name string) error {
	if strings.TrimSpace(name) == "" {
		return errs.ErrBotNameRequired
	}

	if name != strings.TrimSpace(name) {
		return errs.ErrBotNameHasOuterSpaces
	}

	if !botNamePattern.MatchString(name) {
		return errs.ErrBotNameInvalid
	}

	return nil
}

func validateBotUsername(username string) error {
	if strings.TrimSpace(username) == "" {
		return errs.ErrBotUsernameRequired
	}

	if username != strings.TrimSpace(username) {
		return errs.ErrBotUsernameOuterSpaces
	}

	if !botUsernamePattern.MatchString(username) {
		return errs.ErrBotUsernameInvalid
	}

	return nil
}

func validateBotToken(token string) error {
	if strings.TrimSpace(token) == "" {
		return errs.ErrBotTokenRequired
	}

	if token != strings.TrimSpace(token) {
		return errs.ErrBotTokenOuterSpaces
	}

	return nil
}

func validateOAuthClientID(clientID *string) error {
	if clientID == nil {
		return nil
	}

	if strings.TrimSpace(*clientID) == "" {
		return errs.ErrBotOAuthClientIDInvalid
	}

	if *clientID != strings.TrimSpace(*clientID) {
		return errs.ErrBotOAuthClientIDOuterSpaces
	}

	return nil
}

func validateUpdatedAt(updatedAt *time.Time, createdAt time.Time) error {
	if updatedAt == nil {
		return nil
	}

	if updatedAt.IsZero() {
		return errs.ErrBotUpdatedAtInvalid
	}

	if updatedAt.Before(createdAt) {
		return errs.ErrBotUpdatedBeforeCreate
	}

	return nil
}
