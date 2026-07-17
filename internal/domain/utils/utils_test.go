package utils

import (
	"testing"
	"time"
)

func TestGenerateRandomPassport(t *testing.T) {
	passport, err := GenerateRandomPassport()
	if err != nil {
		t.Fatal(err)
	}
	
	if len(passport) != passportLength {
		t.Errorf("passport length = %d, want %d", len(passport), passportLength)
	}
	
	passport2, _ := GenerateRandomPassport()
	if passport == passport2 {
		t.Error("two generated passports should not be equal")
	}
}

func TestGenerateRandomUDID(t *testing.T) {
	udid, err := GenerateRandomUDID()
	if err != nil {
		t.Fatal(err)
	}
	
	if len(udid) != udidLength {
		t.Errorf("UDID length = %d, want %d", len(udid), udidLength)
	}
}

func TestHashQueueNumber(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "cfcd208495d565ef66e7dff9f98764da"},
		{1, "c4ca4238a0b923820dcc509a6f75849b"},
		{42, "a1d0c6e83f027327d8461063f4ac58a6"},
	}
	
	for _, tt := range tests {
		result := HashQueueNumber(tt.input)
		if result != tt.expected {
			t.Errorf("HashQueueNumber(%d) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}

func TestCalculateGoldMiningPerHour(t *testing.T) {
	tests := []struct {
		name    string
		cards   []Card
		want    int
		wantErr bool
	}{
		{
			name: "single card",
			cards: []Card{
				{ID: 1, Power: 100},
			},
			want:    75, // 3 * 100^0.7 ≈ 75
			wantErr: false,
		},
		{
			name: "four cards",
			cards: []Card{
				{ID: 1, Power: 50},
				{ID: 2, Power: 50},
				{ID: 3, Power: 50},
				{ID: 4, Power: 50},
			},
			want:    81, // 3 * 200^0.7 ≈ 81
			wantErr: false,
		},
		{
			name:    "empty cards",
			cards:   []Card{},
			wantErr: true,
		},
		{
			name: "too many cards",
			cards: []Card{
				{}, {}, {}, {}, {},
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CalculateGoldMiningPerHour(tt.cards)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateGoldMiningPerHour() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.want {
				t.Errorf("CalculateGoldMiningPerHour() = %d, want %d", result, tt.want)
			}
		})
	}
}

func TestGetMineOverflowDuration(t *testing.T) {
	duration := GetMineOverflowDuration(1000, 5000)
	expected := time.Duration(18000) * time.Second // 5 hours
	
	if duration != expected {
		t.Errorf("GetMineOverflowDuration() = %v, want %v", duration, expected)
	}
}

func TestSortCardsByPower(t *testing.T) {
	cards := []Card{
		{ID: 1, Power: 50},
		{ID: 2, Power: 100},
		{ID: 3, Power: 25},
		{ID: 4, Power: 75},
	}
	
	// Test descending sort
	result := SortCardsByPower(cards, true, 0, false)
	ids := result.([]int)
	
	if ids[0] != 2 || ids[1] != 4 || ids[2] != 1 || ids[3] != 3 {
		t.Errorf("descending sort = %v, want [2 4 1 3]", ids)
	}
	
	// Test ascending sort
	result = SortCardsByPower(cards, true, 0, true)
	ids = result.([]int)
	
	if ids[0] != 3 || ids[1] != 1 || ids[2] != 4 || ids[3] != 2 {
		t.Errorf("ascending sort = %v, want [3 1 4 2]", ids)
	}
	
	// Test with limit
	result = SortCardsByPower(cards, true, 2, false)
	ids = result.([]int)
	
	if len(ids) != 2 || ids[0] != 2 || ids[1] != 4 {
		t.Errorf("limited sort = %v, want [2 4]", ids)
	}
}

func TestGetStrongestCards(t *testing.T) {
	cards := []Card{
		{ID: 1, Power: 10},
		{ID: 2, Power: 50},
		{ID: 3, Power: 30},
		{ID: 4, Power: 40},
		{ID: 5, Power: 20},
	}
	
	strongest := GetStrongestCards(cards, 3)
	
	if len(strongest) != 3 {
		t.Errorf("length = %d, want 3", len(strongest))
	}
	
	if strongest[0] != 2 || strongest[1] != 4 || strongest[2] != 3 {
		t.Errorf("strongest = %v, want [2 4 3]", strongest)
	}
}

func TestGetReadyAndUnreadyCards(t *testing.T) {
	now := time.Now().Unix()
	
	cards := []Card{
		{ID: 1, BaseCardID: 100, LastUsedAt: now - 100},  // Should be ready (cooldown: 50)
		{ID: 2, BaseCardID: 100, LastUsedAt: now - 20},   // Should be unready
		{ID: 3, BaseCardID: 200, LastUsedAt: now - 200},  // Should be ready (cooldown: 100)
	}
	
	cardData := map[int]CardData{
		100: {CoolDown: 50},
		200: {CoolDown: 100},
	}
	
	result := GetReadyAndUnreadyCards(cards, cardData)
	
	if len(result.Ready) != 2 {
		t.Errorf("ready count = %d, want 2", len(result.Ready))
	}
	
	if len(result.Unready) != 1 {
		t.Errorf("unready count = %d, want 1", len(result.Unready))
	}
	
	if result.Unready[0].ID != 2 {
		t.Errorf("unready card ID = %d, want 2", result.Unready[0].ID)
	}
}

func TestFastCalculateGoldMining(t *testing.T) {
	cards := []Card{
		{ID: 1, Power: 100},
	}
	
	fast, _ := FastCalculateGoldMiningPerHour(cards)
	normal, _ := CalculateGoldMiningPerHour(cards)
	
	if fast != normal {
		t.Errorf("fast = %d, normal = %d, should be equal", fast, normal)
	}
}

// Benchmarks
func BenchmarkGenerateRandomPassport(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		GenerateRandomPassport()
	}
}

func BenchmarkHashQueueNumber(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		HashQueueNumber(i)
	}
}

func BenchmarkFastHashQueueNumber(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		FastHashQueueNumber(i)
	}
}

func BenchmarkCalculateGoldMining(b *testing.B) {
	cards := []Card{
		{ID: 1, Power: 50},
		{ID: 2, Power: 50},
		{ID: 3, Power: 50},
		{ID: 4, Power: 50},
	}
	b.ReportAllocs()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		CalculateGoldMiningPerHour(cards)
	}
}

func BenchmarkFastCalculateGoldMining(b *testing.B) {
	cards := []Card{
		{ID: 1, Power: 50},
		{ID: 2, Power: 50},
		{ID: 3, Power: 50},
		{ID: 4, Power: 50},
	}
	b.ReportAllocs()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		FastCalculateGoldMiningPerHour(cards)
	}
}

func BenchmarkSortCardsByPower(b *testing.B) {
	cards := make([]Card, 100)
	for i := range cards {
		cards[i] = Card{ID: i, Power: i * 10}
	}
	b.ReportAllocs()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		SortCardsByPower(cards, true, 10, false)
	}
}

func BenchmarkGetReadyAndUnreadyCards(b *testing.B) {
	now := time.Now().Unix()
	cards := make([]Card, 50)
	for i := range cards {
		cards[i] = Card{
			ID:         i,
			BaseCardID: 100 + i%10,
			LastUsedAt: now - int64(i*10),
		}
	}
	
	cardData := make(map[int]CardData)
	for i := 0; i < 10; i++ {
		cardData[100+i] = CardData{CoolDown: 30}
	}
	
	b.ReportAllocs()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		GetReadyAndUnreadyCards(cards, cardData)
	}
}