package notifier

import (
	"fmt"
	"testing"
)

type mockBot struct {
	sentMsg string
	err     error
}

func (m *mockBot) Start() {}
func (m *mockBot) Send(msg string) error {
	m.sentMsg = msg
	return m.err
}

func Test_hullDamageEvent(t *testing.T) {
	tests := []struct {
		name    string
		n       *Notifier
		j       journalEvent
		wantMsg string
		sendErr error
	}{
		{
			name: "disabled fighter notifs",
			n: &Notifier{
				cfg: &Cfg{
					FighterNotifs: false,
				},
			},
			j: journalEvent{
				Fighter: true,
			},
		},
		{
			name: "Ship message",
			n: &Notifier{
				cfg: &Cfg{},
			},
			j: journalEvent{
				Fighter: false,
				Health:  0.399871,
			},
			wantMsg: "Ship hull damage detected, integrity is 40%",
		},
		{
			name: "Fighter message",
			n: &Notifier{
				cfg: &Cfg{
					FighterNotifs: true,
				},
			},
			j: journalEvent{
				Fighter: true,
				Health:  0.413871,
			},
			wantMsg: "Fighter hull damage detected, integrity is 41%",
		},
		{
			name: "bot error",
			n: &Notifier{
				cfg: &Cfg{},
			},
			j: journalEvent{
				Health: 0.413871,
			},
			sendErr: fmt.Errorf("fake"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.n.bot = &mockBot{
				err: tt.sendErr,
			}

			err := hullDamageEvent(tt.n, tt.j)

			if err != nil {
				if tt.sendErr == nil {
					t.Fatalf("sendErr: %v, got: %v", tt.sendErr, err)
				}

				e := fmt.Errorf("error sending message: %v", tt.sendErr)
				if e.Error() != err.Error() {
					t.Fatalf("wantErr: %v, got: %v", e, err)
				}
				return
			}

			msg := tt.n.bot.(*mockBot).sentMsg
			if msg != tt.wantMsg {
				t.Fatalf("wantMsg: %s, got: %s", tt.wantMsg, msg)
			}
		})
	}
}

func Test_diedEvent(t *testing.T) {
	tests := []struct {
		name    string
		n       *Notifier
		j       journalEvent
		wantMsg string
		sendErr error
	}{
		{
			name: "message",
			n: &Notifier{
				cfg: &Cfg{},
			},
			j:       journalEvent{},
			wantMsg: "Your ship has been destroyed",
		},
		{
			name: "error",
			n: &Notifier{
				cfg: &Cfg{},
			},
			j:       journalEvent{},
			sendErr: fmt.Errorf("fake"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.n.bot = &mockBot{
				err: tt.sendErr,
			}

			err := diedEvent(tt.n, tt.j)

			if err != nil {
				if tt.sendErr == nil {
					t.Fatalf("sendErr: %v, got: %v", tt.sendErr, err)
				}

				e := fmt.Errorf("error sending message: %v", tt.sendErr)
				if e.Error() != err.Error() {
					t.Fatalf("wantErr: %v, got: %v", e, err)
				}
				return
			}

			msg := tt.n.bot.(*mockBot).sentMsg
			if msg != tt.wantMsg {
				t.Fatalf("wantMsg: %s, got: %s", tt.wantMsg, msg)
			}
		})
	}
}
