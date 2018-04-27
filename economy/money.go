package economy

import "fmt"

type Money struct {
	amount int64
	currency Currency
}

func NewMoney(amount int64, cur Currency) *Money {
	return &Money{amount:amount,currency:cur}
}

func (b *Money) Add(bal *Money) (*Money, error) {
	if b.currency != bal.currency {
		return nil, Err_CoinDiffers
	}
	return NewMoney(b.amount + bal.amount, b.currency), nil
}

func (b *Money) Amount() int64{
	return b.amount
}

func (b *Money) Currency() Currency {
	return b.currency
}

func (b *Money) Copy() *Money {
	return NewMoney(b.amount, b.currency)
}

func (b *Money) Equal(o *Money) bool {
	return b.amount == o.amount && b.currency == o.currency
}

func (b *Money) String() string {
	return fmt.Sprintf("%d %s", b.amount, b.currency)
}
