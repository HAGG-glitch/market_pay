package postgres

import (
	"context"

	"github.com/google/uuid"
	ledgermodel "github.com/marketpay/backend/internal/ledger/domain/model"
	"gorm.io/gorm"
)

// LedgerRepo implements LedgerRepository.
type LedgerRepo struct {
	db *gorm.DB
}

func NewLedgerRepo(db *gorm.DB) *LedgerRepo {
	return &LedgerRepo{db: db}
}

func (r *LedgerRepo) FindAccount(ctx context.Context, accountType ledgermodel.AccountType) (*ledgermodel.Account, error) {
	var account ledgermodel.Account
	err := r.db.WithContext(ctx).Where("type = ?", accountType).First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *LedgerRepo) CreateJournalEntry(ctx context.Context, entry *ledgermodel.JournalEntry) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(entry).Error; err != nil {
			return err
		}
		// Update account balances
		for _, line := range entry.Lines {
			delta := line.Amount
			if line.EntryType == ledgermodel.EntryTypeCredit {
				delta = -line.Amount
			}
			if err := tx.Model(&ledgermodel.Account{}).
				Where("id = ?", line.AccountID).
				UpdateColumn("balance", gorm.Expr("balance + ?", delta)).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *LedgerRepo) UpdateAccountBalance(ctx context.Context, accountID uuid.UUID, delta float64) error {
	return r.db.WithContext(ctx).
		Model(&ledgermodel.Account{}).
		Where("id = ?", accountID).
		UpdateColumn("balance", gorm.Expr("balance + ?", delta)).Error
}

func (r *LedgerRepo) ListEntries(ctx context.Context, offset, limit int) ([]*ledgermodel.JournalEntry, int64, error) {
	var entries []*ledgermodel.JournalEntry
	var count int64
	r.db.WithContext(ctx).Model(&ledgermodel.JournalEntry{}).Count(&count)
	err := r.db.WithContext(ctx).
		Preload("Lines").
		Preload("Lines.Account").
		Order("entry_date DESC").
		Offset(offset).Limit(limit).
		Find(&entries).Error
	return entries, count, err
}
