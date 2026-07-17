package config

import (
	"fmt"
	"sync"
	"time"

	domainErrors "fruitbot/internal/domain/errors"
)

// ============================================================
// Error Mapper - Maps server error codes to domain errors
// ============================================================

type ErrorMapping struct {
	Code    domainErrors.ErrorCode `json:"code"`
	Message string                 `json:"message"`
	Error   *domainErrors.DomainError `json:"-"`
}

type ErrorMapper struct {
	mu       sync.RWMutex
	mappings map[domainErrors.ErrorCode]*domainErrors.DomainError
}

func NewErrorMapper() *ErrorMapper {
	em := &ErrorMapper{
		mappings: make(map[domainErrors.ErrorCode]*domainErrors.DomainError, 200),
	}
	em.registerDefaults()
	return em
}

func (em *ErrorMapper) registerDefaults() {
	// This mapping mirrors the Python configs.py exceptions dict
	mappings := map[domainErrors.ErrorCode]*domainErrors.DomainError{
		domainErrors.CodeGeneralError:              domainErrors.ErrGeneral,
		domainErrors.CodeAccountBlocked:            domainErrors.ErrAccountBlocked,
		domainErrors.CodeCardNotFound:              domainErrors.ErrCardNotFound,
		domainErrors.CodeNoCardSelected:            domainErrors.ErrNoCardSelected,
		domainErrors.CodeCardAlreadyInAuction:      domainErrors.ErrCardAlreadyInAuction,
		domainErrors.CodeNotEnoughCards:            domainErrors.ErrNotEnoughCards,
		domainErrors.CodeInconsistencyError:        domainErrors.ErrInconsistency,
		domainErrors.CodeCardNotInAuction:          domainErrors.ErrCardNotInAuction,
		domainErrors.CodeMaxBidReached:             domainErrors.ErrMaxBidReached,
		domainErrors.CodeCannotBidOwnCard:          domainErrors.ErrCannotBidOwnCard,
		domainErrors.CodeAlreadyHighestBidder:      domainErrors.ErrAlreadyHighestBidder,
		domainErrors.CodeAuctionClosed:             domainErrors.ErrAuctionClosed,
		domainErrors.CodeCardTypeQuery:             domainErrors.ErrCardTypeQuery,
		domainErrors.CodeCardTypeNotFound:          domainErrors.ErrCardTypeNotFound,
		domainErrors.CodeCardTypeQueryDuplicate:    domainErrors.ErrCardTypeQueryDuplicate,
		domainErrors.CodeNotOwner:                  domainErrors.ErrNotOwner,
		domainErrors.CodeAccessDenied:              domainErrors.ErrAccessDenied,
		domainErrors.CodeServerMaintenance:         domainErrors.ErrServerMaintenance,
		domainErrors.CodeNoCardsSpecified:          domainErrors.ErrNoCardsSpecified,
		domainErrors.CodeNoOpponentSpecified:       domainErrors.ErrNoOpponentSpecified,
		domainErrors.CodeBattleNotAvailable:        domainErrors.ErrBattleNotAvailable,
		domainErrors.CodeCannotAttackSelf:          domainErrors.ErrCannotAttackSelf,
		domainErrors.CodePlayerProtected:           domainErrors.ErrPlayerProtected,
		domainErrors.CodeCaptchaRequired:           domainErrors.ErrCaptchaRequired,
		domainErrors.CodePlayingOnAnotherDevice:    domainErrors.ErrPlayingOnAnotherDevice,
		domainErrors.CodeOpponentNotFound:          domainErrors.ErrOpponentNotFound,
		domainErrors.CodeOpponentOutOfRange:        domainErrors.ErrOpponentOutOfRange,
		domainErrors.CodeOpponentDefenseEmpty:      domainErrors.ErrOpponentDefenseEmpty,
		domainErrors.CodeInvalidCaptcha:            domainErrors.ErrInvalidCaptcha,
		domainErrors.CodeCardInUse:                 domainErrors.ErrCardInUse,
		domainErrors.CodeCardMarkedForSacrifice:    domainErrors.ErrCardMarkedForSacrifice,
		domainErrors.CodeMaxPowerReached:           domainErrors.ErrMaxPowerReached,
		domainErrors.CodeMaxEvolutionsExceeded:     domainErrors.ErrMaxEvolutionsExceeded,
		domainErrors.CodeEvolveLimitExceeded:       domainErrors.ErrEvolveLimitExceeded,
		domainErrors.CodeBuildingNotFound:          domainErrors.ErrBuildingNotFound,
		domainErrors.CodeMaxCardsAssigned:          domainErrors.ErrMaxCardsAssigned,
		domainErrors.CodeLiveBattleUnavailable:     domainErrors.ErrLiveBattleUnavailable,
		domainErrors.CodeOpponentOffline:           domainErrors.ErrOpponentOffline,
		domainErrors.CodeAlreadyInLiveBattle:       domainErrors.ErrAlreadyInLiveBattle,
		domainErrors.CodeOpponentBusy:              domainErrors.ErrOpponentBusy,
		domainErrors.CodeCannotAttackTribeMate:     domainErrors.ErrCannotAttackTribeMate,
		domainErrors.CodeBattleIDRequired:          domainErrors.ErrBattleIDRequired,
		domainErrors.CodeInconsistencyErrorUpdate:  domainErrors.ErrInconsistencyUpdate,
		domainErrors.CodeOperationTimeout:          domainErrors.ErrOperationTimeout,
		domainErrors.CodePlayerUnavailable:         domainErrors.ErrPlayerUnavailable,
		domainErrors.CodeBattleNotFound:            domainErrors.ErrBattleNotFound,
		domainErrors.CodeAttackerNotFound:          domainErrors.ErrAttackerNotFound,
		domainErrors.CodeDefenderNotFound:          domainErrors.ErrDefenderNotFound,
		domainErrors.CodeTribeNotFound:             domainErrors.ErrTribeNotFound,
		domainErrors.CodeNotInTribe:                domainErrors.ErrNotInTribe,
		domainErrors.CodeTribeHelpLimit:            domainErrors.ErrTribeHelpLimit,
		domainErrors.CodeMessageSizeExceeded:       domainErrors.ErrMessageSizeExceeded,
		domainErrors.CodeMessageNotFound:           domainErrors.ErrMessageNotFound,
		domainErrors.CodeInvalidInvitationTicket:   domainErrors.ErrInvalidInvitationTicket,
		domainErrors.CodeInvalidRestoreKey:         domainErrors.ErrInvalidRestoreKey,
		domainErrors.CodeAccountChangeLimit:        domainErrors.ErrAccountChangeLimit,
		domainErrors.CodeIdentificationRequired:    domainErrors.ErrIdentificationRequired,
		domainErrors.CodePlayerNotFound:            domainErrors.ErrPlayerNotFound,
		domainErrors.CodeInvitationCodeAlreadyRedeemed: domainErrors.ErrInvitationCodeAlreadyRedeemed,
		domainErrors.CodeInvalidEmail:              domainErrors.ErrInvalidEmail,
		domainErrors.CodeEmailAlreadyExists:        domainErrors.ErrEmailAlreadyExists,
		domainErrors.CodeEmailAlreadyRegistered:    domainErrors.ErrEmailAlreadyRegistered,
		domainErrors.CodeRewardAlreadyReceived:     domainErrors.ErrRewardAlreadyReceived,
		domainErrors.CodeAccountDeactivated:        domainErrors.ErrAccountDeactivated,
		domainErrors.CodeAccountResetFailed:        domainErrors.ErrAccountResetFailed,
		domainErrors.CodeInvalidActivationCode:     domainErrors.ErrInvalidActivationCode,
		domainErrors.CodeCashRewardNotFound:        domainErrors.ErrCashRewardNotFound,
		domainErrors.CodeAvatarNotAvailable:        domainErrors.ErrAvatarNotAvailable,
		domainErrors.CodeNotInLeague:               domainErrors.ErrNotInLeague,
		domainErrors.CodeDeviceVerificationError:   domainErrors.ErrDeviceVerification,
		domainErrors.CodeStoreNotSupported:         domainErrors.ErrStoreNotSupported,
		domainErrors.CodePackInconsistencyError:    domainErrors.ErrPackInconsistency,
		domainErrors.CodePaymentReceiptNotProvided: domainErrors.ErrPaymentReceiptNotProvided,
		domainErrors.CodeCardCannotEvolve:          domainErrors.ErrCardCannotEvolve,
		domainErrors.CodeInternalError:             domainErrors.ErrInternal,
		domainErrors.CodeInternalErrorWait:         domainErrors.ErrInternalWait,
		domainErrors.CodePlayerNotParticipated:     domainErrors.ErrPlayerNotParticipated,
		domainErrors.CodeCardCoolingDown:           domainErrors.ErrCardCoolingDown,
		domainErrors.CodeCardCoolEnough:            domainErrors.ErrCardCoolEnough,
		domainErrors.CodeDeviceInconsistency:       domainErrors.ErrDeviceInconsistency,
		domainErrors.CodeNameTooLong:               domainErrors.ErrNameTooLong,
		domainErrors.CodeNameAlreadyTaken:          domainErrors.ErrNameAlreadyTaken,
		domainErrors.CodeCountryCodeNotSupported:   domainErrors.ErrCountryCodeNotSupported,
		domainErrors.CodeNotEnoughGold:             domainErrors.ErrNotEnoughGold,
		domainErrors.CodeOnlineOnAnotherDevice:     domainErrors.ErrOnlineOnAnotherDevice,
		domainErrors.CodeTribeNameTooLong:          domainErrors.ErrTribeNameTooLong,
		domainErrors.CodeTopTribeEditLimit:         domainErrors.ErrTopTribeEditLimit,
		domainErrors.CodeCannotSellCard:            domainErrors.ErrCannotSellCard,
		domainErrors.CodeStatusRequired:            domainErrors.ErrStatusRequired,
		domainErrors.CodeNameRequired:              domainErrors.ErrNameRequired,
		domainErrors.CodeDescriptionRequired:       domainErrors.ErrDescriptionRequired,
		domainErrors.CodeDescriptionTooLong:        domainErrors.ErrDescriptionTooLong,
		domainErrors.CodeTribeNameExists:           domainErrors.ErrTribeNameExists,
		domainErrors.CodeInvalidChiefPlayers:       domainErrors.ErrInvalidChiefPlayers,
		domainErrors.CodeAlreadyInTribe:            domainErrors.ErrAlreadyInTribe,
		domainErrors.CodeNoTribeBuilding:           domainErrors.ErrNoTribeBuilding,
		domainErrors.CodeMaxTribeChangeLimit:       domainErrors.ErrMaxTribeChangeLimit,
		domainErrors.CodeTribeNoMembers:            domainErrors.ErrTribeNoMembers,
		domainErrors.CodeUndecidedRequest:          domainErrors.ErrUndecidedRequest,
		domainErrors.CodeInvalidDecisionParameter:  domainErrors.ErrInvalidDecisionParam,
		domainErrors.CodeNotJoinRequest:            domainErrors.ErrNotJoinRequest,
		domainErrors.CodeInconsistentDataProvided:  domainErrors.ErrInconsistentData,
		domainErrors.CodeJoinRequestProcessed:      domainErrors.ErrJoinRequestProcessed,
		domainErrors.CodeTribeFull:                 domainErrors.ErrTribeFull,
		domainErrors.CodeTribeAccessDenied:         domainErrors.ErrTribeAccessDenied,
		domainErrors.CodePlayerAlreadyInTribe:      domainErrors.ErrPlayerAlreadyInTribe,
		domainErrors.CodeNotAnInvitation:           domainErrors.ErrNotAnInvitation,
		domainErrors.CodeInvitationProcessed:       domainErrors.ErrInvitationProcessed,
		domainErrors.CodeNoTribesAvailable:         domainErrors.ErrNoTribesAvailable,
		domainErrors.CodePlayerNotInYourTribe:      domainErrors.ErrPlayerNotInYourTribe,
		domainErrors.CodePlayerAlreadyPromoted:     domainErrors.ErrPlayerAlreadyPromoted,
		domainErrors.CodePlayerNotElder:            domainErrors.ErrPlayerNotElder,
		domainErrors.CodeSelfPoke:                  domainErrors.ErrSelfPoke,
		domainErrors.CodeNotAMemberOfTribe:         domainErrors.ErrNotAMemberOfTribe,
		domainErrors.CodeSelfKick:                  domainErrors.ErrSelfKick,
		domainErrors.CodeInsufficientPermission:    domainErrors.ErrInsufficientPermission,
		domainErrors.CodeMaxLevelCooldownBuilding:  domainErrors.ErrMaxLevelCooldown,
		domainErrors.CodeMaxLevelMainHallBuilding:  domainErrors.ErrMaxLevelMainHall,
		domainErrors.CodeMaxLevelDefenseBuilding:   domainErrors.ErrMaxLevelDefense,
		domainErrors.CodeMaxLevelOffenseBuilding:   domainErrors.ErrMaxLevelOffense,
		domainErrors.CodeMaxLevelGoldBuilding:      domainErrors.ErrMaxLevelGold,
		domainErrors.CodeMaxLevelBankBuilding:      domainErrors.ErrMaxLevelBank,
		domainErrors.CodeInvalidBuildingTypeForUpgrade: domainErrors.ErrInvalidBuildingType,
		domainErrors.CodeMinimumDonationAmount:     domainErrors.ErrMinimumDonation,
		domainErrors.CodePlayerNoTribeBuilding:     domainErrors.ErrPlayerNoTribeBuilding,
		domainErrors.CodeUndecidedInvitation:       domainErrors.ErrUndecidedInvitation,
		domainErrors.CodeNotEnoughTribeMoney:       domainErrors.ErrNotEnoughTribeMoney,
		domainErrors.CodeTribeScoreUpdateFailed:    domainErrors.ErrTribeScoreUpdate,
		domainErrors.CodeTribeNoIdentifier:         domainErrors.ErrTribeNoIdentifier,
		domainErrors.CodeInvalidBuildingTypeForCardCapacity: domainErrors.ErrInvalidBuildingCardCapacity,
		domainErrors.CodeOperationFailed:           domainErrors.ErrOperationFailed,
		domainErrors.CodeGooglePlayVerificationFailed: domainErrors.ErrGooglePlayVerification,
		domainErrors.CodeInvalidPurchaseState:      domainErrors.ErrInvalidPurchaseState,
		domainErrors.CodeDataReadingError:          domainErrors.ErrDataReading,
		domainErrors.CodeSibcheVerificationFailed:  domainErrors.ErrSibcheVerification,
		domainErrors.CodeMaxBoostsLimit:            domainErrors.ErrMaxBoostsLimit,
		domainErrors.CodeInvalidCountryCode:        domainErrors.ErrInvalidCountryCode,
		domainErrors.CodeUserNotFound:              domainErrors.ErrUserNotFound,
		domainErrors.CodeUserRecentlyPoked:         domainErrors.ErrUserRecentlyPoked,
		domainErrors.CodeLeagueUpdateInProgress:    domainErrors.ErrLeagueUpdateInProgress,
		domainErrors.CodeFeatureNotImplemented:     domainErrors.ErrFeatureNotImplemented,
		domainErrors.CodeInvalidLeagueID:           domainErrors.ErrInvalidLeagueID,
		domainErrors.CodeTribeHelpAlreadyInProgress: domainErrors.ErrTribeHelpInProgress,
		domainErrors.CodeTutorialUpdateParameters:  domainErrors.ErrTutorialUpdateParams,
		domainErrors.CodeNoTribeAvailableToCoach:   domainErrors.ErrNoTribeAvailableToCoach,
		domainErrors.CodeNotEnoughNectar:           domainErrors.ErrNotEnoughNectar,
		domainErrors.CodeHeroItemNotPurchased:      domainErrors.ErrHeroItemNotPurchased,
		domainErrors.CodeHeroItemAlreadyPurchased:  domainErrors.ErrHeroItemAlreadyPurchased,
		domainErrors.CodeAllHeroesPurchased:        domainErrors.ErrAllHeroesPurchased,
		domainErrors.CodeNotEnoughPotion:           domainErrors.ErrNotEnoughPotion,
		domainErrors.CodeInvalidGiftCode:           domainErrors.ErrInvalidGiftCode,
		domainErrors.CodeGiftCodeAlreadyRedeemed:   domainErrors.ErrGiftCodeAlreadyRedeemed,
		domainErrors.CodeGiftCodeExpired:           domainErrors.ErrGiftCodeExpired,
		domainErrors.CodeHeroLevelRequirement:      domainErrors.ErrHeroLevelRequirement,
		domainErrors.CodeTribeEntryNotAllowed:      domainErrors.ErrTribeEntryNotAllowed,
		domainErrors.CodeLevelRequirement:          domainErrors.ErrLevelRequirement,
		domainErrors.CodePrizeAlreadyReceived:      domainErrors.ErrPrizeAlreadyReceived,
		domainErrors.CodeInvalidMobileNumber:       domainErrors.ErrInvalidMobileNumber,
		domainErrors.CodeInvalidVerificationCode:   domainErrors.ErrInvalidVerificationCode,
		domainErrors.CodeNotSubscribed:             domainErrors.ErrNotSubscribed,
		domainErrors.CodeNotCharged:                domainErrors.ErrNotCharged,
		domainErrors.CodeMaxTribeBroadcastReached:  domainErrors.ErrMaxTribeBroadcast,
		domainErrors.CodeBundlePurchaseError:       domainErrors.ErrBundlePurchase,
		domainErrors.CodeTooManyRequests:           domainErrors.ErrTooManyRequests,
		domainErrors.CodeUnknown:                   domainErrors.ErrUnknown,
	}

	em.mu.Lock()
	for code, err := range mappings {
		em.mappings[code] = err
	}
	em.mu.Unlock()
}

