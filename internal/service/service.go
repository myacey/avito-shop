package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/myacey/avito-shop/internal/apperror"
	"github.com/myacey/avito-shop/internal/hasher"
	"github.com/myacey/avito-shop/internal/jwttoken"
	"github.com/myacey/avito-shop/internal/models"
	"github.com/myacey/avito-shop/internal/repository"
)

var (
	ErrInvalidPassword = errors.New("invalid password")
	ErrNotEnoughMoney  = errors.New("not enough money on account")
	ErrInvalidToken    = errors.New("invalid auth token")
)

const sessionKeyTTL = time.Duration(24 * time.Hour)

type Interface interface {
	// /api/auth
	AuthorizeUser(c context.Context, username, password string) (string, error)

	CheckAuthToken(c context.Context, token string) (string, error)

	// /api/info
	GetFullUserInfo(c context.Context, username string) (*models.User, error)

	// /api/sendCoin
	SendCoin(c context.Context, fromUsername string, toUsername string, amount int32) error

	// /api/buy/{item}
	BuyItem(c context.Context, username string, itemName string) error
}

type Service struct {
	dbConn *sql.DB

	userRepo      repository.UserRepository
	transferRepo  repository.TransferRepository
	inventoryRepo repository.InventoryRepository
	storeRepo     repository.StoreRepository

	tokenMaker  jwttoken.TokenMakerInterface
	sessionRepo repository.SessionRepository

	hasher hasher.Hasher
}

func NewService(
	dbConn *sql.DB,
	ur repository.UserRepository,
	tr repository.TransferRepository,
	ir repository.InventoryRepository,
	sr repository.StoreRepository,
	rsr repository.SessionRepository,
	tokMaker jwttoken.TokenMakerInterface,
	hasher hasher.Hasher,
) Interface {
	return &Service{
		dbConn:        dbConn,
		userRepo:      ur,
		transferRepo:  tr,
		inventoryRepo: ir,
		storeRepo:     sr,
		sessionRepo:   rsr,
		tokenMaker:    tokMaker,
		hasher:        hasher,
	}
}

func (s *Service) createUser(c context.Context, username, password string) (string, error) {
	hashedPassword, err := s.hasher.Generate(password)
	if err != nil {
		if errors.Is(err, hasher.ErrToLong) {
			return "", apperror.NewBadReq("password too long", err)
		}
		return "", apperror.NewInternal("failed to generate password hash", err)
	}

	_, err = s.userRepo.CreateUser(c, username, string(hashedPassword))
	if err != nil {
		return "", apperror.NewInternal("failed to create user", err)
	}

	// create auth token
	token, err := s.tokenMaker.CreateToken(username)
	if err != nil {
		return "", apperror.NewInternal("failed to create token for user", err)
	}

	if err = s.sessionRepo.CreateToken(c, username, token, sessionKeyTTL); err != nil {
		return "", apperror.NewInternal("failed to save token", err)
	}
	return token, nil
}

// Authorization checks user credentials, creates new dbUser if needed.
func (s *Service) AuthorizeUser(c context.Context, username, password string) (string, error) {
	dbUsr, err := s.userRepo.GetUser(c, username)

	// unknown error
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return "", apperror.NewInternal("failed to get user", err)
	}

	// user dont exists -> generate new one
	if errors.Is(err, repository.ErrUserNotFound) {
		return s.createUser(c, username, password)
	}

	// User found, check him
	// check password
	if err = s.hasher.Compare(dbUsr.Password, password); err != nil {
		if errors.Is(err, hasher.ErrDontCompare) {
			return "", apperror.NewNotFound("user not found", err)
		}
		return "", apperror.NewInternal("failed to compare passwords", err)
	}

	// generate token
	newToken, err := s.tokenMaker.CreateToken(username)
	if err != nil {
		return "", apperror.NewInternal("failed to create token", err)
	}

	// save token
	if err = s.sessionRepo.CreateToken(c, username, newToken, sessionKeyTTL); err != nil {
		return "", apperror.NewInternal("failed to save token", err)
	}

	return newToken, nil
}

// CheckAuthToken extracts username from jwt payload, gets a dbToken
// from redis and checks if providen "token==dbToken".
func (s *Service) CheckAuthToken(c context.Context, token string) (string, error) {
	username, err := s.tokenMaker.VerifyToken(token)
	if err != nil {
		switch {
		case errors.Is(err, jwttoken.ErrInvalidKey):
			return "", apperror.NewUnauthorized(jwttoken.ErrInvalidKey.Error(), nil)
		case errors.Is(err, jwttoken.ErrTokenExpired):
			return "", apperror.NewUnauthorized(jwttoken.ErrTokenExpired.Error(), nil)
		default:
			return "", apperror.NewInternal("failed to verify token", err)
		}
	}

	dbToken, err := s.sessionRepo.GetToken(c, username)
	if err != nil {
		if errors.Is(err, repository.ErrTokenNotFound) {
			return "", apperror.NewUnauthorized("invalid token", err)
		}
		return "", apperror.NewInternal("failed to find token", err)
	}

	if dbToken != token {
		return "", apperror.NewUnauthorized("invalid token", ErrInvalidToken)
	}

	return username, nil
}

