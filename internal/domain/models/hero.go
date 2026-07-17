package models

import (
	"errors"
	"fmt"
)

type HeroItemSlot uint8

const (
	HeroItemSlotLeft  HeroItemSlot = iota
	HeroItemSlotRight
)

type HeroWithItems struct {
	baseHeroID     int      `json:"base_hero_id"`
	leftItemIDs    []int    `json:"left_item_ids"`
	rightItemIDs   []int    `json:"right_item_ids"`
}

func NewHeroWithItems(baseHeroID int, leftItemIDs, rightItemIDs []int) (*HeroWithItems, error) {
	if baseHeroID <= 0 {
		return nil, fmt.Errorf("invalid base hero ID: %d", baseHeroID)
	}

	if leftItemIDs == nil {
		leftItemIDs = []int{}
	}
	if rightItemIDs == nil {
		rightItemIDs = []int{}
	}

	for _, id := range leftItemIDs {
		if id <= 0 {
			return nil, fmt.Errorf("invalid left item ID: %d", id)
		}
	}
	for _, id := range rightItemIDs {
		if id <= 0 {
			return nil, fmt.Errorf("invalid right item ID: %d", id)
		}
	}

	leftCopy := make([]int, len(leftItemIDs))
	copy(leftCopy, leftItemIDs)

	rightCopy := make([]int, len(rightItemIDs))
	copy(rightCopy, rightItemIDs)

	return &HeroWithItems{
		baseHeroID:  baseHeroID,
		leftItemIDs: leftCopy,
		rightItemIDs: rightCopy,
	}, nil
}

func MustNewHeroWithItems(baseHeroID int, leftItemIDs, rightItemIDs []int) *HeroWithItems {
	h, err := NewHeroWithItems(baseHeroID, leftItemIDs, rightItemIDs)
	if err != nil {
		panic(fmt.Sprintf("failed to create HeroWithItems: %v", err))
	}
	return h
}

func (h *HeroWithItems) BaseHeroID() int {
	return h.baseHeroID
}

func (h *HeroWithItems) LeftItemIDs() []int {
	result := make([]int, len(h.leftItemIDs))
	copy(result, h.leftItemIDs)
	return result
}

func (h *HeroWithItems) RightItemIDs() []int {
	result := make([]int, len(h.rightItemIDs))
	copy(result, h.rightItemIDs)
	return result
}

func (h *HeroWithItems) TotalItems() int {
	return len(h.leftItemIDs) + len(h.rightItemIDs)
}

func (h *HeroWithItems) HasItems() bool {
	return len(h.leftItemIDs) > 0 || len(h.rightItemIDs) > 0
}

func (h *HeroWithItems) HasLeftItems() bool {
	return len(h.leftItemIDs) > 0
}

func (h *HeroWithItems) HasRightItems() bool {
	return len(h.rightItemIDs) > 0
}

func (h *HeroWithItems) GetItemsBySlot(slot HeroItemSlot) []int {
	switch slot {
	case HeroItemSlotLeft:
		return h.LeftItemIDs()
	case HeroItemSlotRight:
		return h.RightItemIDs()
	default:
		return nil
	}
}

func (h *HeroWithItems) WithLeftItems(itemIDs []int) (*HeroWithItems, error) {
	return NewHeroWithItems(h.baseHeroID, itemIDs, h.rightItemIDs)
}

func (h *HeroWithItems) WithRightItems(itemIDs []int) (*HeroWithItems, error) {
	return NewHeroWithItems(h.baseHeroID, h.leftItemIDs, itemIDs)
}

func (h *HeroWithItems) AddLeftItem(itemID int) (*HeroWithItems, error) {
	if itemID <= 0 {
		return nil, fmt.Errorf("invalid item ID: %d", itemID)
	}
	
	newLeft := make([]int, len(h.leftItemIDs)+1)
	copy(newLeft, h.leftItemIDs)
	newLeft[len(h.leftItemIDs)] = itemID
	
	return NewHeroWithItems(h.baseHeroID, newLeft, h.rightItemIDs)
}

func (h *HeroWithItems) AddRightItem(itemID int) (*HeroWithItems, error) {
	if itemID <= 0 {
		return nil, fmt.Errorf("invalid item ID: %d", itemID)
	}
	
	newRight := make([]int, len(h.rightItemIDs)+1)
	copy(newRight, h.rightItemIDs)
	newRight[len(h.rightItemIDs)] = itemID
	
	return NewHeroWithItems(h.baseHeroID, h.leftItemIDs, newRight)
}

