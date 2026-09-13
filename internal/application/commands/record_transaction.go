package commands

import (
	"context"
	"time"

	"github.com/jatuporn/moo-mii-money-service/internal/application/ports"
	"github.com/jatuporn/moo-mii-money-service/internal/domain/money"
)

type RecordTransactionCommand struct {
	UserID          string
	Type            money.TransactionType
	AmountCents     int64
	Currency        string
	Category        string
	Note            string
	SourceSystem    string
	SourceMessageID string
	OccurredAt      time.Time
}

type RecordTransactionHandler struct {
	writer ports.TransactionWriter
}

func NewRecordTransactionHandler(writer ports.TransactionWriter) RecordTransactionHandler {
	return RecordTransactionHandler{writer: writer}
}

type RecordTransactionResult struct {
	Transaction     money.Transaction
	AlreadyRecorded bool
}

func (h RecordTransactionHandler) Handle(ctx context.Context, cmd RecordTransactionCommand) (RecordTransactionResult, error) {
	userID, err := money.NewUserID(cmd.UserID)
	if err != nil {
		return RecordTransactionResult{}, err
	}
	amount, err := money.NewMoney(cmd.AmountCents, cmd.Currency)
	if err != nil {
		return RecordTransactionResult{}, err
	}
	sourceRef, err := money.NewSourceRef(cmd.SourceSystem, cmd.SourceMessageID)
	if err != nil {
		return RecordTransactionResult{}, err
	}

	category, err := money.Categorize(cmd.Type, cmd.Category, cmd.Note)
	if err != nil {
		return RecordTransactionResult{}, err
	}

	input := money.RecordTransactionInput{
		UserID:     userID,
		Amount:     amount,
		Category:   category,
		Note:       money.NewNote(cmd.Note),
		SourceRef:  sourceRef,
		OccurredAt: money.NewOccurredAt(cmd.OccurredAt),
	}

	var tx money.Transaction
	switch cmd.Type {
	case money.TransactionTypeIncome:
		tx, err = money.RecordIncome(input)
	case money.TransactionTypeExpense:
		tx, err = money.RecordExpense(input)
	default:
		err = money.ErrInvalidTransaction("transaction type must be income or expense")
	}
	if err != nil {
		return RecordTransactionResult{}, err
	}

	saved, err := h.writer.Save(ctx, tx)
	if err != nil {
		return RecordTransactionResult{}, err
	}
	return RecordTransactionResult{
		Transaction:     saved.Transaction,
		AlreadyRecorded: saved.AlreadyRecorded,
	}, nil
}
