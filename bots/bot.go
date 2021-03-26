package bots

type Bot interface {
	Start()
	Send(string) error
}
