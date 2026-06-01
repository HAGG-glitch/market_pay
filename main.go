package main

import (
	"context"
	"fmt"
	"net/http"

	"marketpay/internal/store"
	"marketpay/internal/transport/http"
	"marketpay/internal/ussd"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Initialize logger
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	// Create state store
	stateStore := store.NewInMemoryStateStore()

	// Create flow service
	flowService := ussd.NewMarketPayFlowService(stateStore)

	// Create HTTP handler
	handler := http.NewMarketPayUSSDHandler(flowService)

	// Example 1: Demonstrate flow programmatically
	demonstrateFlow(flowService)

	// Example 2: Start HTTP server
	startHTTPServer(handler)
}

// demonstrateFlow shows how to use the flow service programmatically
func demonstrateFlow(flowService *ussd.MarketPayFlowService) {
	ctx := context.Background()
	sessionID := "demo-session-001"

	fmt.Println("\n=== MarketPay USSD Flow Demo ===\n")

	// Step 1: Show main menu
	fmt.Println("Step 1: Main Menu")
	result, err := flowService.Advance(ctx, ussd.AdvanceFlowInput{
		SessionID:   sessionID,
		CurrentPage: ussd.PageSelectService,
		Values:      map[string]string{},
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to advance flow")
		return
	}
	fmt.Printf("Message: %s\n", result.Message)
	fmt.Printf("Next Page: %s\n\n", result.NextPage)

	// Step 2: Select Pay Vendor service
	fmt.Println("Step 2: Select Pay Vendor Service")
	result, err = flowService.Advance(ctx, ussd.AdvanceFlowInput{
		SessionID:   sessionID,
		CurrentPage: ussd.PageSelectService,
		Values:      map[string]string{"selected_service": "pay_vendor"},
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to advance flow")
		return
	}
	fmt.Printf("Message: %s\n", result.Message)
	fmt.Printf("Next Page: %s\n\n", result.NextPage)

	// Step 3: Enter vendor code
	fmt.Println("Step 3: Enter Vendor Code")
	result, err = flowService.Advance(ctx, ussd.AdvanceFlowInput{
		SessionID:   sessionID,
		CurrentPage: ussd.PageCollectPaymentVendorCode,
		Values:      map[string]string{"payment_vendor_code": "MP123456"},
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to advance flow")
		return
	}
	fmt.Printf("Message: %s\n", result.Message)
	fmt.Printf("Next Page: %s\n\n", result.NextPage)

	// Step 4: Enter payment amount
	fmt.Println("Step 4: Enter Payment Amount")
	result, err = flowService.Advance(ctx, ussd.AdvanceFlowInput{
		SessionID:   sessionID,
		CurrentPage: ussd.PageCollectPaymentAmount,
		Values:      map[string]string{"payment_amount": "100000"},
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to advance flow")
		return
	}
	fmt.Printf("Message: %s\n", result.Message)
	fmt.Printf("Next Page: %s\n\n", result.NextPage)

	// Step 5: Confirm payment
	fmt.Println("Step 5: Confirm Payment with SMS Receipt")
	result, err = flowService.Advance(ctx, ussd.AdvanceFlowInput{
		SessionID:   sessionID,
		CurrentPage: ussd.PageConfirmPaymentChoice,
		Values:      map[string]string{"payment_confirmed": "pay_sms"},
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to advance flow")
		return
	}
	fmt.Printf("Message: %s\n", result.Message)
	fmt.Printf("Next Page: %s\n\n", result.NextPage)

	// Step 6: Submit payment
	fmt.Println("Step 6: Submit Payment")
	result, err = flowService.Advance(ctx, ussd.AdvanceFlowInput{
		SessionID:   sessionID,
		CurrentPage: ussd.PageSubmitVendorPayment,
		Values:      map[string]string{},
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to advance flow")
		return
	}
	fmt.Printf("Message: %s\n", result.Message)
	fmt.Printf("Next Page: %s\n", result.NextPage)
	fmt.Printf("Action: %s\n\n", result.Action)

	fmt.Println("=== End of Demo ===\n")
}

// startHTTPServer starts the HTTP server for USSD requests
func startHTTPServer(handler *http.MarketPayUSSDHandler) {
	http.HandleFunc("/api/ussd/advance", handler.Advance)
	http.HandleFunc("/health", handler.HealthCheck)

	fmt.Println("\nStarting HTTP Server on :8080...")
	fmt.Println("Health Check: http://localhost:8080/health")
	fmt.Println("USSD Endpoint: POST http://localhost:8080/api/ussd/advance")
	fmt.Println("\nExample Request:")
	fmt.Println(`curl -X POST http://localhost:8080/api/ussd/advance \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "test-session",
    "current_page": "mp_select_service",
    "values": {"selected_service": "pay_vendor"}
  }'`)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal().Err(err).Msg("failed to start HTTP server")
	}
}
