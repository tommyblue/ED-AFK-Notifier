package notifier

import (
	"fmt"
)

type eventFn func(*Notifier, journalEvent) error

func hullDamageEvent(e *Notifier, j journalEvent) error {
	if j.Fighter && !e.cfg.FighterNotifs {
		return nil
	}

	prefix := "Ship"
	if j.Fighter {
		prefix = "Fighter"
	}

	if err := e.bot.Send(fmt.Sprintf("%s hull damage detected, integrity is %2f", prefix, j.Health)); err != nil {
		return fmt.Errorf("error sending message: %v", err)
	}

	return nil
}

func diedEvent(e *Notifier, j journalEvent) error {
	if err := e.bot.Send("Your ship has been destroyed"); err != nil {
		return fmt.Errorf("error sending message: %v", err)
	}

	return nil
}
