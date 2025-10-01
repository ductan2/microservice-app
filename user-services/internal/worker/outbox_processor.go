package worker

import (
	"context"
	"log"
	"time"

	"user-services/internal/api/services"
)

// OutboxProcessor handles periodic processing of outbox events
type OutboxProcessor struct {
	service   services.OutboxService
	interval  time.Duration
	batchSize int
	stopChan  chan struct{}
}

// NewOutboxProcessor creates a new outbox processor
func NewOutboxProcessor(service services.OutboxService, interval time.Duration, batchSize int) *OutboxProcessor {
	return &OutboxProcessor{
		service:   service,
		interval:  interval,
		batchSize: batchSize,
		stopChan:  make(chan struct{}),
	}
}

// Start begins processing outbox events in the background
func (p *OutboxProcessor) Start(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	log.Printf("Outbox processor started (interval=%s, batch_size=%d)", p.interval, p.batchSize)

	// Process immediately on start
	if err := p.process(ctx); err != nil {
		log.Printf("Initial outbox processing error: %v", err)
	}

	for {
		select {
		case <-ticker.C:
			if err := p.process(ctx); err != nil {
				log.Printf("Outbox processing error: %v", err)
			}
		case <-p.stopChan:
			log.Println("Outbox processor stopped")
			return
		case <-ctx.Done():
			log.Println("Outbox processor context cancelled")
			return
		}
	}
}

// Stop gracefully stops the processor
func (p *OutboxProcessor) Stop() {
	close(p.stopChan)
}

func (p *OutboxProcessor) process(ctx context.Context) error {
	// Process unpublished events
	if err := p.service.ProcessUnpublishedEvents(ctx, p.batchSize); err != nil {
		return err
	}

	// Periodically cleanup old published events (every hour)
	// We'll do a simple check: cleanup on every 60th processing cycle
	// Or you could add a separate ticker for this

	return nil
}
