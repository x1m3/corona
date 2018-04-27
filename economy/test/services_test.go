package test

import (
	"testing"
	"github.com/nu7hatch/gouuid"
	"github.com/x1m3/elixir/economy/repos/memory"
	"github.com/x1m3/elixir/economy"
)

func TestApplyMovementBetweenAccountsService(t *testing.T) {

	repo := memoryRepo.NewAccountRepo()

	bankID, _ := uuid.NewV4()
	userID, _ := uuid.NewV4()

	bankAccount := economy.NewAccount(bankID, economy.NewMoney(100, "EUR"))
	playerAccount := economy.NewAccount(userID, economy.NewMoney(100, "EUR"))

	repo.Add(bankAccount)
	repo.Add(playerAccount)

	service := economy.NewApplyMovementBetweenAccountsService(bankAccount, playerAccount, repo)
	err := service.Run(economy.NewMoney(40, "EUR"))
	if err!= nil {
		t.Error(err)
	}

	if got, expected := bankAccount.Balance().Amount(), int64(60); got!=expected {
		t.Errorf("Error in bank account. Expecting an amount of <%v>, got <%v>", expected, got)
	}

	if got, expected := playerAccount.Balance().Amount(), int64(140); got!=expected {
		t.Errorf("Error in user account. Expecting an amount of <%v>, got <%v>", expected, got)
	}
}
