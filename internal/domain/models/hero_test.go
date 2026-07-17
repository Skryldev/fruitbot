// internal/domain/models/hero_test.go
package models

import (
	"encoding/json"
	"testing"
)

func TestNewHeroWithItems(t *testing.T) {
	tests := []struct {
		name        string
		baseHeroID  int
		leftItems   []int
		rightItems  []int
		wantErr     bool
	}{
		{
			name:       "valid hero without items",
			baseHeroID: 415,
			leftItems:  []int{},
			rightItems: []int{},
			wantErr:    false,
		},
		{
			name:       "valid hero with items",
			baseHeroID: 415,
			leftItems:  []int{101, 102},
			rightItems: []int{201},
			wantErr:    false,
		},
		{
			name:       "invalid base hero ID",
			baseHeroID: 0,
			leftItems:  []int{},
			rightItems: []int{},
			wantErr:    true,
		},
		{
			name:       "negative base hero ID",
			baseHeroID: -1,
			leftItems:  []int{},
			rightItems: []int{},
			wantErr:    true,
		},
		{
			name:       "invalid left item ID",
			baseHeroID: 415,
			leftItems:  []int{0},
			rightItems: []int{},
			wantErr:    true,
		},
		{
			name:       "invalid right item ID",
			baseHeroID: 415,
			leftItems:  []int{},
			rightItems: []int{-5},
			wantErr:    true,
		},
		{
			name:       "nil slices converted to empty",
			baseHeroID: 415,
			leftItems:  nil,
			rightItems: nil,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hero, err := NewHeroWithItems(tt.baseHeroID, tt.leftItems, tt.rightItems)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewHeroWithItems() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if hero == nil {
					t.Error("expected non-nil hero")
					return
				}
				if hero.BaseHeroID() != tt.baseHeroID {
					t.Errorf("BaseHeroID() = %d, want %d", hero.BaseHeroID(), tt.baseHeroID)
				}
			}
		})
	}
}

func TestHeroWithItems_Immutability(t *testing.T) {
	leftItems := []int{101, 102}
	rightItems := []int{201}
	
	hero, _ := NewHeroWithItems(415, leftItems, rightItems)
	
	// Modify original slices
	leftItems[0] = 999
	rightItems[0] = 888
	
	// Hero should be unaffected
	gotLeft := hero.LeftItemIDs()
	if gotLeft[0] != 101 {
		t.Errorf("LeftItemIDs()[0] = %d, want 101 (immutability violated)", gotLeft[0])
	}
	
	gotRight := hero.RightItemIDs()
	if gotRight[0] != 201 {
		t.Errorf("RightItemIDs()[0] = %d, want 201 (immutability violated)", gotRight[0])
	}
}

func TestHeroWithItems_TotalItems(t *testing.T) {
	hero, _ := NewHeroWithItems(415, []int{1, 2, 3}, []int{4, 5})
	
	if got := hero.TotalItems(); got != 5 {
		t.Errorf("TotalItems() = %d, want 5", got)
	}
}

func TestHeroWithItems_HasMethods(t *testing.T) {
	heroEmpty, _ := NewHeroWithItems(415, []int{}, []int{})
	heroWithItems, _ := NewHeroWithItems(415, []int{1}, []int{2})
	heroLeftOnly, _ := NewHeroWithItems(415, []int{1}, []int{})
	heroRightOnly, _ := NewHeroWithItems(415, []int{}, []int{2})
	
	if heroEmpty.HasItems() {
		t.Error("HasItems() should be false for empty hero")
	}
	if !heroWithItems.HasItems() {
		t.Error("HasItems() should be true for hero with items")
	}
	if !heroLeftOnly.HasLeftItems() {
		t.Error("HasLeftItems() should be true")
	}
	if heroLeftOnly.HasRightItems() {
		t.Error("HasRightItems() should be false")
	}
	if heroRightOnly.HasLeftItems() {
		t.Error("HasLeftItems() should be false")
	}
	if !heroRightOnly.HasRightItems() {
		t.Error("HasRightItems() should be true")
	}
}

func TestHeroWithItems_ImmutableUpdates(t *testing.T) {
	hero, _ := NewHeroWithItems(415, []int{1, 2}, []int{3})
	
	// Add left item
	updated, err := hero.AddLeftItem(4)
	if err != nil {
		t.Fatal(err)
	}
	if updated.TotalItems() != 4 {
		t.Errorf("TotalItems() = %d, want 4", updated.TotalItems())
	}
	if hero.TotalItems() != 3 {
		t.Error("original hero should not be modified")
	}
	
	// Remove right item
	updated, err = updated.RemoveRightItem(3)
	if err != nil {
		t.Fatal(err)
	}
	if updated.TotalItems() != 3 {
		t.Errorf("TotalItems() = %d, want 3", updated.TotalItems())
	}
}

