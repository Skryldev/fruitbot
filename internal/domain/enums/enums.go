package enums

import "fmt"

type Gender uint8

const (
	GenderUndefined Gender = iota
	GenderMale
	GenderFemale
)

var genderStrings = map[Gender]string{
	GenderUndefined: "UNDEFINED",
	GenderMale:      "MALE",
	GenderFemale:    "FEMALE",
}

func (g Gender) String() string {
	if s, ok := genderStrings[g]; ok {
		return s
	}
	return fmt.Sprintf("Gender(%d)", g)
}

func (g Gender) IsValid() bool {
	_, ok := genderStrings[g]
	return ok
}

// ============================================================
// Mood
// ============================================================

// Mood represents player mood/status
type Mood uint8

const (
	MoodHappy       Mood = iota + 1
	MoodAngry                        
	MoodAshamed                     
	MoodBored                       
	MoodCalm                      
	MoodCheerful                 
	MoodCold                   
	MoodProud                   
	MoodConfused           
	MoodCranky                    
	MoodCrazy                     
	MoodCurious                  
	MoodDepressed              
	MoodDisappointed               
	MoodGood                  
	MoodHopeful                    
	MoodHungry                    
	MoodIndifferent              
	MoodApathetic                   
	MoodAlone              
)

var moodStrings = map[Mood]string{
	MoodHappy:        "HAPPY",
	MoodAngry:        "ANGRY",
	MoodAshamed:      "ASHAMED",
	MoodBored:        "BORED",
	MoodCalm:         "CALM",
	MoodCheerful:     "CHEERFUL",
	MoodCold:         "COLD",
	MoodProud:        "PROUD",
	MoodConfused:     "CONFUSED",
	MoodCranky:       "CRANKY",
	MoodCrazy:        "CRAZY",
	MoodCurious:      "CURIOUS",
	MoodDepressed:    "DEPRESSED",
	MoodDisappointed: "DISAPPOINTED",
	MoodGood:         "GOOD",
	MoodHopeful:      "HOPEFUL",
	MoodHungry:       "HUNGRY",
	MoodIndifferent:  "INDIFFERENT",
	MoodApathetic:    "APATHETIC",
	MoodAlone:        "ALONE",
}

func (m Mood) String() string {
	if s, ok := moodStrings[m]; ok {
		return s
	}
	return fmt.Sprintf("Mood(%d)", m)
}

func (m Mood) IsValid() bool {
	_, ok := moodStrings[m]
	return ok
}

// ============================================================
// AuctionCategory
// ============================================================

// AuctionCategory represents card auction categories
type AuctionCategory uint8

const (
	AuctionCategoryCommon    AuctionCategory = iota // 0
	AuctionCategoryChristmas                        // 1
	AuctionCategoryMonster                          // 2
)

var auctionCategoryStrings = map[AuctionCategory]string{
	AuctionCategoryCommon:    "COMMON_CARDS",
	AuctionCategoryChristmas: "CHRISTMAS_CARDS",
	AuctionCategoryMonster:   "MONSTER_CARDS",
}

func (ac AuctionCategory) String() string {
	if s, ok := auctionCategoryStrings[ac]; ok {
		return s
	}
	return fmt.Sprintf("AuctionCategory(%d)", ac)
}

func (ac AuctionCategory) IsValid() bool {
	_, ok := auctionCategoryStrings[ac]
	return ok
}

// ============================================================
// AuctionPriceFilter
// ============================================================

// AuctionPriceFilter represents price sorting direction
type AuctionPriceFilter uint8

const (
	AuctionPriceLowest  AuctionPriceFilter = iota // 0
	AuctionPriceHighest                           // 1
)

var auctionPriceFilterStrings = map[AuctionPriceFilter]string{
	AuctionPriceLowest:  "LOWEST_PRICE",
	AuctionPriceHighest: "HIGHEST_PRICE",
}

func (apf AuctionPriceFilter) String() string {
	if s, ok := auctionPriceFilterStrings[apf]; ok {
		return s
	}
	return fmt.Sprintf("AuctionPriceFilter(%d)", apf)
}

// ============================================================
// BuildingType
// ============================================================

// BuildingType represents different building types in the game
type BuildingType uint16

const (
	BuildingGoldMine       BuildingType = 1001
	BuildingOffense        BuildingType = 1002
	BuildingDefense        BuildingType = 1003
)

var buildingTypeStrings = map[BuildingType]string{
	BuildingGoldMine: "GOLD_MINE",
	BuildingOffense:  "OFFENSE_BUILDING",
	BuildingDefense:  "DEFENSE_BUILDING",
}

func (bt BuildingType) String() string {
	if s, ok := buildingTypeStrings[bt]; ok {
		return s
	}
	return fmt.Sprintf("BuildingType(%d)", bt)
}

func (bt BuildingType) IsValid() bool {
	_, ok := buildingTypeStrings[bt]
	return ok
}

// ============================================================
// CardPackType
// ============================================================

// CardPackType represents different card pack types with their contents
type CardPackType uint8

const (
	CardPackBrown     CardPackType = 1
	CardPackGreen     CardPackType = 2
	CardPackYellow    CardPackType = 3
	CardPackRed       CardPackType = 4
	CardPackSilver    CardPackType = 5
	CardPackGold      CardPackType = 6
	CardPackPlatinum  CardPackType = 7
	CardPackBlack     CardPackType = 8
	CardPackMonster   CardPackType = 16
	CardPackCrystal   CardPackType = 25
	CardPackHero      CardPackType = 32
)

