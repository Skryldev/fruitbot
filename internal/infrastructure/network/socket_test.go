package network

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
)

func TestSocket_SetInfo(t *testing.T) {
	s := NewSocket(nil)
	
	s.SetInfo(12345, 67890, 2, "TestUser")
	
	if s.UserID() != 12345 {
		t.Errorf("UserID() = %d, want 12345", s.UserID())
	}
	if s.TribeID() != 67890 {
		t.Errorf("TribeID() = %d, want 67890", s.TribeID())
	}
	if s.AvatarID() != 2 {
		t.Errorf("AvatarID() = %d, want 2", s.AvatarID())
	}
	if s.UserName() != "TestUser" {
		t.Errorf("UserName() = %s, want TestUser", s.UserName())
	}
}

func TestSocket_AddHandler(t *testing.T) {
	s := NewSocket(nil)
	
	var wg sync.WaitGroup
	wg.Add(1)
	
	handler := func(msg Message) {
		defer wg.Done()
		if msg["test"] != "value" {
			t.Errorf("msg[test] = %v, want 'value'", msg["test"])
		}
	}
	
	s.AddHandler("test_type", handler)
	
	// Simulate message dispatch
	msg := Message{"push_message_type": "test_type", "test": "value"}
	s.dispatchMessage(msg)
	
	wg.Wait()
}

func TestSocket_Stats(t *testing.T) {
	s := NewSocket(nil)
	
	stats := s.Stats()
	
	if stats.Connected {
		t.Error("socket should not be connected initially")
	}
	if stats.MessagesSent != 0 {
		t.Error("messages sent should be 0")
	}
	if stats.MessagesReceived != 0 {
		t.Error("messages received should be 0")
	}
}

func TestSocket_Close(t *testing.T) {
	s := NewSocket(nil)
	s.Close(false)
	
	if s.IsConnected() {
		t.Error("socket should not be connected after close")
	}
}

// Mock server for integration testing
func startMockServer(t *testing.T) (net.Listener, string) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		
		// Echo back any received data
		buf := make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				return
			}
			conn.Write(buf[:n])
		}
	}()
	
	return listener, listener.Addr().String()
}

func BenchmarkSocket_MessageProcessing(b *testing.B) {
	s := NewSocket(nil)
	
	var buffer strings.Builder
	msg := fmt.Sprintf("%s%s%s", startMarker, "encrypted_data", endMarker)
	
	b.ReportAllocs()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		buffer.Reset()
		buffer.WriteString(msg)
		s.processBuffer(&buffer)
	}
}