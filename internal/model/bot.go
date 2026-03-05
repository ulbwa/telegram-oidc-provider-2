package model

import (
	"regexp"
	"strings"
	"time"

	errs "github.com/ulbwa/telegram-oidc-provider/internal/errors"
	"github.com/ulbwa/telegram-oidc-provider/pkg/utils"
)

var (
	botUsernamePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]{2,31}$`)
	botNamePattern     = regexp.MustCompile(`^[\p{L}\p{N}_\- ]{1,128}$`)
)

// Bot is a rich domain entity for Telegram bot credentials used by OIDC flows.
//
// Fields are public for explicit model visibility, while domain methods enforce
// invariants without implicit normalization.
type Bot struct {
	ID            int64
	Name          string
	Username      *string
	Token         string
	OAuthClientID *string
	CreatedAt     time.Time
	UpdatedAt     *time.Time
}

// NewBot creates a new Bot with immutable creation time.
// UpdatedAt is nil until first business update.
func NewBot(id int64, name string, username *string, token string) (*Bot, error) {
	if err := validateID(id); err != nil {
		return nil, err
	}

	if err := validateName(name); err != nil {
		return nil, err
	}

	if err := validateUsername(username); err != nil {
		return nil, err
	}

	if err := validateToken(token); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	if now.IsZero() {
		return nil, errs.ErrBotCreatedAtInvalid
	}

	return &Bot{
		ID:            id,
		Name:          name,
		Username:      utils.PtrClone(username),
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
	username *string,
	token string,
	oauthClientID *string,
	createdAt time.Time,
	updatedAt *time.Time,
) (*Bot, error) {
	if err := validateID(id); err != nil {
		return nil, err
	}

	if err := validateName(name); err != nil {
		return nil, err
	}

	if err := validateUsername(username); err != nil {
		return nil, err
	}

	if err := validateToken(token); err != nil {
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
		Username:      utils.PtrClone(username),
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

	if err := validateName(name); err != nil {
		return err
	}

	b.Name = name

	if err := b.touch(); err != nil {
		return err
	}

	return nil
}

// SetUsername sets Telegram username and validates format. Nil means no username.
func (b *Bot) SetUsername(username *string) error {
	if utils.PtrEqual(b.Username, username) {
		return nil
	}

	if err := validateUsername(username); err != nil {
		return err
	}

	b.Username = utils.PtrClone(username)

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

	if err := validateToken(token); err != nil {
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
// preserving only the last 6 secret characters.
func (b *Bot) MaskedToken() string {
	parts := strings.SplitN(b.Token, ":", 2)
	if len(parts) != 2 || parts[0] == "" {
		tokenLength := len(b.Token)
		if tokenLength <= 6 {
			return "****"
		}

		return strings.Repeat("*", tokenLength-6) + b.Token[tokenLength-6:]
	}

	secret := parts[1]
	if len(secret) <= 6 {
		return parts[0] + ":****" + secret
	}

	maskedSecret := strings.Repeat("*", len(secret)-6) + secret[len(secret)-6:]
	if maskedSecret == "" {
		return "****"
	}

	return parts[0] + ":" + maskedSecret
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
func (b *Bot) UpdateProfile(name string, username *string) error {
	if b.Name == name && utils.PtrEqual(b.Username, username) {
		return nil
	}

	if b.Name != name {
		if err := validateName(name); err != nil {
			return err
		}
	}

	if !utils.PtrEqual(b.Username, username) {
		if err := validateUsername(username); err != nil {
			return err
		}
	}

	b.Name = name
	b.Username = utils.PtrClone(username)

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

func validateID(id int64) error {
	if id <= 0 {
		return errs.ErrBotInvalidID
	}

	return nil
}

func validateName(name string) error {
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

func validateUsername(username *string) error {
	if username == nil {
		return nil
	}

	if strings.TrimSpace(*username) == "" {
		return errs.ErrBotUsernameInvalid
	}

	if *username != strings.TrimSpace(*username) {
		return errs.ErrBotUsernameOuterSpaces
	}

	if !botUsernamePattern.MatchString(*username) {
		return errs.ErrBotUsernameInvalid
	}

	return nil
}

func validateToken(token string) error {
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
