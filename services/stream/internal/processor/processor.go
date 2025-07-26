package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/loyalty/stream/internal/clients"
	"github.com/loyalty/stream/internal/models"
	"github.com/segmentio/kafka-go"
)

type EventProcessor struct {
	ledgerClient     *clients.LedgerClient
	membershipClient *clients.MembershipClient
}

func NewEventProcessor(ledgerURL, membershipURL string) *EventProcessor {
	return &EventProcessor{
		ledgerClient:     clients.NewLedgerClient(ledgerURL),
		membershipClient: clients.NewMembershipClient(membershipURL),
	}
}

func (p *EventProcessor) ProcessEvent(ctx context.Context, message kafka.Message) (*models.ProcessingResult, error) {
	var event models.BaseEvent
	if err := json.Unmarshal(message.Value, &event); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event: %w", err)
	}

	result := &models.ProcessingResult{
		EventID:     event.EventID,
		ProcessedAt: time.Now(),
		Success:     false,
	}

	switch event.EventType {
	case models.EventTypePOSTransaction:
		return p.processPOSTransaction(ctx, &event)
	case models.EventTypeLoyaltyAction:
		return p.processLoyaltyAction(ctx, &event)
	default:
		result.Error = fmt.Sprintf("unknown event type: %s", event.EventType)
		return result, nil
	}
}

func (p *EventProcessor) processPOSTransaction(ctx context.Context, event *models.BaseEvent) (*models.ProcessingResult, error) {
	result := &models.ProcessingResult{
		EventID:     event.EventID,
		ProcessedAt: time.Now(),
		Success:     false,
	}

	var transaction models.POSTransaction
	transactionData, err := json.Marshal(event.Payload)
	if err != nil {
		result.Error = "failed to marshal transaction payload"
		return result, nil
	}

	if err := json.Unmarshal(transactionData, &transaction); err != nil {
		result.Error = "failed to unmarshal transaction data"
		return result, nil
	}

	_, err = p.membershipClient.GetCustomer(event.CustomerID)
	if err != nil {
		result.Error = fmt.Sprintf("failed to get customer: %v", err)
		return result, nil
	}

	org, err := p.membershipClient.GetOrganization(event.OrgID)
	if err != nil {
		result.Error = fmt.Sprintf("failed to get organization: %v", err)
		return result, nil
	}

	pointsEarned := p.calculatePoints(transaction.Amount, org.Settings.PointsPerDollar)
	stampsEarned := org.Settings.StampsPerVisit

	if pointsEarned > 0 {
		_, err := p.ledgerClient.CreatePointsTransfer(
			event.OrgID,
			event.CustomerID,
			pointsEarned,
			fmt.Sprintf("pos_transaction_%s", transaction.TransactionID),
		)
		if err != nil {
			result.Error = fmt.Sprintf("failed to create points transfer: %v", err)
			return result, nil
		}
		result.PointsEarned = pointsEarned
		result.Actions = append(result.Actions, fmt.Sprintf("awarded %d points", pointsEarned))
	}

	if stampsEarned > 0 {
		_, err := p.ledgerClient.CreateStampsTransfer(
			event.OrgID,
			event.CustomerID,
			stampsEarned,
			fmt.Sprintf("pos_transaction_%s", transaction.TransactionID),
		)
		if err != nil {
			result.Error = fmt.Sprintf("failed to create stamps transfer: %v", err)
			return result, nil
		}
		result.StampsEarned = stampsEarned
		result.Actions = append(result.Actions, fmt.Sprintf("awarded %d stamps", stampsEarned))
	}

	rewards := p.checkRewardThresholds(org.Settings.RewardThresholds, pointsEarned, stampsEarned)
	result.RewardsTriggered = rewards

	result.Success = true
	log.Printf("Processed POS transaction %s: %d points, %d stamps, %d rewards",
		transaction.TransactionID, pointsEarned, stampsEarned, len(rewards))

	return result, nil
}

func (p *EventProcessor) processLoyaltyAction(ctx context.Context, event *models.BaseEvent) (*models.ProcessingResult, error) {
	result := &models.ProcessingResult{
		EventID:     event.EventID,
		ProcessedAt: time.Now(),
		Success:     false,
	}

	var action models.LoyaltyAction
	actionData, err := json.Marshal(event.Payload)
	if err != nil {
		result.Error = "failed to marshal loyalty action payload"
		return result, nil
	}

	if err := json.Unmarshal(actionData, &action); err != nil {
		result.Error = "failed to unmarshal loyalty action data"
		return result, nil
	}

	switch action.ActionType {
	case "manual_points":
		if action.Points > 0 {
			_, err := p.ledgerClient.CreatePointsTransfer(
				event.OrgID,
				event.CustomerID,
				action.Points,
				action.Reference,
			)
			if err != nil {
				result.Error = fmt.Sprintf("failed to create points transfer: %v", err)
				return result, nil
			}
			result.PointsEarned = action.Points
			result.Actions = append(result.Actions, fmt.Sprintf("manual award: %d points", action.Points))
		}
	case "bonus_stamps":
		if action.Stamps > 0 {
			_, err := p.ledgerClient.CreateStampsTransfer(
				event.OrgID,
				event.CustomerID,
				action.Stamps,
				action.Reference,
			)
			if err != nil {
				result.Error = fmt.Sprintf("failed to create stamps transfer: %v", err)
				return result, nil
			}
			result.StampsEarned = action.Stamps
			result.Actions = append(result.Actions, fmt.Sprintf("bonus stamps: %d", action.Stamps))
		}
	default:
		result.Error = fmt.Sprintf("unknown loyalty action type: %s", action.ActionType)
		return result, nil
	}

	result.Success = true
	log.Printf("Processed loyalty action %s for customer %s", action.ActionType, event.CustomerID)

	return result, nil
}

func (p *EventProcessor) calculatePoints(amount, pointsPerDollar float64) int {
	if pointsPerDollar <= 0 {
		return 0
	}
	return int(math.Floor(amount * pointsPerDollar))
}

func (p *EventProcessor) checkRewardThresholds(thresholds []clients.RewardThreshold, points, stamps int) []models.RewardTriggered {
	var rewards []models.RewardTriggered

	for _, threshold := range thresholds {
		triggered := false
		
		if threshold.Points > 0 && points >= threshold.Points {
			triggered = true
		}
		
		if threshold.Stamps > 0 && stamps >= threshold.Stamps {
			triggered = true
		}

		if triggered {
			rewards = append(rewards, models.RewardTriggered{
				RewardID:    fmt.Sprintf("reward_%d_%d", threshold.Points, threshold.Stamps),
				RewardType:  threshold.RewardType,
				RewardValue: threshold.RewardValue,
				Description: threshold.Description,
				TriggeredAt: time.Now(),
			})
		}
	}

	return rewards
}