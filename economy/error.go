package economy

import "errors"

var Err_CoinDiffers = errors.New("coin demonination differs")
var Err_AccountNotFound = errors.New("account not found")
var Err_DuplicatedAccount = errors.New("duplicated account")