type IncomeEntry struct {
	FromUser string `json:"fromUser"`
	Amount   int32  `json:"amount"`
}

type OutcomeEntry struct {
	ToUser string `json:"toUser"`
	Amount int32  `json:"amount"`
}

// fillInventory is helper func to fill user's inventory.
// returns apperror.
func (s *Service) fillInventory(c context.Context, usr *models.User, userID int32) error {
	inventory, err := s.inventoryRepo.GetInventory(c, userID)
	if err != nil && !errors.Is(err, repository.ErrNoInventoryItems) {
		return apperror.NewInternal("failed to get user inventory", err)
	}

	// fill inventory
	usr.Inventory = make([]*models.InventoryItem, len(inventory))
	for i, v := range inventory {
		usr.Inventory[i] = &models.InventoryItem{
			Type:     v.ItemType,
			Quantity: v.Quantity,
		}
	}

	return nil
}

// fillEntries is helper func to fill user's entries list (recoieved, send entries).
// returns apperror.
func (s *Service) fillEntries(c context.Context, usr *models.User, username string) error {
	// fill entries
	entries, err := s.transferRepo.GetTransfersWithUser(c, username)
	if err != nil && !errors.Is(err, repository.ErrNoTransfers) {
		return apperror.NewInternal("failed to find transactions", err)
	}

	m := map[string]interface{}{}
	income := make([]*IncomeEntry, 0, len(entries))
	outcome := make([]*OutcomeEntry, 0, len(entries))
	for _, v := range entries {
		if v.ToUsername == username {
			income = append(income, &IncomeEntry{FromUser: v.FromUsername, Amount: v.Amount})
		} else if v.FromUsername == username {
			outcome = append(outcome, &OutcomeEntry{ToUser: v.ToUsername, Amount: v.Amount})
		}
	}
	m["received"] = income
	m["sent"] = outcome
	usr.EntryHistory = m

	return nil
}

// GetFullUserInfo fetchs main usr info, transaction hisotry and items in inventory.
func (s *Service) GetFullUserInfo(c context.Context, username string) (*models.User, error) {
	dbUsr, err := s.userRepo.GetUser(c, username)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, apperror.NewNotFound("user not found", err)
		}
		return nil, apperror.NewInternal("failed to get user", err)
	}

	usr := &models.User{
		ID:           dbUsr.UserID,
		Username:     dbUsr.Username,
		Coins:        dbUsr.Coins,
		Inventory:    nil,
		EntryHistory: nil,
	}

	if err = s.fillInventory(c, usr, dbUsr.UserID); err != nil {
		return nil, err
	}

	if err = s.fillEntries(c, usr, username); err != nil {
		return nil, err
	}

	return usr, nil
}

// SendCoins runs a transacion to create new transaction and update user's coins.
func (s *Service) SendCoin(c context.Context, fromUsername string, toUsername string, amount int32) error {
	if amount <= 0 {
		return apperror.NewBadReq("send coins amont must be positive", nil)
	}

	tx, err := s.dbConn.BeginTx(c, nil)
	if err != nil {
		return apperror.NewInternal("failed to send coins", err)
	}
	defer tx.Rollback()

	_, err = s.userRepo.UpdateTwoUsersBalance(c, fromUsername, toUsername, amount)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return apperror.NewNotFound(fmt.Sprintf("users not found: %s, %s", fromUsername, toUsername), err)
		}
		return apperror.NewInternal("failed to make money transaction", err)
	}
	_, err = s.transferRepo.CreateMoneyTransfer(c, fromUsername, toUsername, amount)
	if err != nil {
		return apperror.NewInternal("failed to create transfer", err)
	}

	return tx.Commit()
}

// BuyItem adds an item to user's inventory.
func (s *Service) BuyItem(c context.Context, username, itemName string) error {
	itemToBuy, err := s.storeRepo.GetItemInfo(c, itemName)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidItemName) {
			return apperror.NewBadReq("invalid item name", err)
		}
		return apperror.NewInternal("failed to get item info", err)
	}

	tx, err := s.dbConn.BeginTx(c, nil)
	if err != nil {
		return apperror.NewInternal("failed to buy item", err)
	}
	defer tx.Rollback()

	dbUsr, err := s.userRepo.GetUserForUpdate(c, username)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return apperror.NewNotFound("user not found", err)
		}
		return apperror.NewInternal("failed to get user", err)
	}

	newCoinsCount := dbUsr.Coins - int32(itemToBuy.ItemPrice)
	if newCoinsCount < 0 {
		return apperror.NewBadReq("not enough money", ErrNotEnoughMoney)
	}

	_, err = s.userRepo.UpdateBalance(c, dbUsr.UserID, newCoinsCount)
	if err != nil {
		return apperror.NewInternal("failed to update balance", err)
	}

	err = s.inventoryRepo.AddItemToInventory(c, dbUsr.UserID, itemName)
	if err != nil {
		return apperror.NewInternal("failed to add item to inventory", err)
	}

	return tx.Commit()
}
