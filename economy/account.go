package economy

import "github.com/nu7hatch/gouuid"

type AccountRepo interface {
	Add(acc *Account) error
	Get(ID *uuid.UUID) (*Account, error)
	ApplyMovement(ID *uuid.UUID, movement *AccountMovement) (*Money, error)
	ApplyMovementBetweenAccounts(fromID *uuid.UUID, toID *uuid.UUID, m *Money) error
}


type Account struct {
	id        *uuid.UUID
	cash      *Money
}

func NewAccount(id *uuid.UUID, c *Money) *Account {
	return &Account{
		id:        id,
		cash:      c,
	}
}

func (a *Account) ID() uuid.UUID {
	return *a.id
}

func (a *Account) Add(m *AccountMovement) (*Money, error) {
	var err error

	a.cash, err = a.cash.Add(m.Money)
	if err != nil {
		return nil, err
	}

	return a.cash.Copy(), nil
}

func (a *Account) Balance() *Money {
	return a.cash.Copy()
}