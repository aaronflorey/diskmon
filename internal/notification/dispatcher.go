package notification

import (
	"context"
	"errors"
	"fmt"
	"time"

	"diskmon/internal/health"
)

const defaultSendTimeout = 5 * time.Second

type Formatter func(driveID string, result health.Result) (subject, body string)

type Dispatcher struct {
	targets        []dispatchTarget
	defaultTimeout time.Duration
	formatter      Formatter
}

type dispatchTarget struct {
	entry  Entry
	sender Sender
}

func NewDispatcher(entries []Entry, factory SenderFactory, defaultTimeout time.Duration) (*Dispatcher, error) {
	if factory == nil {
		factory = NewSenderFactory()
	}
	if defaultTimeout <= 0 {
		defaultTimeout = defaultSendTimeout
	}

	targets := make([]dispatchTarget, 0, len(entries))
	for _, entry := range entries {
		sender, err := factory.Build(entry)
		if err != nil {
			return nil, fmt.Errorf("build notifier %q: %w", entry.Name, err)
		}
		targets = append(targets, dispatchTarget{entry: entry, sender: sender})
	}

	return &Dispatcher{
		targets:        targets,
		defaultTimeout: defaultTimeout,
		formatter:      FormatMessage,
	}, nil
}

func (d *Dispatcher) DispatchIfNeeded(ctx context.Context, req DispatchRequest) (DispatchResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	subject, body := d.formatter(req.DriveID, req.Current)
	result := DispatchResult{
		Subject:  subject,
		Body:     body,
		Outcomes: make([]EntryOutcome, 0, len(d.targets)),
	}

	var errs []error
	for _, target := range d.targets {
		outcome := decideOutcome(target.entry, req.PreviousStatus, req.Current.Status)
		if outcome.Attempted {
			timeout := target.entry.Timeout
			if timeout <= 0 {
				timeout = d.defaultTimeout
			}
			sendCtx, cancel := context.WithTimeout(ctx, timeout)
			err := target.sender.Send(sendCtx, subject, body)
			cancel()
			if err != nil {
				outcome.Err = err
				errs = append(errs, fmt.Errorf("send notification %q: %w", target.entry.Name, err))
			} else {
				outcome.Sent = true
			}
		}
		result.Outcomes = append(result.Outcomes, outcome)
	}

	return result, errors.Join(errs...)
}

func decideOutcome(entry Entry, previous *health.Status, current health.Status) EntryOutcome {
	outcome := EntryOutcome{Name: entry.Name}

	if !entry.Enabled {
		outcome.Reason = ReasonDisabled
		return outcome
	}

	isFail := current != health.StatusGreen
	if previous == nil {
		if isFail {
			if !entry.OnFail {
				outcome.Reason = ReasonFailToggleDisabled
				return outcome
			}
			outcome.Reason = ReasonInitialFail
			outcome.Attempted = true
			return outcome
		}
		outcome.Reason = ReasonInitialPassSuppressed
		return outcome
	}

	if *previous == current {
		outcome.Reason = ReasonUnchanged
		return outcome
	}

	if isFail {
		if !entry.OnFail {
			outcome.Reason = ReasonFailToggleDisabled
			return outcome
		}
		outcome.Reason = ReasonTransitionToFail
		outcome.Attempted = true
		return outcome
	}

	if !entry.OnPass {
		outcome.Reason = ReasonPassToggleDisabled
		return outcome
	}

	outcome.Reason = ReasonTransitionToPass
	outcome.Attempted = true
	return outcome
}
