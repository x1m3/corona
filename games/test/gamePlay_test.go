package test

import (
	"testing"
)

func TestGamePlayInit(t *testing.T) {

	_, _, gamePlay := helperInitGamePlayTest(1000, 1000)

	go gamePlay.Init()

	event := <-gamePlay.EventListener()

	if got, expected := event.(*traceEvent).Msg, "INIT"; got != expected {
		t.Errorf("Received wrong event. Expecting <%s>, got <%s>", got, expected)
	}
}

func TestGamePlayProcessCommand (t *testing.T) {
	_, _, gamePlay := helperInitGamePlayTest(1000, 1000)

	go func() {
		gamePlay.Init()
		gamePlay.ProcessCommand(nil)
	}()

	event := <-gamePlay.EventListener()

	if got, expected := event.(*traceEvent).Msg, "INIT"; got != expected {
		t.Errorf("Received wrong event. Expecting <%s>, got <%s>", expected, got)
	}

	event = <-gamePlay.EventListener()

	if got, expected := event.(*traceEvent).Msg, "PROCESS_COMMAND"; got != expected {
		t.Errorf("Received wrong event. Expecting <%s>, got <%s>", expected, got)
	}

}

func TestGamePlayStop(t *testing.T) {
	_, _, gamePlay := helperInitGamePlayTest(1000, 1000)

	go func() {
		gamePlay.Init()
		gamePlay.Stop()
	}()

	event := <-gamePlay.EventListener()

	if got, expected := event.(*traceEvent).Msg, "INIT"; got != expected {
		t.Errorf("Received wrong event. Expecting <%s>, got <%s>", expected, got)
	}

}
