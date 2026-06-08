package grpc

import (
	"context"

	"github.com/google/uuid"
	loanapp "github.com/marketpay/backend/internal/loan/application"
	loanmodel "github.com/marketpay/backend/internal/loan/domain/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// LoanServer implements the gRPC LoanService.
// To use: register with grpc.RegisterLoanServiceServer after protoc generation.
type LoanServer struct {
	loanSvc *loanapp.LoanService
}

// NewLoanServer constructs a LoanServer.
func NewLoanServer(loanSvc *loanapp.LoanService) *LoanServer {
	return &LoanServer{loanSvc: loanSvc}
}

// GetLoan retrieves a loan by ID (hand-written until protoc runs).
func (s *LoanServer) GetLoan(ctx context.Context, loanID string) (*LoanProto, error) {
	id, err := uuid.Parse(loanID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid loan ID: %v", err)
	}

	loan, err := s.loanSvc.GetByID(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "loan not found: %v", err)
	}

	return loanToProto(loan), nil
}

// ApproveLoan approves a loan via gRPC.
func (s *LoanServer) ApproveLoan(ctx context.Context, loanID, officerID, note string) (*LoanProto, error) {
	lid, err := uuid.Parse(loanID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid loan ID")
	}
	oid, err := uuid.Parse(officerID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid officer ID")
	}

	loan, err := s.loanSvc.Approve(ctx, lid, oid, note)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "%v", err)
	}
	return loanToProto(loan), nil
}

// LoanProto is a hand-written proto-compatible struct (replace with generated code after protoc).
type LoanProto struct {
	LoanID            string
	VendorID          string
	LoanType          string
	State             string
	PrincipalAmount   float64
	TotalAmount       float64
	OutstandingAmount float64
	InterestRate      float64
	InterestType      string
	TermWeeks         int32
	Frequency         string
	Currency          string
	DisbursedAt       *timestamppb.Timestamp
	DueDate           *timestamppb.Timestamp
}

func loanToProto(loan *loanmodel.Loan) *LoanProto {
	p := &LoanProto{
		LoanID:            loan.ID.String(),
		VendorID:          loan.VendorID.String(),
		LoanType:          string(loan.LoanType),
		State:             string(loan.State),
		PrincipalAmount:   loan.PrincipalAmount,
		TotalAmount:       loan.TotalAmount,
		OutstandingAmount: loan.OutstandingAmount,
		InterestRate:      loan.InterestRate,
		InterestType:      string(loan.InterestType),
		TermWeeks:         int32(loan.TermWeeks),
		Frequency:         string(loan.Frequency),
		Currency:          loan.Currency,
	}
	if loan.DisbursedAt != nil {
		p.DisbursedAt = timestamppb.New(*loan.DisbursedAt)
	}
	if loan.DueDate != nil {
		p.DueDate = timestamppb.New(*loan.DueDate)
	}
	return p
}
