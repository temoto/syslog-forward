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
	if err := msg.Parse([]byte("<152>Aug 28 11:13:00 mysqld[3886]: hello world\x00")); err != nil {
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
