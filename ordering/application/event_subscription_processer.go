package application

import (
	"fmt"
	"log"
	"runtime/debug"
	"time"

	"github.com/Bass-Peerapon/eventsource-demo/ordering/domain/core"
	"github.com/Bass-Peerapon/eventsource-demo/ordering/helper"
)

type EventSubscriptionProcessor interface {
	ProcessNewEvents(eventHandler core.AsyncEventHandler)
}

// EventSubscriptionProcessor ใช้สำหรับจัดการ event subscription
type eventSubscriptionProcessor struct {
	subscriptionRepository core.EventSubscriptionRepository
	eventRepository        core.EventRepository
}

func NewEventSubscriptionProcessor(subscriptionRepository core.EventSubscriptionRepository, eventRepository core.EventRepository) EventSubscriptionProcessor {
	return &eventSubscriptionProcessor{
		subscriptionRepository: subscriptionRepository,
		eventRepository:        eventRepository,
	}
}

func (p *eventSubscriptionProcessor) ProcessNewEvents(eventHadnler core.AsyncEventHandler) {
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
			log.Println(err)
		}
	}()
	// Polling every 1 second
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			// Poll and process new events

			if err := p.processNewEvents(eventHadnler); err != nil {
				helper.Println(fmt.Sprintf("Error processing new events: %v", err))
			}
		}
	}
}

// ProcessNewEvents ใช้ในการประมวลผลเหตุการณ์ใหม่
func (p *eventSubscriptionProcessor) processNewEvents(eventHandler core.AsyncEventHandler) error {
	// สร้าง subscription หากยังไม่มี
	err := p.subscriptionRepository.CreateSubscription(eventHandler.GetSubscriptionName())
	if err != nil {
		return err
	}

	// อ่าน checkpoint และล็อก subscription
	tx, checkpoint, err := p.subscriptionRepository.ReadCheckpointAndLockSubscription(eventHandler.GetSubscriptionName())
	defer tx.Rollback()
	if err != nil {
		return fmt.Errorf("failed to read checkpoint: %w", err)
	}

	if checkpoint != nil {
		helper.Println(fmt.Sprintf("Acquired lock on subscription %s, checkpoint = %+v", eventHandler.GetSubscriptionName(), checkpoint))

		// อ่านเหตุการณ์ใหม่ที่อยู่หลัง checkpoint
		events, err := p.subscriptionRepository.ReadEventsAfterCheckpoint(tx, eventHandler.GetAggregateType(), checkpoint.LasttransactionID, checkpoint.LastEventID)
		if err != nil {
			return fmt.Errorf("failed to read new events: %w", err)
		}

		helper.Println(fmt.Sprintf("Fetched %d new event(s) for subscription %s", len(events), eventHandler.GetSubscriptionName()))
		if len(events) > 0 {
			for _, event := range events {
				// ประมวลผลแต่ละเหตุการณ์
				err := eventHandler.HandleEvent(event)
				if err != nil {
					return fmt.Errorf("failed to handle event: %w", err)
				}
			}

			// อัปเดต subscription ด้วยเหตุการณ์ล่าสุดที่ประมวลผลแล้ว
			lastEvent := events[len(events)-1]
			_, err = p.subscriptionRepository.UpdateEventSubscription(tx, eventHandler.GetSubscriptionName(), lastEvent.TransactionID, lastEvent.ID)
			if err != nil {
				return fmt.Errorf("failed to update event subscription: %w", err)
			}
		}
	}

	return tx.Commit()
}
