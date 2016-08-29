package main

import (
	"bytes"
	"net"
	"testing"
	"text/template"
	"time"
)

func BenchmarkRenderTemplate(b *testing.B) {
	msg := tmsg{
		ReceiveTime: time.Now(),
		Addr:        &net.IPAddr{IP: net.ParseIP("1.2.3.4")},
	}
	bs := []byte("<152>Aug 28 11:13:00 mysqld[3886]: hello world\x00")
	if err := msg.Parse(&tconfig{forcePriority: -1}, bs); err != nil {
		b.Fatal("msg.Parse", err)
	}
	formatTemplate := template.Must(template.New("").Parse("from={{.Addr}} {{.Content}}"))
	buf := bytes.NewBuffer(nil)

	b.ResetTimer()
	for i := 1; i <= b.N; i++ {
		buf.Reset()
		if err := formatTemplate.Execute(buf, msg); err != nil {
			b.Fatal(err)
		}
		if buf.String() != "from=1.2.3.4 hello world" {
			b.Fatal("test fail", buf.String())
		}
	}
}

func TestParseV1(t *testing.T) {
	msg := tmsg{
		ReceiveTime: time.Now(),
		Addr:        &net.IPAddr{IP: net.ParseIP("1.2.3.4")},
	}
	bs := []byte(`<13>1 2016-08-28T19:15:47.451655+05:00 GFE app - - [timeQuality tzKnown="1" isSynced="1" syncAccuracy="491213"] sonoff`)
	if err := msg.Parse(&tconfig{forcePriority: -1}, bs); err != nil {
		t.Fatal("msg.Parse", err)
	}
	if msg.Hostname != "GFE" {
		t.Fatalf("invalid hostname: '%s'", msg.Hostname)
	}
	if msg.Priority != 5 {
		t.Fatalf("invalid priority: %v", msg.Priority)
	}
	if msg.Tag != "app" {
		t.Fatalf("invalid tag: '%s'", msg.Tag)
	}
	if msg.Content != "sonoff" {
		t.Fatalf("invalid content: '%s'", msg.Content)
	}
}

func TestParseSimpleWithTag(t *testing.T) {
	msg := tmsg{
		ReceiveTime: time.Now(),
		Addr:        &net.IPAddr{IP: net.ParseIP("1.2.3.4")},
	}
	bs := []byte(`sonoff: Connected`)
	if err := msg.Parse(&tconfig{forcePriority: -1, defaultPriority: 5}, bs); err != nil {
		t.Fatal("msg.Parse", err)
	}
	if msg.Priority != 5 {
		t.Fatalf("invalid priority: %v", msg.Priority)
	}
	if msg.Tag != "sonoff" {
		t.Fatalf("invalid tag: '%s'", msg.Tag)
	}
	if msg.Content != "Connected" {
		t.Fatalf("invalid content: '%s'", msg.Content)
	}
}

func TestParseSimple(t *testing.T) {
	msg := tmsg{
		ReceiveTime: time.Now(),
		Addr:        &net.IPAddr{IP: net.ParseIP("1.2.3.4")},
	}
	bs := []byte(`just a message`)
	if err := msg.Parse(&tconfig{forcePriority: -1, defaultPriority: 6}, bs); err != nil {
		t.Fatal("msg.Parse", err)
	}
	if msg.Priority != 6 {
		t.Fatalf("invalid priority: %v", msg.Priority)
	}
	if msg.Content != "just a message" {
		t.Fatalf("invalid content: '%s'", msg.Content)
	}
}
