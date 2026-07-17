package errors

import (
	"errors"
	"fmt"
	"sync"
)

type ErrorCode int32

const (
	CodeGeneralError              ErrorCode = 100
	CodeAccountBlocked            ErrorCode = 101
	CodeCardNotFound              ErrorCode = 102
	CodeNoCardSelected            ErrorCode = 103
	CodeCardAlreadyInAuction      ErrorCode = 104
	CodeNotEnoughCards            ErrorCode = 105
	CodeInconsistencyError        ErrorCode = 106
	CodeCardNotInAuction          ErrorCode = 107
	CodeMaxBidReached             ErrorCode = 108
	CodeCannotBidOwnCard          ErrorCode = 109
	CodeAlreadyHighestBidder      ErrorCode = 110
	CodeAuctionClosed             ErrorCode = 111
	CodeCardTypeQuery             ErrorCode = 112
	CodeCardTypeNotFound          ErrorCode = 113
	CodeCardTypeQueryDuplicate    ErrorCode = 114
	CodeNotOwner                  ErrorCode = 115
	CodeAccessDenied              ErrorCode = 116
	CodeServerMaintenance         ErrorCode = 117
	CodeNoCardsSpecified          ErrorCode = 118
	CodeNoOpponentSpecified       ErrorCode = 119
	CodeBattleNotAvailable        ErrorCode = 120
	CodeCannotAttackSelf          ErrorCode = 121
	CodePlayerProtected           ErrorCode = 122
	CodeCaptchaRequired           ErrorCode = 123
	CodePlayingOnAnotherDevice    ErrorCode = 124
	CodeOpponentNotFound          ErrorCode = 125
	CodeOpponentOutOfRange        ErrorCode = 126
	CodeOpponentDefenseEmpty      ErrorCode = 127
	CodeInvalidCaptcha            ErrorCode = 128
	CodeCardInUse                 ErrorCode = 129
	CodeCardMarkedForSacrifice    ErrorCode = 130
	CodeMaxPowerReached           ErrorCode = 131
	CodeMaxEvolutionsExceeded     ErrorCode = 132
	CodeEvolveLimitExceeded       ErrorCode = 133
	CodeBuildingNotFound          ErrorCode = 134
	CodeMaxCardsAssigned          ErrorCode = 135
	CodeLiveBattleUnavailable     ErrorCode = 136
	CodeOpponentOffline           ErrorCode = 137
	CodeAlreadyInLiveBattle       ErrorCode = 138
	CodeOpponentBusy              ErrorCode = 139
	CodeCannotAttackTribeMate     ErrorCode = 140
	CodeBattleIDRequired          ErrorCode = 141
	CodeInconsistencyErrorUpdate  ErrorCode = 142
	CodeOperationTimeout          ErrorCode = 143
	CodePlayerUnavailable         ErrorCode = 144
	CodeBattleNotFound            ErrorCode = 145
	CodeAttackerNotFound          ErrorCode = 146
	CodeDefenderNotFound          ErrorCode = 147
	CodeTribeNotFound             ErrorCode = 148
	CodeNotInTribe                ErrorCode = 149
	CodeTribeHelpLimit            ErrorCode = 150
	CodeMessageSizeExceeded       ErrorCode = 151
	CodeMessageNotFound           ErrorCode = 152
	CodeInvalidInvitationTicket   ErrorCode = 153
	CodeInvalidRestoreKey         ErrorCode = 154
	CodeAccountChangeLimit        ErrorCode = 155
	CodeIdentificationRequired    ErrorCode = 156
	CodePlayerNotFound            ErrorCode = 157
	CodeInvitationCodeAlreadyRedeemed ErrorCode = 158
	CodeInvalidEmail              ErrorCode = 159
	CodeEmailAlreadyExists        ErrorCode = 160
	CodeEmailAlreadyRegistered    ErrorCode = 161
	CodeRewardAlreadyReceived     ErrorCode = 162
	CodeAccountDeactivated        ErrorCode = 163
	CodeAccountResetFailed        ErrorCode = 164
	CodeInvalidActivationCode     ErrorCode = 165
	CodeCashRewardNotFound        ErrorCode = 166
	CodeAvatarNotAvailable        ErrorCode = 167
	CodeNotInLeague               ErrorCode = 168
	CodeDeviceVerificationError   ErrorCode = 169
	CodeStoreNotSupported         ErrorCode = 170
	CodePackInconsistencyError    ErrorCode = 171
	CodePaymentReceiptNotProvided ErrorCode = 172
	CodeCardCannotEvolve          ErrorCode = 173
	CodeInternalError             ErrorCode = 174
	CodeInternalErrorWait         ErrorCode = 175
	CodePlayerNotParticipated     ErrorCode = 176
	CodeCardCoolingDown           ErrorCode = 177
	CodeCardCoolEnough            ErrorCode = 178
	CodeDeviceInconsistency       ErrorCode = 179
	CodeNameTooLong               ErrorCode = 180
	CodeNameAlreadyTaken          ErrorCode = 181
	CodeCountryCodeNotSupported   ErrorCode = 182
	CodeNotEnoughGold             ErrorCode = 183
	CodeOnlineOnAnotherDevice     ErrorCode = 184
	CodeTribeNameTooLong          ErrorCode = 185
	CodeTopTribeEditLimit         ErrorCode = 186
	CodeCannotSellCard            ErrorCode = 187
	CodeStatusRequired            ErrorCode = 188
	CodeNameRequired             ErrorCode = 190
	CodeDescriptionRequired      ErrorCode = 191
	CodeDescriptionTooLong       ErrorCode = 192
	CodeTribeNameExists          ErrorCode = 193
	CodeInvalidChiefPlayers      ErrorCode = 194
	CodeAlreadyInTribe           ErrorCode = 195
	CodeNoTribeBuilding          ErrorCode = 196
	CodeMaxTribeChangeLimit      ErrorCode = 197
	CodeTribeNoMembers           ErrorCode = 198
	CodeUndecidedRequest         ErrorCode = 199
	CodeInvalidDecisionParameter ErrorCode = 200
	CodeNotJoinRequest           ErrorCode = 201
	CodeInconsistentDataProvided ErrorCode = 202
	CodeJoinRequestProcessed     ErrorCode = 203
	CodeTribeFull                ErrorCode = 204
	CodeTribeAccessDenied        ErrorCode = 205
	CodePlayerAlreadyInTribe     ErrorCode = 206
	CodeNotAnInvitation          ErrorCode = 207
	CodeInvitationProcessed      ErrorCode = 208
	CodeNoTribesAvailable        ErrorCode = 209
	CodePlayerNotInYourTribe     ErrorCode = 210
	CodePlayerAlreadyPromoted    ErrorCode = 211
	CodePlayerNotElder           ErrorCode = 212
	CodeSelfPoke                 ErrorCode = 213
	CodeNotAMemberOfTribe        ErrorCode = 214
	CodeSelfKick                 ErrorCode = 215
	CodeInsufficientPermission   ErrorCode = 216
	CodeMaxLevelCooldownBuilding  ErrorCode = 217
	CodeMaxLevelMainHallBuilding  ErrorCode = 218
	CodeMaxLevelDefenseBuilding   ErrorCode = 219
	CodeMaxLevelOffenseBuilding   ErrorCode = 220
	CodeMaxLevelGoldBuilding      ErrorCode = 221
	CodeMaxLevelBankBuilding      ErrorCode = 222
	CodeInvalidBuildingTypeForUpgrade ErrorCode = 223
	CodeMinimumDonationAmount     ErrorCode = 224
	CodePlayerNoTribeBuilding     ErrorCode = 225
	CodeUndecidedInvitation       ErrorCode = 226
	CodeNotEnoughTribeMoney       ErrorCode = 227
	CodeTribeScoreUpdateFailed    ErrorCode = 228
	CodeTribeNoIdentifier         ErrorCode = 229
	CodeInvalidBuildingTypeForCardCapacity ErrorCode = 230
	CodeOperationFailed           ErrorCode = 231
	CodeGooglePlayVerificationFailed ErrorCode = 232
	CodeInvalidPurchaseState      ErrorCode = 233
	CodeDataReadingError          ErrorCode = 234
	CodeSibcheVerificationFailed  ErrorCode = 235
	CodeMaxBoostsLimit            ErrorCode = 236
	CodeInvalidCountryCode        ErrorCode = 237
	CodeUserNotFound              ErrorCode = 238
	CodeUserRecentlyPoked         ErrorCode = 239
	CodeLeagueUpdateInProgress    ErrorCode = 240
	CodeFeatureNotImplemented     ErrorCode = 241
	CodeInvalidLeagueID           ErrorCode = 242
	CodeTribeHelpAlreadyInProgress ErrorCode = 243
	CodeTutorialUpdateParameters  ErrorCode = 245
	CodeNoTribeAvailableToCoach   ErrorCode = 246
	CodeNotEnoughNectar           ErrorCode = 247
	CodeHeroItemNotPurchased      ErrorCode = 248
	CodeHeroItemAlreadyPurchased  ErrorCode = 249
	CodeAllHeroesPurchased        ErrorCode = 250
	CodeNotEnoughPotion           ErrorCode = 251
	CodeInvalidGiftCode           ErrorCode = 252
	CodeGiftCodeAlreadyRedeemed   ErrorCode = 253
	CodeGiftCodeExpired           ErrorCode = 254
	CodeHeroLevelRequirement      ErrorCode = 255
	CodeTribeEntryNotAllowed      ErrorCode = 256
	CodeLevelRequirement          ErrorCode = 257
	CodePrizeAlreadyReceived      ErrorCode = 258
	CodeInvalidMobileNumber       ErrorCode = 259
	CodeInvalidVerificationCode   ErrorCode = 260
	CodeNotSubscribed             ErrorCode = 261
	CodeNotCharged                ErrorCode = 262
	CodeMaxTribeBroadcastReached  ErrorCode = 264
	CodeBundlePurchaseError       ErrorCode = 265
	CodeUnknown                   ErrorCode = 0
	CodeTooManyRequests           ErrorCode = 429
)

