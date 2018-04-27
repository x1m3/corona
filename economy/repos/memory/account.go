package memoryRepo

import (
	"github.com/nu7hatch/gouuid"
	"github.com/x1m3/elixir/economy"
	"sync"
	"time"
)

type AccountRepo struct {
	sync.RWMutex
	accounts map[uuid.UUID]*economy.Account
}

func NewAccountRepo() *AccountRepo {
	return &AccountRepo{accounts: make(map[uuid.UUID]*economy.Account)}
}

func (r *AccountRepo) Add(acc *economy.Account) error {
	r.Lock()
	defer r.Unlock()

	if _, found := r.accounts[acc.ID()]; found {
		return economy.Err_DuplicatedAccount
	}
	r.accounts[acc.ID()] = acc
	return nil
}

func (r *AccountRepo) Get(ID *uuid.UUID) (acc *economy.Account, err error) {
	var found bool

	r.RLock()
	defer r.RUnlock()

	if acc, found = r.accounts[*ID]; !found {
		return nil, economy.Err_AccountNotFound
	}
	return acc, nil
}

func (r *AccountRepo) ApplyMovement(ID *uuid.UUID, movement *economy.AccountMovement) (*economy.Money, error) {
	r.Lock()
	defer r.Unlock()

	return r.applyMovement(ID, movement)
}

func (r *AccountRepo) applyMovement(ID *uuid.UUID, movement *economy.AccountMovement) (*economy.Money, error) {
	var acc *economy.Account
	var found bool

	if acc, found = r.accounts[*ID]; !found {
		return nil, economy.Err_AccountNotFound
	}

	return acc.Add(movement)
}

func (r *AccountRepo) ApplyMovementBetweenAccounts(fromID *uuid.UUID, toID *uuid.UUID, m *economy.Money) error {
	r.Lock()
	defer r.Unlock()

	_, err := r.applyMovement(fromID, economy.NewAccountMovement(economy.NewMoney(-m.Amount(), m.Currency()), "", time.Now()))
	if err!=nil {
		return err
	}
	_, err = r.applyMovement(toID, economy.NewAccountMovement(m, "", time.Now()))
	if err!=nil {
		return err
	}
	return nil
}
