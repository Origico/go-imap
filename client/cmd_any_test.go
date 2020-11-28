package client

import (
	"testing"

	"github.com/emersion/go-imap"
)

func TestClient_Capability(t *testing.T) {
	c, s := newTestClient(t)
	defer s.Close()

	var caps map[string]bool
	done := make(chan error, 1)
	go func() {
		var err error
		caps, err = c.Capability()
		done <- err
	}()

	tag, cmd := s.ScanCmd()
	if cmd != "CAPABILITY" {
		t.Fatalf("client sent command %v, want CAPABILITY", cmd)
	}
	s.WriteString("* CAPABILITY IMAP4rev1 XTEST\r\n")
	s.WriteString(tag + " OK CAPABILITY completed.\r\n")

	if err := <-done; err != nil {
		t.Error("c.Capability() = ", err)
	}

	if !caps["XTEST"] {
		t.Error("XTEST capability missing")
	}
}

func TestClient_Noop(t *testing.T) {
	c, s := newTestClient(t)
	defer s.Close()

	done := make(chan error, 1)
	go func() {
		done <- c.Noop()
	}()

	tag, cmd := s.ScanCmd()
	if cmd != "NOOP" {
		t.Fatalf("client sent command %v, want NOOP", cmd)
	}
	s.WriteString(tag + " OK NOOP completed\r\n")

	if err := <-done; err != nil {
		t.Error("c.Noop() = ", err)
	}
}

func TestClient_Logout(t *testing.T) {
	c, s := newTestClient(t)
	defer s.Close()

	done := make(chan error, 1)
	go func() {
		done <- c.Logout()
	}()

	tag, cmd := s.ScanCmd()
	if cmd != "LOGOUT" {
		t.Fatalf("client sent command %v, want LOGOUT", cmd)
	}
	s.WriteString("* BYE Client asked to close the connection.\r\n")
	s.WriteString(tag + " OK LOGOUT completed\r\n")

	if err := <-done; err != nil {
		t.Error("c.Logout() =", err)
	}

	if state := c.State(); state != imap.LogoutState {
		t.Errorf("c.State() = %v, want %v", state, imap.LogoutState)
	}
}

func TestClient_Fubb(t *testing.T) {
	c, s := newTestClient(t)
	defer s.Close()
	setClientState(c, imap.SelectedState, nil)
	seqset, _ := imap.ParseSeqSet("1:4")
	fields := []imap.FetchItem{imap.FetchUid, imap.FetchModSeq, imap.FetchFlags}

	done := make(chan error, 1)
	messages := make(chan *imap.Message, 2)
	go func() {
		done <- c.Fetch(seqset, fields, messages)
	}()
	tag, cmd := s.ScanCmd()
	if cmd != "FETCH 1:4 (UID MODSEQ FLAGS)" {
		t.Fatalf("client sent command %v, want %v", cmd, "FETCH 1:4 (UID BODY[])")
	}
	// s.WriteString("* 1 FETCH (MODSEQ (5) FLAGS (\\Flagged \\Seen) UID 1)\r\n")
	// s.WriteString("* 2 FETCH (MODSEQ (10) FLAGS (\\Flagged \\Seen) UID 2)\r\n")
	// s.WriteString("* 3 FETCH (MODSEQ (15) FLAGS (\\Flagged \\Seen) UID 3)\r\n")
	// s.WriteString("* 4 FETCH (MODSEQ (20) FLAGS (\\Flagged \\Seen) UID 4)\r\n")
	// s.WriteString(tag + " OK FETCH completed\r\n")

	s.WriteString("* 1 FETCH (MODSEQ (5) BODY[1] {3}\r\n")
	s.WriteString("Hey")
	s.WriteString(")\r\n")

	s.WriteString(tag + " OK FETCH completed\r\n")

	// fmt.Println("Sended")

	if err := <-done; err != nil {
		t.Fatalf("c.Fetch() = %v", err)
	}

	<-messages

	// section, _ := imap.ParseBodySectionName("BODY[]")
	// msg := <-messages
	// fmt.Println(msg)
	// if msg.SeqNum != 2 {
	// 	t.Errorf("First message has bad sequence number: %v", msg.SeqNum)
	// }
	// if msg.Uid != 42 {
	// 	t.Errorf("First message has bad UID: %v", msg.Uid)
	// }
	// if body, _ := ioutil.ReadAll(msg.GetBody(section)); string(body) != "I love potatoes." {
	// 	t.Errorf("First message has bad body: %q", body)
	// }

	// msg = <-messages
	// if msg.SeqNum != 3 {
	// 	t.Errorf("First message has bad sequence number: %v", msg.SeqNum)
	// }
	// if msg.Uid != 28 {
	// 	t.Errorf("Second message has bad UID: %v", msg.Uid)
	// }
	// if body, _ := ioutil.ReadAll(msg.GetBody(section)); string(body) != "Hello World!" {
	// 	t.Errorf("Second message has bad body: %q", body)
	// }
}
