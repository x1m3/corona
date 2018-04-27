package test

import (
	"testing"
	"github.com/x1m3/elixir/economy"
	"github.com/x1m3/elixir/economy/repos/memory"
	"github.com/nu7hatch/gouuid"
	"time"
)

func TestAccountRepo(t *testing.T) {
	repos := make(map[string]economy.AccountRepo)
	repos["memory"] = memoryRepo.NewAccountRepo()

	for name, repo := range repos {
		testduplicatedAccount(name, repo, t)
		testAccountNotFound(name, repo, t)
		testAccountMovements(name, repo, t)
	}
}

func testduplicatedAccount(name string, repo economy.AccountRepo, t *testing.T) {
	ID, _ := uuid.NewV4()
	money := economy.NewMoney(100, "EUR")

	if err := repo.Add(economy.NewAccount(ID, money)); err != nil {
		t.Errorf("Repo %s, err:<%s>", name, err)
	}

	if got, expected := repo.Add(economy.NewAccount(ID, money)), economy.Err_DuplicatedAccount; got != expected {
		t.Errorf("Repo %s, Expecting and error <%s>. Got <%v>", name, expected, got)
	}
}

func testAccountNotFound(name string, repo economy.AccountRepo, t *testing.T) {
	ID, _ := uuid.NewV4()
	_, err := repo.Get(ID)
	if got, expected := err, economy.Err_AccountNotFound; got != expected {
		t.Errorf("Repo %s, Expecting and error <%s>. Got <%v>", name, expected, got)
	}
}

func testAccountMovements(name string, repo economy.AccountRepo, t *testing.T) {

	ID, _ := uuid.NewV4()
	money := economy.NewMoney(0, "EUR")

	if err := repo.Add(economy.NewAccount(ID, money)); err != nil {
		t.Errorf("Repo %s, err:<%s>", name, err)
	}

	Eur_100 := economy.NewMoney(100, "EUR")
	mov_100 := economy.NewAccountMovement(Eur_100, "Win 100", time.Now())

	cash, err := repo.ApplyMovement(ID, mov_100)
	if err!=nil {
		t.Error(err)
	}
	// Checking return value
	if got, expected := string(cash.Currency()), "EUR"; got!=expected {
		t.Error("Repo %s, Wrong currency. Expecting <%s>, got <%s>", name, expected, got)
	}
	if got, expected := cash.Amount(), int64(100); got!=expected {
		t.Error("Repo %s, Wrong amount. Expecting <%v>, got <%v>", name, expected, got)
	}

	// Now, checking Get() method
	acc, err := repo.Get(ID)
	if err!=nil {
		t.Error(err)
	}

	if got, expected := string(acc.Balance().Currency()), "EUR"; got!=expected {
		t.Error("Repo %s, Wrong currency. Expecting <%s>, got <%s>", name, expected, got)
	}
	if got, expected := acc.Balance().Amount(), int64(100); got!=expected {
		t.Error("Repo %s, Wrong amount. Expecting <%v>, got <%v>", name, expected, got)
	}
}




