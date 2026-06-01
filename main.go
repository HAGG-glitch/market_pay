package main

import (
	"context"
	"fmt"

	"marketpay/internal/store"
	"marketpay/internal/transport/http"
	"marketpay/internal/ussd"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Initialize logger
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	// Set Gin to release mode for production
	// gin.SetMode(gin.ReleaseMode)

	// Create state store
	stateStore := store.NewInMemoryStateStore()

	// Create flow service
	flowService := ussd.NewMarketPayFlowService(stateStore)

	// Create HTTP handler
	handler := http.NewMarketPayUSSDHandler(flowService)

	// Example 1: Demonstrate flow programmatically
	demonstrateFlow(flowService)

	// Example 2: Start HTTP server with Gin
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

// startHTTPServer starts the HTTP server for USSD requests using Gin
func startHTTPServer(handler *http.MarketPayUSSDHandler) {
	// Create Gin router
	router := gin.Default()

	// Add middleware for logging
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Setup all routes
	handler.SetupRoutes(router)

	// Print available routes
	fmt.Println("\n=== MarketPay USSD Service Started ===\n")
	fmt.Println("📍 Server running on http://localhost:8080")
	fmt.Println("\n📚 Available Endpoints:")
	fmt.Println("   • GET  /health")
	fmt.Println("   • GET  /healthz")
	fmt.Println("   • GET  /api/info")
	fmt.Println("   • POST /api/ussd/advance (legacy)")
	fmt.Println("   • POST /api/v1/ussd/advance (v1)")
	fmt.Println("\n📝 Example Request:")
	fmt.Println(`   curl -X POST http://localhost:8080/api/v1/ussd/advance \
     -H "Content-Type: application/json" \
     -d '{
       "session_id": "user-001",
       "current_page": "mp_select_service",
       "values": {"selected_service": "pay_vendor"}
     }'`)
	fmt.Println("\n✅ Service is ready to handle requests\n")

	// Start server
	if err := router.Run(":8080"); err != nil {
		log.Fatal().Err(err).Msg("failed to start HTTP server")
	}
}
