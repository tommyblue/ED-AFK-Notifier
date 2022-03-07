package notifier

import (
	"fmt"
	"log"
	"math"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type eventFn func(*Notifier, journalEvent, bool) error

func (e *Notifier) notify(msg string, skipNotify bool) error {
	if skipNotify {
		return nil
	}

	if err := e.bot.Send(msg); err != nil {
		return fmt.Errorf("error sending message: %v", err)
	}

	return nil
}

func hullDamageEvent(e *Notifier, j journalEvent, skipNotify bool) error {
	if j.Fighter && !e.cfg.FighterNotifs {
		return nil
	}

	prefix := "Ship"
	if j.Fighter {
		prefix = "Fighter"
	}

	h := int(math.Round(j.Health * 100))
	return e.notify(fmt.Sprintf("%s hull damage detected, integrity is %d%%", prefix, h), skipNotify)

}

func diedEvent(e *Notifier, j journalEvent, skipNotify bool) error {
	return e.notify("Your ship has been destroyed", skipNotify)
}

func shieldStateEvent(e *Notifier, j journalEvent, skipNotify bool) error {
	var msg string
	if j.ShieldsUp {
		msg = "Shields are up again"
	} else {
		msg = "Shields are down!"
	}

	return e.notify(msg, skipNotify)
}

func bountyEvent(e *Notifier, j journalEvent, skipNotify bool) error {
	e.totalPiratesReward += j.TotalPiratesReward
	e.killedPirates++

	log.Println("Pirates killed:", e.killedPirates)

	p := message.NewPrinter(language.Make("en"))
	bounties := p.Sprintf("%d", e.totalPiratesReward)
	log.Println("Total bounty rewards:", bounties)

	if !e.cfg.KillsNotifs {
		return nil
	}

	if !e.cfg.KillsSilentNotifs || e.killedPirates%10 == 0 {
		return e.notify(fmt.Sprintf("Total rewards: %s credits\nPirates killed: %d", bounties, e.killedPirates), skipNotify)
	}

	return nil
}

func missionAcceptedEvent(e *Notifier, j journalEvent, skipNotify bool) error {
	e.activeMissions++
	e.loggedMissions[j.MissionID] = false

	log.Println("Active missions:", e.activeMissions)

	return nil
}

func missionRedirectedEvent(e *Notifier, j journalEvent, skipNotify bool) error {
	if e.loggedMissions[j.MissionID] {
		return nil
	}

	e.activeMissions--
	e.loggedMissions[j.MissionID] = true

	log.Println("Active missions:", e.activeMissions)

	if e.activeMissions == 0 {
		return e.notify("No more active missions, go collect new ones!", skipNotify)
	}

	return nil
}

func missionCompletedEvent(e *Notifier, j journalEvent, skipNotify bool) error {
	if e.loggedMissions[j.MissionID] {
		return nil
	}

	e.activeMissions--
	delete(e.loggedMissions, j.MissionID)

	e.totalMissionsReward += j.MissionReward
	log.Println("Obtained reward for missions until now:", e.totalMissionsReward)

	log.Println("Active missions:", e.activeMissions)

	if e.activeMissions == 0 {
		return e.notify("No more active missions, go collect new ones!", skipNotify)
	}

	return nil
}

func missionAbandonedEvent(e *Notifier, j journalEvent, skipNotify bool) error {
	e.activeMissions--
	delete(e.loggedMissions, j.MissionID)

	return nil
}

func missionsInitEvent(e *Notifier, j journalEvent, skipNotify bool) error {
	log.Println("Found missions log message, starting new initialization")
	e.initNotifier()

	return nil
}
