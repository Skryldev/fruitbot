// internal/domain/utils/utils.go
package utils

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"sort"
	"sync"
	"time"
	"unsafe"

	"fruitbot/internal/infrastructure/config"
)

const (
	lowercaseChars = "abcdefghijklmnopqrstuvwxyz"
	digitChars     = "0123456789"
	alphanumChars  = lowercaseChars + digitChars
	
	// Lengths
	passportLength = 32
	udidLength     = 16
)

var (
	alphanumBytes = []byte(alphanumChars)
	alphanumLen   = big.NewInt(int64(len(alphanumChars)))
)

func GenerateRandomPassport() (string, error) {
	return generateRandomString(passportLength)
}

func MustGenerateRandomPassport() string {
	passport, err := GenerateRandomPassport()
	if err != nil {
		panic(fmt.Sprintf("failed to generate passport: %v", err))
	}
	return passport
}

func GenerateRandomUDID() (string, error) {
	return generateRandomString(udidLength)
}

func MustGenerateRandomUDID() string {
	udid, err := GenerateRandomUDID()
	if err != nil {
		panic(fmt.Sprintf("failed to generate UDID: %v", err))
	}
	return udid
}

func generateRandomString(length int) (string, error) {
	result := make([]byte, length)
	
	for i := 0; i < length; i++ {
		idx, err := rand.Int(rand.Reader, alphanumLen)
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		result[i] = alphanumBytes[idx.Int64()]
	}
	
	return *(*string)(unsafe.Pointer(&result)), nil
}

// ============================================================
// Device Model Selection
// ============================================================

func GetRandomMobileModel(df *config.DeviceFingerprinter) string {
	if df == nil {
		return "Unknown Device"
	}
	return df.GetRandomModel()
}

// ============================================================
// Hashing Utilities
// ============================================================

func HashQueueNumber(q int) string {
	data := []byte(fmt.Sprintf("%d", q))
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

var hashPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 0, 32)
		return &buf
	},
}

func FastHashQueueNumber(q int) string {
	bufPtr := hashPool.Get().(*[]byte)
	defer hashPool.Put(bufPtr)
	
	buf := (*bufPtr)[:0]
	buf = fmt.Appendf(buf, "%d", q)
	
	hash := md5.Sum(buf)
	return hex.EncodeToString(hash[:])
}

// ============================================================
// Card Power Calculations
// ============================================================

type Card struct {
	ID         int    `json:"id"`
	BaseCardID int    `json:"base_card_id"`
	Power      int    `json:"power"`
	LastUsedAt int64  `json:"last_used_at"`
}

type CardData struct {
	CoolDown int `json:"cooldown"`
}

func CalculateGoldMiningPerHour(cards []Card) (int, error) {
	if len(cards) < 1 || len(cards) > 4 {
		return 0, fmt.Errorf("expected 1-4 cards, got %d", len(cards))
	}
	
	var powerSum int
	for i := range cards {
		powerSum += cards[i].Power
	}
	
	result := 3.0 * math.Pow(float64(powerSum), 0.7)
	return int(result), nil
}

func MustCalculateGoldMiningPerHour(cards []Card) int {
	rate, err := CalculateGoldMiningPerHour(cards)
	if err != nil {
		panic(fmt.Sprintf("failed to calculate gold mining: %v", err))
	}
	return rate
}

func GetMineOverflowDuration(goldPerHour, storageLimit int) time.Duration {
	if goldPerHour <= 0 || storageLimit <= 0 {
		return time.Duration(math.MaxInt64)
	}
	
	seconds := 3600.0 / (float64(goldPerHour) / float64(storageLimit))
	return time.Duration(seconds) * time.Second
}

// ============================================================
// Card Sorting Utilities
// ============================================================

type CardSorter struct {
	cards []Card
	less  func(i, j int) bool
}

func (s CardSorter) Len() int           { return len(s.cards) }
func (s CardSorter) Less(i, j int) bool { return s.less(i, j) }
func (s CardSorter) Swap(i, j int)      { s.cards[i], s.cards[j] = s.cards[j], s.cards[i] }

func SortCardsByPower(cards []Card, returnIDsOnly bool, limit int, ascending bool) interface{} {
	sorted := make([]Card, len(cards))
	copy(sorted, cards)
	
	sorter := &CardSorter{
		cards: sorted,
		less: func(i, j int) bool {
			if ascending {
				return sorted[i].Power < sorted[j].Power
			}
			return sorted[i].Power > sorted[j].Power
		},
	}
	sort.Sort(sorter)
	
	if limit > 0 && limit < len(sorted) {
		sorted = sorted[:limit]
	}
	
	if returnIDsOnly {
		ids := make([]int, len(sorted))
		for i := range sorted {
			ids[i] = sorted[i].ID
		}
		return ids
	}
	
	return sorted
}

func GetStrongestCards(cards []Card, count int) []int {
	if count <= 0 {
		count = 4
	}
	result := SortCardsByPower(cards, true, count, false)
	return result.([]int)
}

func GetWeakestCards(cards []Card, count int) []int {
	if count <= 0 {
		count = 1
	}
	result := SortCardsByPower(cards, true, count, true)
	return result.([]int)
}

// ============================================================
// Card Readiness Calculation
// ============================================================

type CardReadinessResult struct {
	Ready   []Card
	Unready []Card
}

func GetReadyAndUnreadyCards(cards []Card, cardDataMap map[int]CardData) CardReadinessResult {
	now := time.Now().Unix()
	
	result := CardReadinessResult{
		Ready:   make([]Card, 0, len(cards)),
		Unready: make([]Card, 0, len(cards)),
	}
	
	for i := range cards {
		card := cards[i]
		data, exists := cardDataMap[card.BaseCardID]
		if !exists {
			result.Unready = append(result.Unready, card)
			continue
		}
		
		elapsed := now - card.LastUsedAt
		if elapsed < int64(data.CoolDown) {
			result.Unready = append(result.Unready, card)
		} else {
			result.Ready = append(result.Ready, card)
		}
	}
	
	return result
}

// ============================================================
// Generic Utilities
// ============================================================

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Clamp(value, minVal, maxVal int) int {
	return Min(Max(value, minVal), maxVal)
}

// ============================================================
// Pre-computed Constants (for hot paths)
// ============================================================

var goldMiningTable []int

func init() {
	goldMiningTable = make([]int, 4001)
	for i := 1; i <= 4000; i++ {
		goldMiningTable[i] = int(3.0 * math.Pow(float64(i), 0.7))
	}
}

func FastCalculateGoldMiningPerHour(cards []Card) (int, error) {
	if len(cards) < 1 || len(cards) > 4 {
		return 0, fmt.Errorf("expected 1-4 cards, got %d", len(cards))
	}
	
	var powerSum int
	for i := range cards {
		powerSum += cards[i].Power
	}
	
	if powerSum >= len(goldMiningTable) {
		return CalculateGoldMiningPerHour(cards)
	}
	
	return goldMiningTable[powerSum], nil
}