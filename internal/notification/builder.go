package notification

import (
	"context"
	"errors"
	"fmt"
	stdhttp "net/http"
	"strings"

	"github.com/nikoksr/notify"
	notifydiscord "github.com/nikoksr/notify/service/discord"
	notifyhttp "github.com/nikoksr/notify/service/http"
	notifyslack "github.com/nikoksr/notify/service/slack"
)

type Sender interface {
	Send(ctx context.Context, subject, body string) error
}

type SenderFactory interface {
	Build(entry Entry) (Sender, error)
}

type defaultSenderFactory struct{}

func NewSenderFactory() SenderFactory {
	return defaultSenderFactory{}
}

func (defaultSenderFactory) Build(entry Entry) (Sender, error) {
	providerType := strings.ToLower(strings.TrimSpace(entry.Provider.Type))

	switch providerType {
	case ProviderHTTP:
		return buildHTTP(entry)
	case ProviderSlack:
		return buildSlack(entry)
	case ProviderDiscord:
		return buildDiscord(entry)
	default:
		return nil, fmt.Errorf("unsupported notification provider type %q", entry.Provider.Type)
	}
}

func buildHTTP(entry Entry) (Sender, error) {
	url := strings.TrimSpace(entry.Provider.HTTP.URL)
	if url == "" {
		return nil, errors.New("http url is required")
	}

	svc := notifyhttp.New()
	svc.AddReceiversURLs(url)

	return notify.NewWithServices(svc), nil
}

func buildSlack(entry Entry) (Sender, error) {
	cfg := entry.Provider.Slack
	mode := strings.ToLower(strings.TrimSpace(cfg.Mode))
	if mode == "" {
		mode = ModeSDK
	}

	switch mode {
	case ModeSDK:
		token := strings.TrimSpace(cfg.APIToken)
		if token == "" {
			return nil, errors.New("slack api token is required for sdk mode")
		}

		receivers := trimNonEmpty(cfg.ChannelIDs)
		if len(receivers) == 0 {
			return nil, errors.New("at least one slack channel id is required for sdk mode")
		}

		svc := notifyslack.New(token)
		svc.AddReceivers(receivers...)

		return notify.NewWithServices(svc), nil
	case ModeWebhook:
		url := strings.TrimSpace(cfg.WebhookURL)
		if url == "" {
			return nil, errors.New("slack webhook url is required for webhook mode")
		}

		svc := notifyhttp.New()
		svc.AddReceivers(&notifyhttp.Webhook{
			URL:         url,
			Method:      stdhttp.MethodPost,
			ContentType: "application/json; charset=utf-8",
			BuildPayload: func(subject, message string) any {
				return map[string]string{"text": subject + "\n" + message}
			},
		})

		return notify.NewWithServices(svc), nil
	default:
		return nil, fmt.Errorf("unsupported slack mode %q", cfg.Mode)
	}
}

func buildDiscord(entry Entry) (Sender, error) {
	cfg := entry.Provider.Discord
	mode := strings.ToLower(strings.TrimSpace(cfg.Mode))
	if mode == "" {
		mode = ModeSDK
	}

	switch mode {
	case ModeSDK:
		receivers := trimNonEmpty(cfg.ChannelIDs)
		if len(receivers) == 0 {
			return nil, errors.New("at least one discord channel id is required for sdk mode")
		}

		svc := notifydiscord.New()
		botToken := strings.TrimSpace(cfg.BotToken)
		oauthToken := strings.TrimSpace(cfg.OAuthToken)
		switch {
		case botToken != "":
			if err := svc.AuthenticateWithBotToken(botToken); err != nil {
				return nil, fmt.Errorf("authenticate discord bot token: %w", err)
			}
		case oauthToken != "":
			if err := svc.AuthenticateWithOAuth2Token(oauthToken); err != nil {
				return nil, fmt.Errorf("authenticate discord oauth token: %w", err)
			}
		default:
			return nil, errors.New("discord bot token or oauth token is required for sdk mode")
		}
		svc.AddReceivers(receivers...)

		return notify.NewWithServices(svc), nil
	case ModeWebhook:
		url := strings.TrimSpace(cfg.WebhookURL)
		if url == "" {
			return nil, errors.New("discord webhook url is required for webhook mode")
		}

		svc := notifyhttp.New()
		svc.AddReceivers(&notifyhttp.Webhook{
			URL:         url,
			Method:      stdhttp.MethodPost,
			ContentType: "application/json; charset=utf-8",
			BuildPayload: func(subject, message string) any {
				return map[string]string{"content": subject + "\n" + message}
			},
		})

		return notify.NewWithServices(svc), nil
	default:
		return nil, fmt.Errorf("unsupported discord mode %q", cfg.Mode)
	}
}

func trimNonEmpty(values []string) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		trimmed := strings.TrimSpace(v)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
