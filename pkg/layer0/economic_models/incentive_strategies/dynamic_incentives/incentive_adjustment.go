package dynamic_incentives

import (
	"fmt"
	"math/big"
	"sync"
	"time"
)

// IncentiveAdjustmentType defines the type of incentive adjustment
type IncentiveAdjustmentType string

const (
	IncreaseIncentive IncentiveAdjustmentType = "increase"
	DecreaseIncentive IncentiveAdjustmentType = "decrease"
)

// IncentiveAdjustment defines the structure of an incentive adjustment
type IncentiveAdjustment struct {
	Type        IncentiveAdjustmentType
	Amount      *big.Int
	AdjustedAt  time.Time
	Description string
}

// DynamicIncentiveManager manages the dynamic incentives in the network
type DynamicIncentiveManager struct {
	incentives map[string][]*Incentive
	mu         sync.Mutex
}

// NewDynamicIncentiveManager initializes a new DynamicIncentiveManager
func NewDynamicIncentiveManager() *DynamicIncentiveManager {
	return &DynamicIncentiveManager{
		incentives: make(map[string][]*Incentive),
	}
}

// IssueIncentive issues a new dynamic incentive to a user
func (dim *DynamicIncentiveManager) IssueIncentive(recipient string, incentiveType IncentiveType, amount *big.Int, duration time.Duration) (*Incentive, error) {
	if amount.Cmp(big.NewInt(0)) <= 0 {
		return nil, fmt.Errorf("incentive amount must be positive")
	}

	incentive := &Incentive{
		Type:      incentiveType,
		Amount:    amount,
		IssuedAt:  time.Now(),
		Expiry:    time.Now().Add(duration),
		Recipient: recipient,
	}

	dim.mu.Lock()
	defer dim.mu.Unlock()
	dim.incentives[recipient] = append(dim.incentives[recipient], incentive)

	fmt.Printf("Issued incentive of type %s to user %s with amount %s\n", incentiveType, recipient, amount.String())
	return incentive, nil
}

// RedeemIncentive redeems an incentive for a user
func (dim *DynamicIncentiveManager) RedeemIncentive(recipient string, incentiveType IncentiveType) (*Incentive, error) {
	dim.mu.Lock()
	defer dim.mu.Unlock()

	userIncentives, exists := dim.incentives[recipient]
	if !exists || len(userIncentives) == 0 {
		return nil, fmt.Errorf("no incentives found for user %s", recipient)
	}

	var redeemedIncentive *Incentive
	for i, incentive := range userIncentives {
		if incentive.Type == incentiveType && time.Now().Before(incentive.Expiry) {
			redeemedIncentive = incentive
			// Remove redeemed incentive
			dim.incentives[recipient] = append(userIncentives[:i], userIncentives[i+1:]...)
			break
		}
	}

	if redeemedIncentive == nil {
		return nil, fmt.Errorf("no valid incentive of type %s found for user %s", incentiveType, recipient)
	}

	fmt.Printf("Redeemed incentive of type %s for user %s with amount %s\n", incentiveType, recipient, redeemedIncentive.Amount.String())
	return redeemedIncentive, nil
}

// GetIncentives lists all incentives for a user
func (dim *DynamicIncentiveManager) GetIncentives(recipient string) ([]*Incentive, error) {
	dim.mu.Lock()
	defer dim.mu.Unlock()

	userIncentives, exists := dim.incentives[recipient]
	if !exists {
		return nil, fmt.Errorf("no incentives found for user %s", recipient)
	}

	return userIncentives, nil
}

// CleanExpiredIncentives removes expired incentives from the system
func (dim *DynamicIncentiveManager) CleanExpiredIncentives() {
	dim.mu.Lock()
	defer dim.mu.Unlock()

	for recipient, incentives := range dim.incentives {
		var validIncentives []*Incentive
		for _, incentive := range incentives {
			if time.Now().Before(incentive.Expiry) {
				validIncentives = append(validIncentives, incentive)
			} else {
				fmt.Printf("Removed expired incentive of type %s for user %s with amount %s\n", incentive.Type, recipient, incentive.Amount.String())
			}
		}
		dim.incentives[recipient] = validIncentives
	}
}

