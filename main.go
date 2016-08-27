package main

import (
	"bytes"
	"errors"
	"flag"
	"github.com/coreos/go-systemd/daemon"
	"github.com/coreos/go-systemd/journal"
	"log"
	"net"
	"regexp"
	"strconv"
	"text/template"
	"time"
)

type tmsg struct {
	ReceiveTime time.Time
	Timestamp   time.Time
	Addr        net.Addr
	Priority    byte
	Tag         string
	Content     string
}

type tconfig struct {
	format                 string
	listenUDP              string
	listenTCP              string
	queueLength            int
	tweakUDPBufferSize     int
	tweakTCPBufferSize     int
	tweakJournalBufferSize int
	queue                  chan tmsg
}

var (
	errUnknownMessageFormat = errors.New("unknown message format")
	reSyslogFormat1         = regexp.MustCompile(`^\<([0-9]{1,3})\>([A-Za-z]{3} [ 0-9][0-9] [0-9][0-9]:[0-9][0-9]:[0-9][0-9]) ([^:]+): (.+)?\x00$`)
)

func main() {
	daemon.SdNotify("READY=0\nSTATUS=init\n")
	var config tconfig
	flag.StringVar(&config.format, "format", "from={{.Addr}} {{.Tag}}: {{.Content}}",
		"Fields: ReceiveTime, Timestamp, Addr, Priority, Tag, Content. See syntax at https://golang.org/pkg/text/template/")
	flag.StringVar(&config.listenUDP, "listen-udp", ":514", "Empty string to disable")
	flag.StringVar(&config.listenTCP, "listen-tcp", "", "Empty string to disable")
	flag.IntVar(&config.queueLength, "queue-length", 100, "")
	flag.IntVar(&config.tweakUDPBufferSize, "udp-buffer-size", 16<<10, "")
	// flag.IntVar(&config.tweakTCPBufferSize, "tcp-buffer-size", 16<<10, "")
	flag.IntVar(&config.tweakJournalBufferSize, "journal-buffer-size", 32<<10, "")
	flag.Parse()
	formatTemplate := template.Must(template.New("").Parse(config.format))
	config.queue = make(chan tmsg, config.queueLength)
	if !journal.Enabled() {
		log.Fatal("systemd journal connection failed")
	}

	if (config.listenUDP == "") && (config.listenTCP == "") {
		log.Fatal("no listeners - nothing to do")
	}
	go readUDP(&config)
	go readTCP(&config)

	buf := bytes.NewBuffer(make([]byte, 0, config.tweakJournalBufferSize))
	vars := map[string]string{
		"OBJECT_COMM":       "",
		"SYSLOG_IDENTIFIER": "",
	}
	daemon.SdNotify("READY=1\nSTATUS=work\n")
	for msg := range config.queue {
		buf.Reset()
		vars["OBJECT_COMM"] = msg.Addr.String()
		vars["SYSLOG_IDENTIFIER"] = msg.Tag
		if err := formatTemplate.Execute(buf, msg); err != nil {
			log.Fatal(err)
		}
		journal.Send(buf.String(), journal.Priority(msg.Priority), vars)
	}
}

func readUDP(config *tconfig) {
	if config.listenUDP == "" {
		return
	}
	conn, err := net.ListenPacket("udp", config.listenUDP)
	if err != nil {
		log.Fatal(err)
	}

	buf := make([]byte, config.tweakUDPBufferSize)
	for {
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			log.Printf("readUDP() ReadFrom error n=%d addr=%s err=%s", n, addr, err)
			continue
		}
		if n == 0 {
			log.Printf("readUDP() received 0 length packet addr=%s", addr)
			continue
		}
		msg := tmsg{
			ReceiveTime: time.Now(),
			Addr:        addr,
		}
		err = msg.Parse(buf[:n])
		if err != nil {
			log.Printf("readUDP() msg.Parse error n=%d addr=%s buf='%s' err=%s", n, addr, string(buf[:n]), err)
		}

		config.queue <- msg
	}
}

func readTCP(config *tconfig) {
	if config.listenTCP == "" {
		return
	}
	log.Fatal("readTCP 501 not implemented yet")
	// l, err := net.Listen("tcp", bindSpec)
	// if err != nil {
	// 	log.Fatal(err)
	// }
}

func (msg *tmsg) Parse(b []byte) (err error) {
	match := reSyslogFormat1.FindStringSubmatch(string(b))
	if len(match) == 5 {
		var x uint64
		x, err = strconv.ParseUint(match[1], 10, 8)
		if err != nil {
			return err
		}
		msg.Priority = byte(x)

		msg.Timestamp, err = time.Parse(time.Stamp, match[2])
		if err != nil {
			return err
		}

		msg.Tag = match[3]
		msg.Content = match[4]
	} else {
		return errUnknownMessageFormat
	}

	return nil
}
