package notifier

import (
	"fmt"
	"math"
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

	h := int(math.Round(j.Health * 100))
	if err := e.bot.Send(fmt.Sprintf("%s hull damage detected, integrity is %d%%", prefix, h)); err != nil {
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