type DomainError struct {
	Code    ErrorCode
	Message string
	Params  []any
	cause   error
}

func (e *DomainError) Error() string {
	if len(e.Params) > 0 {
		return fmt.Sprintf("[%d] %s", e.Code, fmt.Sprintf(e.Message, e.Params...))
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func (e *DomainError) Unwrap() error {
	return e.cause
}

func (e *DomainError) Is(target error) bool {
	t, ok := target.(*DomainError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

func (e *DomainError) WithParams(params ...any) *DomainError {
	return &DomainError{
		Code:    e.Code,
		Message: e.Message,
		Params:  params,
		cause:   e.cause,
	}
}

func (e *DomainError) Wrap(cause error) *DomainError {
	return &DomainError{
		Code:    e.Code,
		Message: e.Message,
		Params:  e.Params,
		cause:   cause,
	}
}

var (
	ErrGeneral              = &DomainError{Code: CodeGeneralError, Message: "An unexpected error occurred, please try again."}
	ErrAccountBlocked       = &DomainError{Code: CodeAccountBlocked, Message: "Your account is blocked."}
	ErrCardNotFound         = &DomainError{Code: CodeCardNotFound, Message: "Card not found."}
	ErrNoCardSelected       = &DomainError{Code: CodeNoCardSelected, Message: "You should choose at least one card!"}
	ErrCardAlreadyInAuction = &DomainError{Code: CodeCardAlreadyInAuction, Message: "Card is already in auction."}
	ErrNotEnoughCards       = &DomainError{Code: CodeNotEnoughCards, Message: "Not enough cards."}
	ErrInconsistency        = &DomainError{Code: CodeInconsistencyError, Message: "An inconsistency error occurred about your action, please try again."}
	ErrCardNotInAuction     = &DomainError{Code: CodeCardNotInAuction, Message: "Card not found in auctions."}
	ErrMaxBidReached        = &DomainError{Code: CodeMaxBidReached, Message: "You cannot bid on this item, it's already reached the maximum price."}
	ErrCannotBidOwnCard     = &DomainError{Code: CodeCannotBidOwnCard, Message: "You cannot bid on your own cards."}
	ErrAlreadyHighestBidder = &DomainError{Code: CodeAlreadyHighestBidder, Message: "You are already the highest bidder."}
	ErrAuctionClosed        = &DomainError{Code: CodeAuctionClosed, Message: "Too late! Auction is closed."}
	ErrCardTypeQuery        = &DomainError{Code: CodeCardTypeQuery, Message: "What kind of card should I look for?"}
	ErrCardTypeNotFound     = &DomainError{Code: CodeCardTypeNotFound, Message: "We cannot find the card type selected."}
	ErrCardTypeQueryDuplicate = &DomainError{Code: CodeCardTypeQueryDuplicate, Message: "What kind of card should I look for?"}
	ErrNotOwner             = &DomainError{Code: CodeNotOwner, Message: "You are not the owner. Cannot proceed."}
	ErrAccessDenied         = &DomainError{Code: CodeAccessDenied, Message: "Access denied."}
	ErrServerMaintenance    = &DomainError{Code: CodeServerMaintenance, Message: "Server is in maintenance. We will be available soon."}
	ErrNoCardsSpecified     = &DomainError{Code: CodeNoCardsSpecified, Message: "You need to specify some cards for this operation."}
	ErrNoOpponentSpecified  = &DomainError{Code: CodeNoOpponentSpecified, Message: "You need to specify an opponent for this operation."}
	ErrBattleNotAvailable   = &DomainError{Code: CodeBattleNotAvailable, Message: "You can't start the battle right now, please try again."}
	ErrCannotAttackSelf     = &DomainError{Code: CodeCannotAttackSelf, Message: "You cannot attack yourself."}
	ErrPlayerProtected      = &DomainError{Code: CodePlayerProtected, Message: "Player is protected by shield, wait a few hours before attacking them again."}
	ErrCaptchaRequired      = &DomainError{Code: CodeCaptchaRequired, Message: "You have to enter captcha code to proceed."}
	ErrPlayingOnAnotherDevice = &DomainError{Code: CodePlayingOnAnotherDevice, Message: "You are currently playing on another device, please close the game on other devices and try again in a few seconds."}
	ErrOpponentNotFound     = &DomainError{Code: CodeOpponentNotFound, Message: "Opponent not found."}
	ErrOpponentOutOfRange   = &DomainError{Code: CodeOpponentOutOfRange, Message: "Opponent not in your range."}
	ErrOpponentDefenseEmpty = &DomainError{Code: CodeOpponentDefenseEmpty, Message: "Opponent's defence deck is empty."}
	ErrInvalidCaptcha       = &DomainError{Code: CodeInvalidCaptcha, Message: "Captcha is invalid."}
	ErrCardInUse            = &DomainError{Code: CodeCardInUse, Message: "Card is used in one of your buildings and cannot be used here."}
	ErrCardMarkedForSacrifice = &DomainError{Code: CodeCardMarkedForSacrifice, Message: "Card is already marked for sacrifice."}
	ErrMaxPowerReached      = &DomainError{Code: CodeMaxPowerReached, Message: "Card has reached its maximum power."}
	ErrMaxEvolutionsExceeded = &DomainError{Code: CodeMaxEvolutionsExceeded, Message: "Cannot evolve more than two cards."}
	ErrEvolveLimitExceeded  = &DomainError{Code: CodeEvolveLimitExceeded, Message: "Only two cards of the same type can be evolved."}
	ErrBuildingNotFound     = &DomainError{Code: CodeBuildingNotFound, Message: "The specified building was not found."}
	ErrMaxCardsAssigned     = &DomainError{Code: CodeMaxCardsAssigned, Message: "Maximum number of cards assigned."}
	ErrLiveBattleUnavailable = &DomainError{Code: CodeLiveBattleUnavailable, Message: "Live-battle is not available due to some inconsistencies. We will be available soon."}
	ErrOpponentOffline      = &DomainError{Code: CodeOpponentOffline, Message: "Opponent is no longer online."}
	ErrAlreadyInLiveBattle  = &DomainError{Code: CodeAlreadyInLiveBattle, Message: "Already in live-battle."}
	ErrOpponentBusy         = &DomainError{Code: CodeOpponentBusy, Message: "Opponent is busy."}
	ErrCannotAttackTribeMate = &DomainError{Code: CodeCannotAttackTribeMate, Message: "You cannot attack your tribemate."}
	ErrBattleIDRequired     = &DomainError{Code: CodeBattleIDRequired, Message: "Battle ID required."}
	ErrInconsistencyUpdate  = &DomainError{Code: CodeInconsistencyErrorUpdate, Message: "An inconsistency error occurred. Please update your game to the latest version and try again."}
	ErrOperationTimeout     = &DomainError{Code: CodeOperationTimeout, Message: "The operation is timed out. Please try again."}
	ErrPlayerUnavailable    = &DomainError{Code: CodePlayerUnavailable, Message: "Player is not available right now. Please try again in a moment."}
	ErrBattleNotFound       = &DomainError{Code: CodeBattleNotFound, Message: "Battle not found. Please try again."}
	ErrAttackerNotFound     = &DomainError{Code: CodeAttackerNotFound, Message: "Attacker not found. Please try again."}
	ErrDefenderNotFound     = &DomainError{Code: CodeDefenderNotFound, Message: "Defender not found. Please try again."}
	ErrTribeNotFound        = &DomainError{Code: CodeTribeNotFound, Message: "Tribe not found. Please try again."}
	ErrNotInTribe           = &DomainError{Code: CodeNotInTribe, Message: "You are not in a tribe."}
	ErrTribeHelpLimit       = &DomainError{Code: CodeTribeHelpLimit, Message: "Your tribe cannot use help for not conforming to the tribe member limitation rule."}
	ErrMessageSizeExceeded  = &DomainError{Code: CodeMessageSizeExceeded, Message: "Message size exceeded."}
	ErrMessageNotFound      = &DomainError{Code: CodeMessageNotFound, Message: "Message not found."}
	ErrInvalidInvitationTicket = &DomainError{Code: CodeInvalidInvitationTicket, Message: "Invalid invitation ticket."}
	ErrInvalidRestoreKey    = &DomainError{Code: CodeInvalidRestoreKey, Message: "You have entered an invalid restore key. Please contact customer care for more information."}
	ErrAccountChangeLimit   = &DomainError{Code: CodeAccountChangeLimit, Message: "You cannot change your account more than once a day."}
	ErrIdentificationRequired = &DomainError{Code: CodeIdentificationRequired, Message: "We need your name or restore key for your identification."}
	ErrPlayerNotFound       = &DomainError{Code: CodePlayerNotFound, Message: "Player not found, please try again. If problem persists, be sure to let us know."}
	ErrInvitationCodeAlreadyRedeemed = &DomainError{Code: CodeInvitationCodeAlreadyRedeemed, Message: "You have already redeemed an invitation code."}
	ErrInvalidEmail         = &DomainError{Code: CodeInvalidEmail, Message: "Please enter a valid e-Mail address."}
	ErrEmailAlreadyExists   = &DomainError{Code: CodeEmailAlreadyExists, Message: "e-Mail already exists."}
	ErrEmailAlreadyRegistered = &DomainError{Code: CodeEmailAlreadyRegistered, Message: "You have already registered a valid e-Mail address."}
	ErrRewardAlreadyReceived = &DomainError{Code: CodeRewardAlreadyReceived, Message: "You have already received this reward."}
	ErrAccountDeactivated   = &DomainError{Code: CodeAccountDeactivated, Message: "Your account is deactivated."}
	ErrAccountResetFailed   = &DomainError{Code: CodeAccountResetFailed, Message: "Account reset failed."}
	ErrInvalidActivationCode = &DomainError{Code: CodeInvalidActivationCode, Message: "Your activation code is invalid. Please contact customer care for more information."}
	ErrCashRewardNotFound   = &DomainError{Code: CodeCashRewardNotFound, Message: "The specified cash reward cannot be found."}
	ErrAvatarNotAvailable   = &DomainError{Code: CodeAvatarNotAvailable, Message: "Avatar is not available. Please try another one."}
	ErrNotInLeague          = &DomainError{Code: CodeNotInLeague, Message: "You are not in a league."}
	ErrDeviceVerification   = &DomainError{Code: CodeDeviceVerificationError, Message: "We are having trouble to verify your device. Please contact customer care for more information."}
	ErrStoreNotSupported    = &DomainError{Code: CodeStoreNotSupported, Message: "The selected store is not supported yet."}
	ErrPackInconsistency    = &DomainError{Code: CodePackInconsistencyError, Message: "An inconsistency occurred about your selected pack. Please restart your game and try again."}
	ErrPaymentReceiptNotProvided = &DomainError{Code: CodePaymentReceiptNotProvided, Message: "The receipt for your payment is not provided. If the problem persists, please contact customer care."}
	ErrCardCannotEvolve     = &DomainError{Code: CodeCardCannotEvolve, Message: "You cannot evolve this card further."}
	ErrInternal             = &DomainError{Code: CodeInternalError, Message: "An internal error occurred. We are trying to fix it ASAP."}
	ErrInternalWait         = &DomainError{Code: CodeInternalErrorWait, Message: "An internal error occurred. Please wait for a moment and try again."}
	ErrPlayerNotParticipated = &DomainError{Code: CodePlayerNotParticipated, Message: "Player did not participate in this battle."}
	ErrCardCoolingDown      = &DomainError{Code: CodeCardCoolingDown, Message: "Card is cooling down and unavailable right now."}
	ErrCardCoolEnough       = &DomainError{Code: CodeCardCoolEnough, Message: "Card is cool enough."}
	ErrDeviceInconsistency  = &DomainError{Code: CodeDeviceInconsistency, Message: "There is an inconsistency about your device. Please contact customer care for more information."}
	ErrNameTooLong          = &DomainError{Code: CodeNameTooLong, Message: "Name is too long. You can specify %s characters for your name."}
	ErrNameAlreadyTaken     = &DomainError{Code: CodeNameAlreadyTaken, Message: "Name is already taken."}
	ErrCountryCodeNotSupported = &DomainError{Code: CodeCountryCodeNotSupported, Message: "Your selected country code is not supported yet."}
	ErrNotEnoughGold        = &DomainError{Code: CodeNotEnoughGold, Message: "You need %s more gold."}
	ErrOnlineOnAnotherDevice = &DomainError{Code: CodeOnlineOnAnotherDevice, Message: "You are currently online on another device."}
	ErrTribeNameTooLong     = &DomainError{Code: CodeTribeNameTooLong, Message: "Tribe name cannot exceed 30 characters."}
	ErrTopTribeEditLimit    = &DomainError{Code: CodeTopTribeEditLimit, Message: "Top 25 tribes cannot edit their information."}
	ErrCannotSellCard       = &DomainError{Code: CodeCannotSellCard, Message: "You cannot sell crystal or hero cards."}
	ErrStatusRequired       = &DomainError{Code: CodeStatusRequired, Message: "You need to specify a status!"}
	ErrNameRequired         = &DomainError{Code: CodeNameRequired, Message: "You should write a name"}
	ErrDescriptionRequired  = &DomainError{Code: CodeDescriptionRequired, Message: "Description Required."}
	ErrDescriptionTooLong   = &DomainError{Code: CodeDescriptionTooLong, Message: "Description is Too Long."}
	ErrTribeNameExists      = &DomainError{Code: CodeTribeNameExists, Message: "A tribe with the same name already exists."}
	ErrInvalidChiefPlayers  = &DomainError{Code: CodeInvalidChiefPlayers, Message: "Invalid chief players for tribe id %d"}
	ErrAlreadyInTribe       = &DomainError{Code: CodeAlreadyInTribe, Message: "You are already a member of this tribe."}
	ErrNoTribeBuilding      = &DomainError{Code: CodeNoTribeBuilding, Message: "You do not have the tribe building yet."}
	ErrMaxTribeChangeLimit  = &DomainError{Code: CodeMaxTribeChangeLimit, Message: "You have reached the maximum tribe change limitation."}
	ErrTribeNoMembers       = &DomainError{Code: CodeTribeNoMembers, Message: "Inconsistency exception! Tribe has no members."}
	ErrUndecidedRequest     = &DomainError{Code: CodeUndecidedRequest, Message: "You already have an undecided request."}
	ErrInvalidDecisionParam = &DomainError{Code: CodeInvalidDecisionParameter, Message: "Invalid decision parameter. Please try again."}
	ErrNotJoinRequest       = &DomainError{Code: CodeNotJoinRequest, Message: "This is not a join request. Please contact customer care for more information."}
	ErrInconsistentData     = &DomainError{Code: CodeInconsistentDataProvided, Message: "Inconsistent Data Provided, please try again. If problem persists, be sure to let us know."}
	ErrJoinRequestProcessed = &DomainError{Code: CodeJoinRequestProcessed, Message: "Join request has already been processed."}
	ErrTribeFull            = &DomainError{Code: CodeTribeFull, Message: "Tribe is full."}
	ErrTribeAccessDenied    = &DomainError{Code: CodeTribeAccessDenied, Message: "Tribe access permission denied."}
	ErrPlayerAlreadyInTribe = &DomainError{Code: CodePlayerAlreadyInTribe, Message: "Player is already a member of this tribe."}
	ErrNotAnInvitation      = &DomainError{Code: CodeNotAnInvitation, Message: "This is not an invitation."}
	ErrInvitationProcessed  = &DomainError{Code: CodeInvitationProcessed, Message: "Invitation has already been processed."}
	ErrNoTribesAvailable    = &DomainError{Code: CodeNoTribesAvailable, Message: "You do not have any tribes."}
	ErrPlayerNotInYourTribe = &DomainError{Code: CodePlayerNotInYourTribe, Message: "Player is not in your tribe."}
	ErrPlayerAlreadyPromoted = &DomainError{Code: CodePlayerAlreadyPromoted, Message: "Player is already promoted."}
	ErrPlayerNotElder       = &DomainError{Code: CodePlayerNotElder, Message: "Player is not an elder."}
	ErrSelfPoke             = &DomainError{Code: CodeSelfPoke, Message: "You can not poke yourself."}
	ErrNotAMemberOfTribe    = &DomainError{Code: CodeNotAMemberOfTribe, Message: "You are not a member of this tribe."}
	ErrSelfKick             = &DomainError{Code: CodeSelfKick, Message: "You can not kick yourself."}
	ErrInsufficientPermission = &DomainError{Code: CodeInsufficientPermission, Message: "You don't have enough permission to do this operation."}
	ErrMaxLevelCooldown     = &DomainError{Code: CodeMaxLevelCooldownBuilding, Message: "Maximum level reached for Cooldown building."}
	ErrMaxLevelMainHall     = &DomainError{Code: CodeMaxLevelMainHallBuilding, Message: "Maximum level reached for Main Hall building."}
	ErrMaxLevelDefense      = &DomainError{Code: CodeMaxLevelDefenseBuilding, Message: "Maximum level reached for Defense building."}
	ErrMaxLevelOffense      = &DomainError{Code: CodeMaxLevelOffenseBuilding, Message: "Maximum level reached for Offense building."}
	ErrMaxLevelGold         = &DomainError{Code: CodeMaxLevelGoldBuilding, Message: "Maximum level reached for Gold building."}
	ErrMaxLevelBank         = &DomainError{Code: CodeMaxLevelBankBuilding, Message: "Maximum level reached for Bank building."}
	ErrInvalidBuildingType  = &DomainError{Code: CodeInvalidBuildingTypeForUpgrade, Message: "Invalid building type for upgrade."}
	ErrMinimumDonation      = &DomainError{Code: CodeMinimumDonationAmount, Message: "Minimum donation amount is %s golds."}
	ErrPlayerNoTribeBuilding = &DomainError{Code: CodePlayerNoTribeBuilding, Message: "Player does not have the tribe building yet."}
	ErrUndecidedInvitation  = &DomainError{Code: CodeUndecidedInvitation, Message: "User has an undecided invitation already."}
	ErrNotEnoughTribeMoney  = &DomainError{Code: CodeNotEnoughTribeMoney, Message: "Not enough tribe money."}
	ErrTribeScoreUpdate     = &DomainError{Code: CodeTribeScoreUpdateFailed, Message: "Failed to update tribe score, please try again. If problem persists, be sure to let us know."}
	ErrTribeNoIdentifier    = &DomainError{Code: CodeTribeNoIdentifier, Message: "Tribe has no identifier, please try again. If problem persists, be sure to let us know."}
	ErrInvalidBuildingCardCapacity = &DomainError{Code: CodeInvalidBuildingTypeForCardCapacity, Message: "Invalid building type for card capacity."}
	ErrOperationFailed      = &DomainError{Code: CodeOperationFailed, Message: "Operation failed, please try again. If problem persists, be sure to let us know."}
	ErrGooglePlayVerification = &DomainError{Code: CodeGooglePlayVerificationFailed, Message: "Google Play verification failed."}
	ErrInvalidPurchaseState = &DomainError{Code: CodeInvalidPurchaseState, Message: "purchaseState is invalid. (%s)"}
	ErrDataReading          = &DomainError{Code: CodeDataReadingError, Message: "Problem reading data from server"}
	ErrSibcheVerification   = &DomainError{Code: CodeSibcheVerificationFailed, Message: "Sibche verification failed."}
	ErrMaxBoostsLimit       = &DomainError{Code: CodeMaxBoostsLimit, Message: "You cannot buy more boosts."}
	ErrInvalidCountryCode   = &DomainError{Code: CodeInvalidCountryCode, Message: "Invalid country code."}
	ErrUserNotFound         = &DomainError{Code: CodeUserNotFound, Message: "User Not Found"}
	ErrUserRecentlyPoked    = &DomainError{Code: CodeUserRecentlyPoked, Message: "User has been recently poked."}
	ErrLeagueUpdateInProgress = &DomainError{Code: CodeLeagueUpdateInProgress, Message: "Updating league is in progress. Please wait."}
	ErrFeatureNotImplemented = &DomainError{Code: CodeFeatureNotImplemented, Message: "This feature is not implemented yet :)"}
	ErrInvalidLeagueID      = &DomainError{Code: CodeInvalidLeagueID, Message: "Operation failed (invalid league ID). Please contact customer care for more information."}
	ErrTribeHelpInProgress  = &DomainError{Code: CodeTribeHelpAlreadyInProgress, Message: "You should be fast! Your tribe mates are already helping"}
	ErrTutorialUpdateParams = &DomainError{Code: CodeTutorialUpdateParameters, Message: "Tutorial updating requires more parameters"}
	ErrNoTribeAvailableToCoach = &DomainError{Code: CodeNoTribeAvailableToCoach, Message: "No tribe available to coach, Try later"}
	ErrNotEnoughNectar      = &DomainError{Code: CodeNotEnoughNectar, Message: "You need %s more nectar."}
	ErrHeroItemNotPurchased = &DomainError{Code: CodeHeroItemNotPurchased, Message: "You've not bought this hero item."}
	ErrHeroItemAlreadyPurchased = &DomainError{Code: CodeHeroItemAlreadyPurchased, Message: "You've already bought this hero item."}
	ErrAllHeroesPurchased   = &DomainError{Code: CodeAllHeroesPurchased, Message: "You've bought all of the heroes!"}
	ErrNotEnoughPotion      = &DomainError{Code: CodeNotEnoughPotion, Message: "You have not enough potion."}
	ErrInvalidGiftCode      = &DomainError{Code: CodeInvalidGiftCode, Message: "Wrong gift code entered."}
	ErrGiftCodeAlreadyRedeemed = &DomainError{Code: CodeGiftCodeAlreadyRedeemed, Message: "You've already redeemed this gift code"}
	ErrGiftCodeExpired      = &DomainError{Code: CodeGiftCodeExpired, Message: "Gift code is expired."}
	ErrHeroLevelRequirement = &DomainError{Code: CodeHeroLevelRequirement, Message: "You need at least one hero with level %s or above."}
	ErrTribeEntryNotAllowed = &DomainError{Code: CodeTribeEntryNotAllowed, Message: "You can't enter this tribe. Try another one."}
	ErrLevelRequirement     = &DomainError{Code: CodeLevelRequirement, Message: "You should reach at least level %s !"}
	ErrPrizeAlreadyReceived = &DomainError{Code: CodePrizeAlreadyReceived, Message: "You have got enough prize! Wait a little"}
	ErrInvalidMobileNumber  = &DomainError{Code: CodeInvalidMobileNumber, Message: "You should enter a valid mobile number!"}
	ErrInvalidVerificationCode = &DomainError{Code: CodeInvalidVerificationCode, Message: "Not a valid verification code! Try again"}
	ErrNotSubscribed        = &DomainError{Code: CodeNotSubscribed, Message: "You are not subscribed! First subscribe then try again"}
	ErrNotCharged           = &DomainError{Code: CodeNotCharged, Message: "You are not charged! First get charged then try again"}
	ErrMaxTribeBroadcast    = &DomainError{Code: CodeMaxTribeBroadcastReached, Message: "You reached max tribe broadcast! Wait a while and try again"}
	ErrBundlePurchase       = &DomainError{Code: CodeBundlePurchaseError, Message: "Error occurred in purchasing bundle. Please contact customer care."}
	ErrUnknown              = &DomainError{Code: CodeUnknown, Message: "An unknown error occurred."}
	ErrTooManyRequests      = &DomainError{Code: CodeTooManyRequests, Message: "You have sent too many requests in a given amount of time."}
)

func (c ErrorCode) String() string {
	if name, ok := errorCodeNames[c]; ok {
		return name
	}
	return fmt.Sprintf("ErrorCode(%d)", c)
}

var errorCodeNames = map[ErrorCode]string{
	CodeGeneralError:              "GENERAL_ERROR",
	CodeAccountBlocked:            "ACCOUNT_BLOCKED",
	CodeCardNotFound:              "CARD_NOT_FOUND",
	CodeAccessDenied:              "ACCESS_DENIED",
	CodeNotEnoughGold:             "NOT_ENOUGH_GOLD",
	CodeTooManyRequests:           "TOO_MANY_REQUESTS",
	// Add more as needed...
}

var errorByCode sync.Map

func init() {
	errors := []*DomainError{
		ErrGeneral, ErrAccountBlocked, ErrCardNotFound, ErrNoCardSelected,
		ErrCardAlreadyInAuction, ErrNotEnoughCards, ErrInconsistency,
		ErrCardNotInAuction, ErrMaxBidReached, ErrCannotBidOwnCard,
		ErrAlreadyHighestBidder, ErrAuctionClosed, ErrCardTypeQuery,
		ErrCardTypeNotFound, ErrCardTypeQueryDuplicate, ErrNotOwner,
		ErrAccessDenied, ErrServerMaintenance, ErrNoCardsSpecified,
		ErrNoOpponentSpecified, ErrBattleNotAvailable, ErrCannotAttackSelf,
		ErrPlayerProtected, ErrCaptchaRequired, ErrPlayingOnAnotherDevice,
		ErrOpponentNotFound, ErrOpponentOutOfRange, ErrOpponentDefenseEmpty,
		ErrInvalidCaptcha, ErrCardInUse, ErrCardMarkedForSacrifice,
		ErrMaxPowerReached, ErrMaxEvolutionsExceeded, ErrEvolveLimitExceeded,
		ErrBuildingNotFound, ErrMaxCardsAssigned, ErrLiveBattleUnavailable,
		ErrOpponentOffline, ErrAlreadyInLiveBattle, ErrOpponentBusy,
		ErrCannotAttackTribeMate, ErrBattleIDRequired, ErrInconsistencyUpdate,
		ErrOperationTimeout, ErrPlayerUnavailable, ErrBattleNotFound,
		ErrAttackerNotFound, ErrDefenderNotFound, ErrTribeNotFound,
		ErrNotInTribe, ErrTribeHelpLimit, ErrMessageSizeExceeded,
		ErrMessageNotFound, ErrInvalidInvitationTicket, ErrInvalidRestoreKey,
		ErrAccountChangeLimit, ErrIdentificationRequired, ErrPlayerNotFound,
		ErrInvitationCodeAlreadyRedeemed, ErrInvalidEmail, ErrEmailAlreadyExists,
		ErrEmailAlreadyRegistered, ErrRewardAlreadyReceived, ErrAccountDeactivated,
		ErrAccountResetFailed, ErrInvalidActivationCode, ErrCashRewardNotFound,
		ErrAvatarNotAvailable, ErrNotInLeague, ErrDeviceVerification,
		ErrStoreNotSupported, ErrPackInconsistency, ErrPaymentReceiptNotProvided,
		ErrCardCannotEvolve, ErrInternal, ErrInternalWait,
		ErrPlayerNotParticipated, ErrCardCoolingDown, ErrCardCoolEnough,
		ErrDeviceInconsistency, ErrNameTooLong, ErrNameAlreadyTaken,
		ErrCountryCodeNotSupported, ErrNotEnoughGold, ErrOnlineOnAnotherDevice,
		ErrTribeNameTooLong, ErrTopTribeEditLimit, ErrCannotSellCard,
		ErrStatusRequired, ErrNameRequired, ErrDescriptionRequired,
		ErrDescriptionTooLong, ErrTribeNameExists, ErrInvalidChiefPlayers,
		ErrAlreadyInTribe, ErrNoTribeBuilding, ErrMaxTribeChangeLimit,
		ErrTribeNoMembers, ErrUndecidedRequest, ErrInvalidDecisionParam,
		ErrNotJoinRequest, ErrInconsistentData, ErrJoinRequestProcessed,
		ErrTribeFull, ErrTribeAccessDenied, ErrPlayerAlreadyInTribe,
		ErrNotAnInvitation, ErrInvitationProcessed, ErrNoTribesAvailable,
		ErrPlayerNotInYourTribe, ErrPlayerAlreadyPromoted, ErrPlayerNotElder,
		ErrSelfPoke, ErrNotAMemberOfTribe, ErrSelfKick,
		ErrInsufficientPermission, ErrMaxLevelCooldown, ErrMaxLevelMainHall,
		ErrMaxLevelDefense, ErrMaxLevelOffense, ErrMaxLevelGold,
		ErrMaxLevelBank, ErrInvalidBuildingType, ErrMinimumDonation,
		ErrPlayerNoTribeBuilding, ErrUndecidedInvitation, ErrNotEnoughTribeMoney,
		ErrTribeScoreUpdate, ErrTribeNoIdentifier, ErrInvalidBuildingCardCapacity,
		ErrOperationFailed, ErrGooglePlayVerification, ErrInvalidPurchaseState,
		ErrDataReading, ErrSibcheVerification, ErrMaxBoostsLimit,
		ErrInvalidCountryCode, ErrUserNotFound, ErrUserRecentlyPoked,
		ErrLeagueUpdateInProgress, ErrFeatureNotImplemented, ErrInvalidLeagueID,
		ErrTribeHelpInProgress, ErrTutorialUpdateParams,
		ErrNoTribeAvailableToCoach, ErrNotEnoughNectar, ErrHeroItemNotPurchased,
		ErrHeroItemAlreadyPurchased, ErrAllHeroesPurchased, ErrNotEnoughPotion,
		ErrInvalidGiftCode, ErrGiftCodeAlreadyRedeemed, ErrGiftCodeExpired,
		ErrHeroLevelRequirement, ErrTribeEntryNotAllowed, ErrLevelRequirement,
		ErrPrizeAlreadyReceived, ErrInvalidMobileNumber, ErrInvalidVerificationCode,
		ErrNotSubscribed, ErrNotCharged, ErrMaxTribeBroadcast,
		ErrBundlePurchase, ErrUnknown, ErrTooManyRequests,
	}
	for _, err := range errors {
		errorByCode.Store(err.Code, err)
	}
}

func New(code ErrorCode, params ...any) *DomainError {
	if base, ok := errorByCode.Load(code); ok {
		err := base.(*DomainError)
		if len(params) > 0 {
			return err.WithParams(params...)
		}
		return err
	}
	return &DomainError{Code: code, Message: "Unknown error", Params: params}
}

func Wrap(code ErrorCode, cause error, params ...any) *DomainError {
	err := New(code, params...)
	return err.Wrap(cause)
}

func IsCode(err error, code ErrorCode) bool {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr.Code == code
	}
	return false
}

func GetCode(err error) ErrorCode {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr.Code
	}
	return CodeUnknown
}

func Is(err error, target *DomainError) bool {
	return errors.Is(err, target)
}