func (em *ErrorMapper) MapError(code domainErrors.ErrorCode) (*domainErrors.DomainError, bool) {
	em.mu.RLock()
	err, ok := em.mappings[code]
	em.mu.RUnlock()
	return err, ok
}

func (em *ErrorMapper) MustMapError(code domainErrors.ErrorCode) *domainErrors.DomainError {
	if err, ok := em.MapError(code); ok {
		return err
	}
	return domainErrors.ErrUnknown
}

// ============================================================
// Device Fingerprint Configuration
// ============================================================

type DeviceModel struct {
	Name     string `json:"name"`
	Brand    string `json:"brand"`
	Model    string `json:"model"`
	OSVersion string `json:"os_version"`
}

type DeviceFingerprinter struct {
	mu           sync.RWMutex
	models       []string
	modelObjects []DeviceModel
	index        int
}

func NewDeviceFingerprinter() *DeviceFingerprinter {
	df := &DeviceFingerprinter{
		models: make([]string, 0, 50),
	}
	df.registerDefaultModels()
	return df
}

func (df *DeviceFingerprinter) registerDefaultModels() {
	defaultModels := []string{
		"Samsung Galaxy S22 Ultra",
		"Samsung Galaxy S21 FE",
		"Samsung Galaxy A52",
		"Samsung Galaxy A72",
		"Google Pixel 6 Pro",
		"Google Pixel 6",
		"Google Pixel 5a",
		"OnePlus 9 Pro",
		"OnePlus 9",
		"OnePlus Nord 2",
		"Xiaomi Mi 11 Ultra",
		"Xiaomi Mi 11",
		"Xiaomi Redmi Note 11 Pro",
		"Xiaomi Poco F3",
		"Huawei P50 Pro",
		"Huawei Mate 40 Pro",
		"Huawei Nova 9",
		"Oppo Find X5 Pro",
		"Oppo Reno 7 Pro",
		"Motorola Edge 20 Pro",
		"Motorola Moto G Stylus 2022",
		"Sony Xperia 1 III",
		"Sony Xperia 5 III",
		"Nokia X20",
		"Nokia G50",
		"Realme GT 2 Pro",
		"Realme 8 Pro",
		"Vivo X70 Pro",
		"Vivo V21 5G",
		"Asus ROG Phone 5",
		"Asus Zenfone 8",
		"TCL 20 Pro 5G",
		"ZTE Axon 20 5G",
		"Honor 50 Pro",
		"Honor 50",
		"Infinix Zero 5G",
		"Lava Agni 5G",
		"Micromax IN Note 2",
	}

	df.mu.Lock()
	df.models = make([]string, len(defaultModels))
	copy(df.models, defaultModels)
	df.mu.Unlock()
}

