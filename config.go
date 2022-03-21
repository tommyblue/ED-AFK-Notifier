package notifier

import log "github.com/sirupsen/logrus"

type config struct {
	journalPath       string
	botToken          string
	botChannelId      int64
	logDebug          bool
	notifyShields     bool
	notifyFighter     bool
	notifyKills       bool
	notifySilentKills bool
}

// LogConfig prints the current configuration to the log.
func (e *Notifier) LogConfig() {
	log.Infof("Config:")
	log.Infof("  Notify fighter status: %t", e.cfg.notifyFighter)
	log.Infof("  Notify shields status: %t", e.cfg.notifyShields)
	log.Infof("  Notify on kills: %t (silent: %t)", e.cfg.notifyKills, e.cfg.notifySilentKills)
	log.Infof("  Journal file path: %s", e.cfg.journalPath)
}

func (e *Notifier) configString(key string) string {
	return e.gui.App.Preferences().String(key)
}

func (e *Notifier) configBool(key string) bool {
	return e.gui.App.Preferences().Bool(key)
}

func (e *Notifier) configInt64(key string) int64 {
	return int64(e.gui.App.Preferences().Int(key))
}

func (e *Notifier) syncConfig() {
	sync := func() {
		e.cfg.journalPath = e.configString(CONFIG_JOURNAL_PATH)
		e.cfg.botToken = e.configString(CONFIG_BOT_TOKEN)
		e.cfg.botChannelId = e.configInt64(CONFIG_BOT_CHANNEL_ID)
		e.cfg.logDebug = e.configBool(CONFIG_LOG_DEBUG)
		e.cfg.notifyShields = e.configBool(CONFIG_NOTIFY_SHIELDS)
		e.cfg.notifyFighter = e.configBool(CONFIG_NOTIFY_FIGHTER)
		e.cfg.notifyKills = e.configBool(CONFIG_NOTIFY_KILLS)
		e.cfg.notifySilentKills = e.configBool(CONFIG_NOTIFY_SILENT_KILLS)
	}
	sync()

	go func() {
		for {
			select {
			case <-e.stopCh:
				log.Debugln("closing sync config observer")

				return
			case <-e.propagateConfCh:
				sync()
				log.Debugln("sending sync conf signal")
				e.syncConfCh <- struct{}{}
			}
		}
	}()
}