func TestHeroWithItems_Equal(t *testing.T) {
	h1, _ := NewHeroWithItems(415, []int{1, 2}, []int{3})
	h2, _ := NewHeroWithItems(415, []int{1, 2}, []int{3})
	h3, _ := NewHeroWithItems(415, []int{1}, []int{3})
	h4, _ := NewHeroWithItems(515, []int{1, 2}, []int{3})
	
	if !h1.Equal(h2) {
		t.Error("Equal() should be true for identical heroes")
	}
	if h1.Equal(h3) {
		t.Error("Equal() should be false for different items")
	}
	if h1.Equal(h4) {
		t.Error("Equal() should be false for different base ID")
	}
	if h1.Equal(nil) {
		t.Error("Equal() should be false for nil")
	}
}

func TestHeroWithItems_Clone(t *testing.T) {
	hero, _ := NewHeroWithItems(415, []int{1, 2}, []int{3})
	clone := hero.Clone()
	
	if !hero.Equal(clone) {
		t.Error("Clone() should be equal to original")
	}
	
	// Verify deep copy
	clone, _ = clone.AddLeftItem(4)
	if hero.TotalItems() != 3 {
		t.Error("original should not be affected by clone modification")
	}
}

func TestHeroWithItems_Validate(t *testing.T) {
	tests := []struct {
		name    string
		hero    func() *HeroWithItems
		wantErr bool
	}{
		{
			name: "valid hero",
			hero: func() *HeroWithItems {
				h, _ := NewHeroWithItems(415, []int{1, 2}, []int{3})
				return h
			},
			wantErr: false,
		},
		{
			name: "duplicate items in left slot",
			hero: func() *HeroWithItems {
				return &HeroWithItems{
					baseHeroID:  415,
					leftItemIDs: []int{1, 1},
					rightItemIDs: []int{2},
				}
			},
			wantErr: true,
		},
		{
			name: "duplicate items across slots",
			hero: func() *HeroWithItems {
				return &HeroWithItems{
					baseHeroID:  415,
					leftItemIDs: []int{1},
					rightItemIDs: []int{1},
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hero := tt.hero()
			err := hero.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHeroBuilder(t *testing.T) {
	builder := NewHeroBuilder(415).
		WithLeftItems(1, 2, 3).
		WithRightItems(4, 5)
	
	hero, err := builder.Build()
	if err != nil {
		t.Fatal(err)
	}
	
	if hero.BaseHeroID() != 415 {
		t.Errorf("BaseHeroID() = %d, want 415", hero.BaseHeroID())
	}
	if hero.TotalItems() != 5 {
		t.Errorf("TotalItems() = %d, want 5", hero.TotalItems())
	}
}

func TestMustFunctions(t *testing.T) {
	// MustNewHeroWithItems should not panic with valid input
	hero := MustNewHeroWithItems(415, []int{1}, []int{2})
	if hero == nil {
		t.Error("MustNewHeroWithItems returned nil")
	}
	
	// MustBuild should not panic with valid input
	hero = NewHeroBuilder(415).WithLeftItems(1).MustBuild()
	if hero == nil {
		t.Error("MustBuild returned nil")
	}
	
	// Test panic recovery
	defer func() {
		if r := recover(); r != nil {
			t.Log("Expected panic caught:", r)
		}
	}()
	MustNewHeroWithItems(0, nil, nil)
	t.Error("Should have panicked")
}

func TestJSONSerialization(t *testing.T) {
	hero, _ := NewHeroWithItems(415, []int{1, 2}, []int{3})
	
	data, err := json.Marshal(hero)
	if err != nil {
		t.Fatal(err)
	}
	
	var decoded HeroWithItems
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatal(err)
	}
	
	if !hero.Equal(&decoded) {
		t.Error("JSON round-trip failed")
	}
}

// Benchmarks
func BenchmarkNewHeroWithItems(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		NewHeroWithItems(415, []int{1, 2, 3}, []int{4, 5})
	}
}

func BenchmarkHeroWithItems_Clone(b *testing.B) {
	hero, _ := NewHeroWithItems(415, []int{1, 2, 3}, []int{4, 5})
	b.ReportAllocs()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		hero.Clone()
	}
}

func BenchmarkHeroWithItems_AddItem(b *testing.B) {
	hero, _ := NewHeroWithItems(415, []int{1, 2}, []int{3})
	b.ReportAllocs()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		hero.AddLeftItem(i)
	}
}

func BenchmarkHeroBuilder_Build(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		NewHeroBuilder(415).
			WithLeftItems(1, 2, 3).
			WithRightItems(4, 5).
			Build()
	}
}