func (h *HeroWithItems) RemoveLeftItem(itemID int) (*HeroWithItems, error) {
	idx := -1
	for i, id := range h.leftItemIDs {
		if id == itemID {
			idx = i
			break
		}
	}
	if idx == -1 {
		return nil, fmt.Errorf("item %d not found in left slot", itemID)
	}
	
	newLeft := make([]int, 0, len(h.leftItemIDs)-1)
	newLeft = append(newLeft, h.leftItemIDs[:idx]...)
	newLeft = append(newLeft, h.leftItemIDs[idx+1:]...)
	
	return NewHeroWithItems(h.baseHeroID, newLeft, h.rightItemIDs)
}

func (h *HeroWithItems) RemoveRightItem(itemID int) (*HeroWithItems, error) {
	idx := -1
	for i, id := range h.rightItemIDs {
		if id == itemID {
			idx = i
			break
		}
	}
	if idx == -1 {
		return nil, fmt.Errorf("item %d not found in right slot", itemID)
	}
	
	newRight := make([]int, 0, len(h.rightItemIDs)-1)
	newRight = append(newRight, h.rightItemIDs[:idx]...)
	newRight = append(newRight, h.rightItemIDs[idx+1:]...)
	
	return NewHeroWithItems(h.baseHeroID, h.leftItemIDs, newRight)
}

func (h *HeroWithItems) Equal(other *HeroWithItems) bool {
	if other == nil {
		return false
	}
	if h.baseHeroID != other.baseHeroID {
		return false
	}
	if !intSliceEqual(h.leftItemIDs, other.leftItemIDs) {
		return false
	}
	if !intSliceEqual(h.rightItemIDs, other.rightItemIDs) {
		return false
	}
	return true
}

func (h *HeroWithItems) Clone() *HeroWithItems {
	return &HeroWithItems{
		baseHeroID:  h.baseHeroID,
		leftItemIDs: h.LeftItemIDs(),
		rightItemIDs: h.RightItemIDs(),
	}
}

func (h *HeroWithItems) String() string {
	return fmt.Sprintf("HeroWithItems{BaseHeroID: %d, LeftItems: %v, RightItems: %v, Total: %d}",
		h.baseHeroID, h.leftItemIDs, h.rightItemIDs, h.TotalItems())
}

func (h *HeroWithItems) Validate() error {
	if h.baseHeroID <= 0 {
		return errors.New("base hero ID must be positive")
	}
	
	allItems := make(map[int]bool, h.TotalItems())
	for _, id := range h.leftItemIDs {
		if allItems[id] {
			return fmt.Errorf("duplicate item ID %d in left slot", id)
		}
		allItems[id] = true
	}
	for _, id := range h.rightItemIDs {
		if allItems[id] {
			return fmt.Errorf("duplicate item ID %d across slots", id)
		}
		allItems[id] = true
	}
	
	return nil
}

func intSliceEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// ============================================================
// HeroBuilder - Builder pattern for complex hero construction
// ============================================================

type HeroBuilder struct {
	baseHeroID  int
	leftItems   []int
	rightItems  []int
	err         error
}

func NewHeroBuilder(baseHeroID int) *HeroBuilder {
	return &HeroBuilder{
		baseHeroID: baseHeroID,
		leftItems:  make([]int, 0),
		rightItems: make([]int, 0),
	}
}

func (b *HeroBuilder) WithLeftItems(ids ...int) *HeroBuilder {
	if b.err != nil {
		return b
	}
	b.leftItems = append(b.leftItems, ids...)
	return b
}

func (b *HeroBuilder) WithRightItems(ids ...int) *HeroBuilder {
	if b.err != nil {
		return b
	}
	b.rightItems = append(b.rightItems, ids...)
	return b
}

func (b *HeroBuilder) Build() (*HeroWithItems, error) {
	if b.err != nil {
		return nil, b.err
	}
	return NewHeroWithItems(b.baseHeroID, b.leftItems, b.rightItems)
}

func (b *HeroBuilder) MustBuild() *HeroWithItems {
	h, err := b.Build()
	if err != nil {
		panic(err)
	}
	return h
}