func (df *DeviceFingerprinter) GetRandomModel() string {
	df.mu.RLock()
	defer df.mu.RUnlock()
	
	if len(df.models) == 0 {
		return "Unknown Device"
	}
	
	idx := df.index % len(df.models)
	df.index++
	
	return df.models[idx]
}

func (df *DeviceFingerprinter) GetAllModels() []string {
	df.mu.RLock()
	defer df.mu.RUnlock()
	
	models := make([]string, len(df.models))
	copy(models, df.models)
	return models
}

func (df *DeviceFingerprinter) AddModel(model string) {
	df.mu.Lock()
	df.models = append(df.models, model)
	df.mu.Unlock()
}

// ============================================================
// Application Configuration
// ============================================================

type Config struct {
	ErrorMapper        *ErrorMapper
	DeviceFingerprinter *DeviceFingerprinter
	
	// Server configuration
	ServerAddr     string        `json:"server_addr"`
	ConnectTimeout time.Duration `json:"connect_timeout"`
	ReadTimeout    time.Duration `json:"read_timeout"`
	WriteTimeout   time.Duration `json:"write_timeout"`
	
	// Session configuration
	MaxRetries     int           `json:"max_retries"`
	RetryDelay     time.Duration `json:"retry_delay"`
	SessionTTL     time.Duration `json:"session_ttl"`
	
	// Rate limiting
	MaxRequestsPerSecond int `json:"max_requests_per_second"`
	BurstSize            int `json:"burst_size"`
}

