package notification

import (
	"time"

	"diskmon/internal/health"
)

const (
	ProviderHTTP    = "http"
	ProviderSlack   = "slack"
	ProviderDiscord = "discord"
)

const (
	ModeSDK     = "sdk"
	ModeWebhook = "webhook"
)

type DecisionReason string

const (
	ReasonDisabled              DecisionReason = "disabled"
	ReasonInitialPassSuppressed DecisionReason = "initial_pass_suppressed"
	ReasonInitialFail           DecisionReason = "initial_fail"
	ReasonUnchanged             DecisionReason = "unchanged"
	ReasonTransitionToPass      DecisionReason = "transition_to_pass"
	ReasonTransitionToFail      DecisionReason = "transition_to_fail"
	ReasonPassToggleDisabled    DecisionReason = "pass_toggle_disabled"
	ReasonFailToggleDisabled    DecisionReason = "fail_toggle_disabled"
)

type Entry struct {
	Name     string
	Enabled  bool
	OnPass   bool
	OnFail   bool
	Timeout  time.Duration
	Provider Provider
}

type Provider struct {
	Type    string
	HTTP    HTTPProvider
	Slack   SlackProvider
	Discord DiscordProvider
}

type HTTPProvider struct {
	URL string
}

type SlackProvider struct {
	Mode       string
	APIToken   string
	ChannelIDs []string
	WebhookURL string
}

type DiscordProvider struct {
	Mode       string
	BotToken   string
	OAuthToken string
	ChannelIDs []string
	WebhookURL string
}

type DispatchRequest struct {
	DriveID        string
	PreviousStatus *health.Status
	Current        health.Result
}

type EntryOutcome struct {
	Name      string
	Sent      bool
	Reason    DecisionReason
	Err       error
	Attempted bool
}

type DispatchResult struct {
	Subject  string
	Body     string
	Outcomes []EntryOutcome
}
