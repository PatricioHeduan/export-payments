package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"export-payments/config"
	"export-payments/google"
	"export-payments/mercadopago"
	"export-payments/scheduler"
)

func main() {
	log.Println("[Main] Starting MercadoPago to Google Sheets payment export service...")

	// 1. Load configuration strictly from environment variables
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[Main] Configuration failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 2. Instantiate MercadoPago client
	mpClient := mercadopago.NewClient(cfg.MPBaseURL, cfg.MPAccessToken)

	// 3. Instantiate Google Sheets client
	sheetsClient, err := google.NewSheetsClient(ctx, cfg)
	if err != nil {
		log.Fatalf("[Main] Failed to initialize Google Sheets client: %v", err)
	}

	// 4. Ensure header rows are present in both target sheets (Income and Expense)
	if err := sheetsClient.EnsureHeaderRows(ctx); err != nil {
		log.Printf("[Main] Warning: Could not verify or append header rows: %v", err)
	}

	// 5. Define synchronization task
	syncJob := func(jobCtx context.Context) error {
		log.Println("[Sync] Fetching transactions from MercadoPago...")

		resp, err := mpClient.SearchPayments(jobCtx, mercadopago.SearchParams{
			Sort:      "date_created",
			Criteria:  "asc",
			BeginDate: "NOW-90DAYS",
			EndDate:   "NOW",
			Limit:     500,
		})
		if err != nil {
			return err
		}

		log.Printf("[Sync] Retrieved %d transactions (total: %d)", len(resp.Results), resp.Paging.Total)

		if err := sheetsClient.AppendPayments(jobCtx, resp.Results); err != nil {
			return err
		}

		log.Println("[Sync] Transactions appended to Google Sheets successfully.")
		return nil
	}

	// Run initial synchronization once at application startup
	log.Println("[Main] Triggering initial synchronization on startup...")
	if err := syncJob(ctx); err != nil {
		log.Printf("[Main] Initial synchronization encountered an error: %v", err)
	}

	// 6. Setup and start cron scheduler
	sched := scheduler.NewScheduler(cfg.CronSchedule, syncJob)
	if err := sched.Start(ctx); err != nil {
		log.Fatalf("[Main] Failed to start cron scheduler: %v", err)
	}

	// 7. Handle OS signals for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	sig := <-quit
	log.Printf("[Main] Caught OS signal '%v'. Initiating shutdown...", sig)

	// Stop scheduler and wait for active jobs
	stopCtx := sched.Stop()
	select {
	case <-stopCtx.Done():
		log.Println("[Main] Scheduler jobs completed cleanly.")
	case <-time.After(15 * time.Second):
		log.Println("[Main] Shutdown timeout exceeded (15s). Exiting forceably.")
	}

	cancel()
	log.Println("[Main] Service stopped.")
}
