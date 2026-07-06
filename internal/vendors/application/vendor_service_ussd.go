package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	shared "github.com/marketpay/backend/internal/shared/domain/model"
	vendormodel "github.com/marketpay/backend/internal/vendors/domain/model"
	apperrors "github.com/marketpay/backend/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// USSDRegisterInput holds vendor registration from Monime USSD (no email).
type USSDRegisterInput struct {
	FirstName    string
	LastName     string
	Phone        string
	MarketName   string
	NationalID   string
	BusinessType string
	PIN          string
	UserID       uuid.UUID
	FieldAgentID *uuid.UUID
	IsDemo       bool
}

// RegisterFromUSSD registers a vendor using phone as primary identity.
func (s *VendorService) RegisterFromUSSD(ctx context.Context, input USSDRegisterInput) (*vendormodel.Vendor, error) {
	s.log.Info("ussd_reg step1 - find existing")
	existing, _ := s.vendors.FindByPhone(ctx, input.Phone)
	if existing != nil {
		return nil, apperrors.ErrAlreadyExists("vendor with this phone")
	}

	s.log.Info("ussd_reg step2 - list markets")
	markets, _ := s.vendors.ListMarketAssociations(ctx)
	var marketID uuid.UUID
	for _, m := range markets {
		if strings.EqualFold(m.Name, input.MarketName) {
			marketID = m.ID
			break
		}
	}
	if marketID == uuid.Nil && len(markets) > 0 {
		marketID = markets[0].ID
	}
	s.log.Info("ussd_reg step3 - market", zap.String("market_id", marketID.String()))

	s.log.Info("ussd_reg step4 - bcrypt start")
	pinHash, err := bcrypt.GenerateFromPassword([]byte(input.PIN), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperrors.ErrInternalServer(err)
	}
	s.log.Info("ussd_reg step5 - bcrypt done")

	code := fmt.Sprintf("MP%05d", time.Now().Unix()%100000)
	vendor := &vendormodel.Vendor{
		UserID:              input.UserID,
		FirstName:           input.FirstName,
		LastName:            input.LastName,
		Phone:               input.Phone,
		NationalIDNumber:    input.NationalID,
		NationalIDType:      "NATIONAL_ID",
		DateOfBirth:         time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		MarketAssociationID: marketID,
		BusinessType:        input.BusinessType,
		KYCStatus:           shared.KYCPending,
		Status:              vendormodel.VendorStatusPending,
		PINHash:             string(pinHash),
		VendorCode:          code,
		IsDemo:              input.IsDemo,
		FieldAgentID:        input.FieldAgentID,
	}

	s.log.Info("ussd_reg step6 - vendor create start")
	if err := s.vendors.Create(ctx, vendor); err != nil {
		s.log.Error("ussd_reg step6 - vendor create failed", zap.Error(err))
		return nil, apperrors.ErrInternalServer(err)
	}
	s.log.Info("ussd_reg step7 - vendor created", zap.String("vendor_id", vendor.ID.String()))

	s.log.Info("ussd_reg step8 - publish event")
	_ = s.events.Publish(ctx, "VendorCreated", vendor.ID.String(), map[string]interface{}{
		"vendor_id":    vendor.ID.String(),
		"vendor_code":  code,
		"phone":        vendor.Phone,
		"source":       "ussd",
	})

	s.log.Info("ussd_reg step9 - done")
	return vendor, nil
}

func (s *VendorService) FindByPhone(ctx context.Context, phone string) (*vendormodel.Vendor, error) {
	return s.vendors.FindByPhone(ctx, phone)
}

func (s *VendorService) CheckEligibilityByPhone(ctx context.Context, phone string) (bool, string, error) {
	vendor, err := s.vendors.FindByPhone(ctx, phone)
	if err != nil || vendor == nil {
		return false, "Register as a vendor first.", nil
	}
	if vendor.FrozenAt != nil {
		return false, "Your account is frozen. Contact your field agent.", nil
	}
	if err := vendor.IsEligibleForLoan(); err != nil {
		return false, friendlyEligibilityError(err), nil
	}
	if vendor.CreditScore < 50 {
		return false, "Credit score too low. Keep trading to improve eligibility.", nil
	}
	return true, "You are eligible to apply for a loan.", nil
}

func friendlyEligibilityError(err error) string {
	msg := err.Error()
	if strings.Contains(msg, "transaction history") {
		return "You need to use the system for 6 months before you can get a loan."
	}
	return msg
}
