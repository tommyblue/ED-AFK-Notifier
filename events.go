package notifier

import (
	"fmt"
	"log"
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
	e.totalPiratesReward += j.TotalPiratesReward
	e.killedPirates++

	log.Println("Pirates killed:", e.killedPirates)
	log.Println("Total bounty rewards:", e.totalPiratesReward)

	if !e.cfg.KillsNotifs {
		return nil
	}

	if !e.cfg.KillsSilentNotifs || e.killedPirates%10 == 0 {
		if err := e.bot.Send(fmt.Sprintf("Total rewards: %d credits\nPirates killed: %d", e.totalPiratesReward, e.killedPirates)); err != nil {
			return fmt.Errorf("error sending message: %v", err)
		}
	}

	return nil
}

func missionAcceptedEvent(e *Notifier, j journalEvent) error {
	e.activeMissions++
	e.loggedMissions[j.MissionID] = false

	log.Println("Active missions:", e.activeMissions)

	return nil
}

func missionRedirectedEvent(e *Notifier, j journalEvent) error {
	if e.loggedMissions[j.MissionID] {
		return nil
	}

	e.activeMissions--
	e.loggedMissions[j.MissionID] = true

	log.Println("Active missions:", e.activeMissions)

	if e.activeMissions == 0 {
		if err := e.bot.Send("No more active missions, go collect new ones!"); err != nil {
			return fmt.Errorf("error sending message: %v", err)
		}
	}

	return nil
}

func missionCompletedEvent(e *Notifier, j journalEvent) error {
	if e.loggedMissions[j.MissionID] {
		return nil
	}

	e.activeMissions--
	delete(e.loggedMissions, j.MissionID)

	e.totalMissionsReward += j.MissionReward
	log.Println("Obtained reward for missions until now:", e.totalMissionsReward)

	log.Println("Active missions:", e.activeMissions)

	if e.activeMissions == 0 {
		if err := e.bot.Send("No more active missions, go collect new ones!"); err != nil {
			return fmt.Errorf("error sending message: %v", err)
		}
	}

	return nil
}

func missionAbandonedEvent(e *Notifier, j journalEvent) error {
	e.activeMissions--
	delete(e.loggedMissions, j.MissionID)

	return nil
}
