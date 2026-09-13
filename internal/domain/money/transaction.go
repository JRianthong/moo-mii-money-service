package money

import "time"

type Transaction struct {
	id         TransactionID
	userID     UserID
	txType     TransactionType
	amount     Money
	category   Category
	note       Note
	sourceRef  SourceRef
	occurredAt OccurredAt
	createdAt  time.Time
}

type RecordTransactionInput struct {
	UserID     UserID
	Type       TransactionType
	Amount     Money
	Category   Category
	Note       Note
	SourceRef  SourceRef
	OccurredAt OccurredAt
}

func RecordIncome(input RecordTransactionInput) (Transaction, error) {
	input.Type = TransactionTypeIncome
	return recordTransaction(input)
}

func RecordExpense(input RecordTransactionInput) (Transaction, error) {
	input.Type = TransactionTypeExpense
	return recordTransaction(input)
}

func RehydrateTransaction(id TransactionID, input RecordTransactionInput, createdAt time.Time) (Transaction, error) {
	tx, err := recordTransaction(input)
	if err != nil {
		return Transaction{}, err
	}
	tx.id = id
	tx.createdAt = createdAt
	return tx, nil
}

func recordTransaction(input RecordTransactionInput) (Transaction, error) {
	tx := Transaction{
		userID:     input.UserID,
		txType:     input.Type,
		amount:     input.Amount,
		category:   input.Category,
		note:       input.Note,
		sourceRef:  input.SourceRef,
		occurredAt: input.OccurredAt,
	}
	return tx, tx.validate()
}

func (t Transaction) validate() error {
	if t.userID.String() == "" {
		return ErrInvalidTransaction("user id is required")
	}
	if !t.txType.IsValid() {
		return ErrInvalidTransaction("transaction type must be income or expense")
	}
	if t.amount.AmountCents() <= 0 {
		return ErrInvalidTransaction("amount must be greater than zero")
	}
	if err := t.category.ValidateFor(t.txType); err != nil {
		return err
	}
	if t.sourceRef.System() == "" || t.sourceRef.MessageID() == "" {
		return ErrInvalidTransaction("source reference is required")
	}
	if t.occurredAt.Time().IsZero() {
		return ErrInvalidTransaction("occurred at is required")
	}
	return nil
}

func (t Transaction) ID() TransactionID {
	return t.id
}

func (t Transaction) UserID() UserID {
	return t.userID
}

func (t Transaction) Type() TransactionType {
	return t.txType
}

func (t Transaction) Amount() Money {
	return t.amount
}

func (t Transaction) Category() Category {
	return t.category
}

func (t Transaction) Note() Note {
	return t.note
}

func (t Transaction) SourceRef() SourceRef {
	return t.sourceRef
}

func (t Transaction) OccurredAt() time.Time {
	return t.occurredAt.Time()
}

func (t Transaction) CreatedAt() time.Time {
	return t.createdAt
}
