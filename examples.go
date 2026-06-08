package main

import (
	"context"
	"fmt"

	"marketpay/internal/store"
	"marketpay/internal/ussd"
)

// ExampleScenario1VendorRegistration demonstrates vendor registration flow
func ExampleScenario1VendorRegistration() {
	fmt.Println("\n=== Scenario 1: Vendor Registration ===\n")

	stateStore := store.NewInMemoryStateStore()
	flowService := ussd.NewMarketPayFlowService(stateStore)
	ctx := context.Background()
	sessionID := "vendor-reg-001"

	// Step 1: Main menu
	result, _ := flowService.Advance(ctx, ussd.AdvanceFlowInput{
		SessionID:   sessionID,
		CurrentPage: ussd.PageSelectService,
		Values:      map[string]string{},
	})
	fmt.Printf("1. Menu:\n%s\n\n", result.Message)

	// Step 2: Select register vendor
	result, _ = flowService.Advance(ctx, ussd.AdvanceFlowInput{
		SessionID:   sessionID,
		CurrentPage: ussd.PageSelectService,
		Values:      map[string]string{"selected_service": "register_vendor"},
	})
	fmt.Printf("2. Redirect to name collection:\n%s\n\n", result.Message)

	// Step 3: Enter vendor name
	result, _ = flowService.Advance(ctx, ussd.AdvanceFlowInput{
		SessionID:   sessionID,
		CurrentPage: ussd.PageCollectVendorName,
		Values:      map[string]string{"registration_vendor_name": "John's Electronics"},
	})
	fmt.Printf("3. Collected vendor name, next:\n%s\n\n", result.Message)

	// Step 4: Enter market name
	result, _ = flowService.Advance(ctx, ussd.AdvanceFlowInput{
		SessionID:   sessionID,
		CurrentPage: ussd.PageCollectMarketName,
		Values:      map[string]string{"registration_market_name": "East Point Market"},
	})
	fmt.Printf("4. Collected market name, next:\n%s\n\n", result.Message)

	// Step 5: Submit registration
	result, _ = flowService.Advance(ctx, ussd.AdvanceFlowInput{
		SessionID:   sessionID,
		CurrentPage: ussd.PageSubmitVendorRegistration,
		Values:      map[string]string{},
	})
	fmt.Printf("5. Registration submitted:\n%s\n\n", result.Message)

	// Step 6: Show result
	result, _ = flowService.Advance(ctx, ussd.AdvanceFlowInput{
		SessionID:   sessionID,
		CurrentPage: ussd.PageShowVendorRegistrationResult,
		Values:      map[string]string{},
	})
	fmt.Printf("6. Result displayed:\n%s\n\n", result.Message)
}

// ExampleScenario2LoanApplication demonstrates loan application flow
func ExampleScenario2LoanApplication() {
	fmt.Println("\n=== Scenario 2: Loan Application ===\n")

	stateStore := store.NewInMemoryStateStore()
	flowService := ussd.NewMarketPayFlowService(stateStore)
	ctx := context.Background()
	sessionID := "loan-app-001"

	// Step 1: Main menu
	result, _ := flowService.Advance(ctx, ussd.AdvanceFlowInput{
		SessionID:   sessionID,
		CurrentPage: ussd.PageSelectService,
		Values:      map[string]string{},
	})
	fmt.Printf("1. Menu shown\n\n")

	// Step 2: Select loan application
	result, _ = flowService.Advance(ctx, ussd.AdvanceFlowInput{
		SessionID:   sessionID,
		CurrentPage: ussd.PageSelectService,
		Values:      map[string]string{"selected_service": "loan_application"},
	})
	fmt.Printf("2. Selected loan application:\n%s\n\n", result.Message)

	// Step 3: Enter loan amount
	result, _ = flowService.Advance(ctx, ussd.AdvanceFlowInput{
		SessionID:   sessionID,
		CurrentPage: ussd.PageCollectLoanAmount,
		Values:      map[string]string{"loan_amount": "1000000"},
	})
	fmt.Printf("3. Loan amount entered:\n%s\n\n", result.Message)

	// Step 4: Confirm application
	result, _ = flowService.Advance(ctx, ussd.AdvanceFlowInput{
		SessionID:   sessionID,
		CurrentPage: ussd.PageConfirmLoanApplication,
		Values:      map[string]string{"loan_confirmed": "submit"},
	})
	fmt.Printf("4. Application confirmed:\n%s\n\n", result.Message)

	// Step 5: Submit application
	result, _ = flowService.Advance(ctx, ussd.AdvanceFlowInput{
		SessionID:   sessionID,
		CurrentPage: ussd.PageSubmitLoanApplication,
		Values:      map[string]string{},
	})
	fmt.Printf("5. Application submitted:\n%s\n\n", result.Message)

	// Step 6: Show result
	result, _ = flowService.Advance(ctx, ussd.AdvanceFlowInput{
		SessionID:   sessionID,
		CurrentPage: ussd.PageShowLoanApplicationResult,
		Values:      map[string]string{},
	})
	fmt.Printf("6. Result displayed:\n%s\n\n", result.Message)
}

