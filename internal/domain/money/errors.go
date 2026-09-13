package money

type ErrInvalidTransaction string

func (e ErrInvalidTransaction) Error() string {
	return string(e)
}
