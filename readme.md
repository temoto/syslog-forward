What
====

This little service listens for Syslog packets on network and forwards them to local Systemd journal.

Some configuration is available via command line.

    % ./syslog-journal -help

    Usage of ./syslog-journal:
      -format string
            Fields: ReceiveTime, Timestamp, Addr, Priority, Tag, Content.
            See syntax at https://golang.org/pkg/text/template/
            (default "from={{.Addr}} {{.Tag}}: {{.Content}}")
      -journal-buffer-size int
             (default 32768)
      -listen-udp string
             (default ":514")
      -queue-length int
             (default 100)
      -udp-buffer-size int
             (default 16384)

Why
---

It may be possible to configure rsyslog or syslog-ng to do the same, and I've had success with plain forwarding
via rsyslog, but didn't find a way to add source address info. Now with Systemd spread there's no need in
full featured syslog server so I think this program is a better replacement for complex configuration.

Contribute
==========

Please send patches to temotor@gmail.com or Github https://github.com/temoto/syslog-journal