// ExampleScenario3PaymentCancellation demonstrates payment cancellation flow
func ExampleScenario3PaymentCancellation() {
	fmt.Println("\n=== Scenario 3: Payment Cancellation ===\n")

	stateStore := store.NewInMemoryStateStore()
	flowService := ussd.NewMarketPayFlowService(stateStore)
	ctx := context.Background()
	sessionID := "payment-cancel-001"

	// Step 1: Main menu
	flowService.Advance(ctx, ussd.AdvanceFlowInput{
		SessionID:   sessionID,
		CurrentPage: ussd.PageSelectService,
		Values:      map[string]string{},
	})

	// Step 2: Select pay vendor
	flowService.Advance(ctx, ussd.AdvanceFlowInput{
		SessionID:   sessionID,
		CurrentPage: ussd.PageSelectService,
		Values:      map[string]string{"selected_service": "pay_vendor"},
	})

	// Step 3: Enter vendor code
	flowService.Advance(ctx, ussd.AdvanceFlowInput{
		SessionID:   sessionID,
		CurrentPage: ussd.PageCollectPaymentVendorCode,
		Values:      map[string]string{"payment_vendor_code": "MP654321"},
	})

	// Step 4: Enter amount
	flowService.Advance(ctx, ussd.AdvanceFlowInput{
		SessionID:   sessionID,
		CurrentPage: ussd.PageCollectPaymentAmount,
		Values:      map[string]string{"payment_amount": "50000"},
	})

	// Step 5: Cancel payment
	result, _ := flowService.Advance(ctx, ussd.AdvanceFlowInput{
		SessionID:   sessionID,
		CurrentPage: ussd.PageConfirmPaymentChoice,
		Values:      map[string]string{"payment_confirmed": "cancel"},
	})
	fmt.Printf("1. Payment cancelled:\n%s\n\n", result.Message)

	// Step 6: Show cancellation message
	result, _ = flowService.Advance(ctx, ussd.AdvanceFlowInput{
		SessionID:   sessionID,
		CurrentPage: ussd.PageShowPaymentCancelled,
		Values:      map[string]string{},
	})
	fmt.Printf("2. Cancellation confirmed:\n%s\n", result.Message)
}

// ExampleScenario4BalanceCheck demonstrates balance check flow
func ExampleScenario4BalanceCheck() {
	fmt.Println("\n=== Scenario 4: Balance Check ===\n")

	stateStore := store.NewInMemoryStateStore()
	flowService := ussd.NewMarketPayFlowService(stateStore)
	ctx := context.Background()
	sessionID := "balance-check-001"

	// Step 1: Main menu
	flowService.Advance(ctx, ussd.AdvanceFlowInput{
		SessionID:   sessionID,
		CurrentPage: ussd.PageSelectService,
		Values:      map[string]string{},
	})

	// Step 2: Select balance check
	flowService.Advance(ctx, ussd.AdvanceFlowInput{
		SessionID:   sessionID,
		CurrentPage: ussd.PageSelectService,
		Values:      map[string]string{"selected_service": "balance_check"},
	})

	// Step 3: Fetch balance
	result, _ := flowService.Advance(ctx, ussd.AdvanceFlowInput{
		SessionID:   sessionID,
		CurrentPage: ussd.PageFetchBalance,
		Values:      map[string]string{},
	})
	fmt.Printf("1. Balance fetching:\n%s\n\n", result.Message)

	// Step 4: Show balance
	result, _ = flowService.Advance(ctx, ussd.AdvanceFlowInput{
		SessionID:   sessionID,
		CurrentPage: ussd.PageShowBalance,
		Values:      map[string]string{},
	})
	fmt.Printf("2. Balance displayed:\n%s\n", result.Message)
}

// ExampleScenario5InvalidInput demonstrates validation
func ExampleScenario5InvalidInput() {
	fmt.Println("\n=== Scenario 5: Invalid Input Handling ===\n")

	stateStore := store.NewInMemoryStateStore()
	flowService := ussd.NewMarketPayFlowService(stateStore)
	ctx := context.Background()
	sessionID := "invalid-input-001"

	// Step 1: Main menu and select vendor registration
	flowService.Advance(ctx, ussd.AdvanceFlowInput{
		SessionID:   sessionID,
		CurrentPage: ussd.PageSelectService,
		Values:      map[string]string{},
	})
	flowService.Advance(ctx, ussd.AdvanceFlowInput{
		SessionID:   sessionID,
		CurrentPage: ussd.PageSelectService,
		Values:      map[string]string{"selected_service": "register_vendor"},
	})

	// Step 2: Enter invalid vendor name (too short)
	result, _ := flowService.Advance(ctx, ussd.AdvanceFlowInput{
		SessionID:   sessionID,
		CurrentPage: ussd.PageCollectVendorName,
		Values:      map[string]string{"registration_vendor_name": "A"},
	})
	fmt.Printf("1. Invalid vendor name (too short):\n%s\n\n", result.Message)

	// Step 3: Enter invalid vendor code
	flowService.Advance(ctx, ussd.AdvanceFlowInput{
		SessionID:   sessionID,
		CurrentPage: ussd.PageSelectService,
		Values:      map[string]string{"selected_service": "pay_vendor"},
	})
	result, _ = flowService.Advance(ctx, ussd.AdvanceFlowInput{
		SessionID:   sessionID,
		CurrentPage: ussd.PageCollectPaymentVendorCode,
		Values:      map[string]string{"payment_vendor_code": "INVALID"},
	})
	fmt.Printf("2. Invalid vendor code:\n%s\n\n", result.Message)

	// Step 4: Enter invalid payment amount
	flowService.Advance(ctx, ussd.AdvanceFlowInput{
		SessionID:   sessionID,
		CurrentPage: ussd.PageCollectPaymentVendorCode,
		Values:      map[string]string{"payment_vendor_code": "MP123456"},
	})
	result, _ = flowService.Advance(ctx, ussd.AdvanceFlowInput{
		SessionID:   sessionID,
		CurrentPage: ussd.PageCollectPaymentAmount,
		Values:      map[string]string{"payment_amount": "99999999999"},
	})
	fmt.Printf("3. Invalid payment amount (too large):\n%s\n", result.Message)
}
