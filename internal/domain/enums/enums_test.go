package enums

import (
	"testing"
)

func TestGender_String(t *testing.T) {
	tests := []struct {
		gender   Gender
		expected string
	}{
		{GenderUndefined, "UNDEFINED"},
		{GenderMale, "MALE"},
		{GenderFemale, "FEMALE"},
		{Gender(99), "Gender(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.gender.String(); got != tt.expected {
				t.Errorf("Gender.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGender_IsValid(t *testing.T) {
	if !GenderMale.IsValid() {
		t.Error("GenderMale should be valid")
	}
	if Gender(99).IsValid() {
		t.Error("Gender(99) should not be valid")
	}
}

func TestMood_Values(t *testing.T) {
	// Verify all moods are correctly sequenced
	if MoodHappy != 1 {
		t.Errorf("MoodHappy = %d, want 1", MoodHappy)
	}
	if MoodAlone != 20 {
		t.Errorf("MoodAlone = %d, want 20", MoodAlone)
	}
}

func TestCardPackType_Description(t *testing.T) {
	if desc := CardPackBrown.Description(); desc != "2 Level 1 cards" {
		t.Errorf("CardPackBrown.Description() = %v, want '2 Level 1 cards'", desc)
	}
}

func TestTribeStatus_String(t *testing.T) {
	tests := []struct {
		status   TribeStatus
		expected string
	}{
		{TribeStatusOpen, "OPEN"},
		{TribeStatusInviteOnly, "INVITE_ONLY"},
		{TribeStatusClosed, "CLOSED"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.status.String(); got != tt.expected {
				t.Errorf("TribeStatus.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestHeroCardPackType_IsValid(t *testing.T) {
	validHeroes := []HeroCardPackType{
		HeroCardXakhmi, HeroCardXebelus, HeroCardHushidar, HeroCardSibilu,
	}
	for _, hero := range validHeroes {
		if !hero.IsValid() {
			t.Errorf("%v should be valid", hero)
		}
	}
	
	if HeroCardPackType(999).IsValid() {
		t.Error("HeroCardPackType(999) should not be valid")
	}
}

func TestValidateEnum(t *testing.T) {
	// Valid enum
	if err := ValidateEnum(GenderMale); err != nil {
		t.Errorf("ValidateEnum(GenderMale) should not error: %v", err)
	}
	
	// Invalid enum
	if err := ValidateEnum(Gender(99)); err == nil {
		t.Error("ValidateEnum(Gender(99)) should error")
	}
}

// Benchmark tests for zero-allocation verification
func BenchmarkGenderString(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = GenderMale.String()
	}
}

func BenchmarkMoodString(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = MoodCrazy.String()
	}
}

func BenchmarkCardPackTypeString(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = CardPackGold.String()
	}
}

func BenchmarkEnumValidation(b *testing.B) {
	b.Run("Valid", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			GenderMale.IsValid()
		}
	})
	
	b.Run("Invalid", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			Gender(99).IsValid()
		}
	})
}