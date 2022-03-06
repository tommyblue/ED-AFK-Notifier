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

func shieldStateEvent(e *Notifier, j journalEvent) error {
	var msg string
	if j.ShieldsUp {
		msg = "Shields are up again"
	} else {
		msg = "Shields are down!"
	}

	if err := e.bot.Send(msg); err != nil {
		return fmt.Errorf("error sending message: %v", err)
	}

	return nil
}

func bountyEvent(e *Notifier, j journalEvent) error {
	if !e.cfg.KillsNotifs {
		return nil
	}

	e.totalReward += j.TotalReward
	e.killedPirates++

	if !e.cfg.KillsSilentNotifs || e.killedPirates%10 == 0 {
		if err := e.bot.Send(fmt.Sprintf("Total rewards: %d credits\nPirates killed: %d", e.totalReward, e.killedPirates)); err != nil {
			return fmt.Errorf("error sending message: %v", err)
		}
	}

	return nil
}