var cardPackTypeStrings = map[CardPackType]string{
	CardPackBrown:    "BROWN_PACK",
	CardPackGreen:    "GREEN_PACK",
	CardPackYellow:   "YELLOW_PACK",
	CardPackRed:      "RED_PACK",
	CardPackSilver:   "SILVER_PACK",
	CardPackGold:     "GOLD_PACK",
	CardPackPlatinum: "PLATINUM_PACK",
	CardPackBlack:    "BLACK_PACK",
	CardPackMonster:  "MONSTER_PACK",
	CardPackCrystal:  "CRYSTAL_PACK",
	CardPackHero:     "HERO_PACK",
}

// CardPackDescription provides details about each pack's contents
var cardPackDescriptions = map[CardPackType]string{
	CardPackBrown:    "2 Level 1 cards",
	CardPackGreen:    "2 Level 2 cards",
	CardPackYellow:   "2 Level 3 cards",
	CardPackRed:      "2 Level 3 or 4 cards",
	CardPackSilver:   "20 Level 1 cards and 10 Level 2 cards",
	CardPackGold:     "2 Level 4 or 5 cards",
	CardPackPlatinum: "2 Level 5 or 6 cards",
	CardPackBlack:    "2 Level 6 or 7 cards",
	CardPackMonster:  "1 Super Powerful Monster card",
	CardPackCrystal:  "1 Crystal card",
	CardPackHero:     "1 Hero card of your choice",
}

func (cpt CardPackType) String() string {
	if s, ok := cardPackTypeStrings[cpt]; ok {
		return s
	}
	return fmt.Sprintf("CardPackType(%d)", cpt)
}

func (cpt CardPackType) Description() string {
	if d, ok := cardPackDescriptions[cpt]; ok {
		return d
	}
	return "Unknown pack"
}

func (cpt CardPackType) IsValid() bool {
	_, ok := cardPackTypeStrings[cpt]
	return ok
}

// ============================================================
// HeroCardPackType
// ============================================================

// HeroCardPackType represents specific hero card types
type HeroCardPackType uint16

const (
	HeroCardXakhmi   HeroCardPackType = 415
	HeroCardXebelus  HeroCardPackType = 515
	HeroCardHushidar HeroCardPackType = 615
	HeroCardSibilu   HeroCardPackType = 715
)

var heroCardPackStrings = map[HeroCardPackType]string{
	HeroCardXakhmi:   "XAKHMI",
	HeroCardXebelus:  "XEBELUS",
	HeroCardHushidar: "HUSHIDAR",
	HeroCardSibilu:   "SIBILU",
}

func (hct HeroCardPackType) String() string {
	if s, ok := heroCardPackStrings[hct]; ok {
		return s
	}
	return fmt.Sprintf("HeroCardPackType(%d)", hct)
}

func (hct HeroCardPackType) IsValid() bool {
	_, ok := heroCardPackStrings[hct]
	return ok
}

// ============================================================
// TribeCapability
// ============================================================

// TribeCapability represents different tribe upgrade capabilities
type TribeCapability uint16

const (
	TribeCapAttackBonus    TribeCapability = 1002
	TribeCapDefenseBonus   TribeCapability = 1003
	TribeCapRecoveryTime   TribeCapability = 1004
	TribeCapTribeCapacity  TribeCapability = 1005
)

var tribeCapabilityStrings = map[TribeCapability]string{
	TribeCapAttackBonus:   "ATTACK_BONUS",
	TribeCapDefenseBonus:  "DEFENSE_BONUS",
	TribeCapRecoveryTime:  "RECOVERY_TIME",
	TribeCapTribeCapacity: "TRIBE_CAPACITY",
}

func (tc TribeCapability) String() string {
	if s, ok := tribeCapabilityStrings[tc]; ok {
		return s
	}
	return fmt.Sprintf("TribeCapability(%d)", tc)
}

func (tc TribeCapability) IsValid() bool {
	_, ok := tribeCapabilityStrings[tc]
	return ok
}

// ============================================================
// TribeStatus
// ============================================================

// TribeStatus represents tribe membership status
type TribeStatus uint8

const (
	TribeStatusOpen       TribeStatus = 1
	TribeStatusInviteOnly TribeStatus = 2
	TribeStatusClosed     TribeStatus = 3
)

var tribeStatusStrings = map[TribeStatus]string{
	TribeStatusOpen:       "OPEN",
	TribeStatusInviteOnly: "INVITE_ONLY",
	TribeStatusClosed:     "CLOSED",
}

func (ts TribeStatus) String() string {
	if s, ok := tribeStatusStrings[ts]; ok {
		return s
	}
	return fmt.Sprintf("TribeStatus(%d)", ts)
}

func (ts TribeStatus) IsValid() bool {
	_, ok := tribeStatusStrings[ts]
	return ok
}

// ============================================================
// Generic Enum Validation Helper
// ============================================================

// Validator is an interface for types that can be validated
type Validator interface {
	IsValid() bool
}

// ValidateEnum is a generic function to validate any enum type
func ValidateEnum[T Validator](v T) error {
	if !v.IsValid() {
		return fmt.Errorf("invalid enum value: %v", v)
	}
	return nil
}