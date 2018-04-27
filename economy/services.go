package economy

type ApplyMovementBetweenAccountsService struct {
	bankAccount *Account
	playerAccount *Account
	repo AccountRepo
}

func NewApplyMovementBetweenAccountsService(bank *Account, player *Account, repo AccountRepo) *ApplyMovementBetweenAccountsService{
	return &ApplyMovementBetweenAccountsService{
		bankAccount:bank,
		playerAccount:player,
		repo:repo,
	}
}

func (s *ApplyMovementBetweenAccountsService) Run(m *Money) error{
	return s.repo.ApplyMovementBetweenAccounts(s.bankAccount.id, s.playerAccount.id, m)
}