func DefaultConfig() *Config {
	return &Config{
		ErrorMapper:         NewErrorMapper(),
		DeviceFingerprinter: NewDeviceFingerprinter(),
		ServerAddr:          "game.fruitcraft.com:443",
		ConnectTimeout:      10 * time.Second,
		ReadTimeout:         30 * time.Second,
		WriteTimeout:        30 * time.Second,
		MaxRetries:          3,
		RetryDelay:          1 * time.Second,
		SessionTTL:          24 * time.Hour,
		MaxRequestsPerSecond: 10,
		BurstSize:           5,
	}
}

type Option func(*Config)

func WithServerAddr(addr string) Option {
	return func(c *Config) {
		c.ServerAddr = addr
	}
}

func WithTimeouts(connect, read, write time.Duration) Option {
	return func(c *Config) {
		c.ConnectTimeout = connect
		c.ReadTimeout = read
		c.WriteTimeout = write
	}
}

func WithRetryConfig(maxRetries int, delay time.Duration) Option {
	return func(c *Config) {
		c.MaxRetries = maxRetries
		c.RetryDelay = delay
	}
}

func NewConfig(opts ...Option) *Config {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

func (c *Config) String() string {
	return fmt.Sprintf(
		"Config{Server: %s, Timeouts: connect=%v/read=%v/write=%v, Retries: %d, RateLimit: %d/s}",
		c.ServerAddr, c.ConnectTimeout, c.ReadTimeout, c.WriteTimeout,
		c.MaxRetries, c.MaxRequestsPerSecond,
	)
}

func (c *Config) Validate() error {
	if c.ServerAddr == "" {
		return fmt.Errorf("server address is required")
	}
	if c.ConnectTimeout <= 0 {
		return fmt.Errorf("connect timeout must be positive")
	}
	if c.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}
	return nil
}