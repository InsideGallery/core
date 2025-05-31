package honeypot

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"time"
)

func Honeypot(listenPort string) error {
	listener, err := net.Listen("tcp", ":"+listenPort)
	if err != nil {
		return err
	}

	defer listener.Close()

	slog.Default().Info("Honeypot running and listening on port", "listenPort", listenPort)

	for {
		conn, err := listener.Accept()
		if err != nil {
			slog.Default().Error("Error accepting connection", "err", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	slog.Default().Info("Intrusion attempt detected",
		"at", time.Now().Format(time.RFC3339),
		"from", conn.RemoteAddr().String())

	_, err := fmt.Fprintln(conn, "SSH-2.0-OpenSSH_7.9p1 Debian-10+deb10u2") // Fake SSH banner
	if err != nil {
		slog.Default().Error("Error writing connect line", "err", err)
	}

	// Attempt to read data from the attacker
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		slog.Default().Info("Received input",
			"from", conn.RemoteAddr().String(),
			"line", line)
	}

	if err := scanner.Err(); err != nil {
		slog.Default().Error("Error reading",
			"from", conn.RemoteAddr().String(),
			"err", err)
	}
}