// AdjustIncentives dynamically adjusts the incentives based on network conditions
func (dim *DynamicIncentiveManager) AdjustIncentives(networkCondition string) {
	dim.mu.Lock()
	defer dim.mu.Unlock()

	for _, incentives := range dim.incentives {
		for _, incentive := range incentives {
			switch networkCondition {
			case "high_congestion":
				incentive.Amount = new(big.Int).Div(incentive.Amount, big.NewInt(2)) // Halve the incentive amount
				fmt.Printf("Adjusted incentive of type %s for user %s due to high congestion. New amount: %s\n", incentive.Type, incentive.Recipient, incentive.Amount.String())
			case "low_activity":
				incentive.Amount = new(big.Int).Mul(incentive.Amount, big.NewInt(2)) // Double the incentive amount
				fmt.Printf("Adjusted incentive of type %s for user %s due to low activity. New amount: %s\n", incentive.Type, incentive.Recipient, incentive.Amount.String())
			}
		}
	}
}

// RecordAdjustment records an incentive adjustment
func (dim *DynamicIncentiveManager) RecordAdjustment(incentive *Incentive, adjustmentType IncentiveAdjustmentType, amount *big.Int, description string) (*IncentiveAdjustment, error) {
	if amount.Cmp(big.NewInt(0)) <= 0 {
		return nil, fmt.Errorf("adjustment amount must be positive")
	}

	adjustment := &IncentiveAdjustment{
		Type:        adjustmentType,
		Amount:      amount,
		AdjustedAt:  time.Now(),
		Description: description,
	}

	// Apply adjustment
	switch adjustmentType {
	case IncreaseIncentive:
		incentive.Amount = new(big.Int).Add(incentive.Amount, amount)
	case DecreaseIncentive:
		if incentive.Amount.Cmp(amount) < 0 {
			return nil, fmt.Errorf("adjustment amount cannot exceed incentive amount")
		}
		incentive.Amount = new(big.Int).Sub(incentive.Amount, amount)
	}

	fmt.Printf("Recorded adjustment: %s incentive for user %s by %s with new amount %s\n", adjustmentType, incentive.Recipient, amount.String(), incentive.Amount.String())
	return adjustment, nil
}

func main() {
	dim := NewDynamicIncentiveManager()
	amount := big.NewInt(1000)

	// Issue an incentive
	incentive, err := dim.IssueIncentive("user1", MiningIncentive, amount, 7*24*time.Hour) // 7 days expiry
	if err != nil {
		fmt.Println("Error issuing incentive:", err)
		return
	}

	fmt.Println("Issued incentive:", incentive)

	// Adjust incentive
	adjustedAmount := big.NewInt(200)
	_, err = dim.RecordAdjustment(incentive, IncreaseIncentive, adjustedAmount, "Reward for high participation")
	if err != nil {
		fmt.Println("Error adjusting incentive:", err)
		return
	}

	fmt.Println("Adjusted incentive:", incentive)

	// Redeem an incentive
	redeemedIncentive, err := dim.RedeemIncentive("user1", MiningIncentive)
	if err != nil {
		fmt.Println("Error redeeming incentive:", err)
		return
	}

	fmt.Println("Redeemed incentive:", redeemedIncentive)

	// List incentives for a user
	incentives, err := dim.GetIncentives("user1")
	if err != nil {
		fmt.Println("Error listing incentives:", err)
		return
	}

	fmt.Println("Incentives for user1:", incentives)

	// Adjust incentives based on network conditions
	dim.AdjustIncentives("high_congestion")
	dim.AdjustIncentives("low_activity")

	// Clean expired incentives
	dim.CleanExpiredIncentives()
}
