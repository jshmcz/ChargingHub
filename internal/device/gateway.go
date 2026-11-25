package device

import (
	"bufio"
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type Gateway struct {
	DB    *sql.DB
	mu    sync.Mutex
	conns map[string]net.Conn
	log   *log.Logger
}

func NewGateway(db *sql.DB) *Gateway {
	return &Gateway{DB: db, conns: map[string]net.Conn{}, log: log.New(os.Stdout, "[device] ", log.LstdFlags)}
}

func (g *Gateway) Start(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	g.log.Printf("device gateway listening on %s", addr)
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				g.log.Printf("accept: %v", err)
				continue
			}
			go g.handleConn(conn)
		}
	}()
	return nil
}

func (g *Gateway) handleConn(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		_ = conn.SetReadDeadline(time.Now().Add(5 * time.Minute))
		b, err := reader.ReadByte()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				g.log.Printf("read: %v", err)
			}
			return
		}
		if b != 0x68 {
			continue
		}
		cmd, err := reader.ReadByte()
		if err != nil {
			return
		}
		devLen, err := reader.ReadByte()
		if err != nil {
			return
		}
		devRaw := make([]byte, devLen)
		if _, err = io.ReadFull(reader, devRaw); err != nil {
			return
		}
		dataLen, err := reader.ReadByte()
		if err != nil {
			return
		}
		data := make([]byte, dataLen)
		if _, err = io.ReadFull(reader, data); err != nil {
			return
		}
		crcRaw := make([]byte, 2)
		if _, err = io.ReadFull(reader, crcRaw); err != nil {
			return
		}
		raw := append([]byte{0x68, cmd, devLen}, devRaw...)
		raw = append(raw, dataLen)
		raw = append(raw, data...)
		raw = append(raw, crcRaw...)
		frame, err := parseFrame(raw)
		if err != nil {
			g.log.Printf("bad frame: %v", err)
			continue
		}
		g.handleFrame(conn, frame)
	}
}

func (g *Gateway) handleFrame(conn net.Conn, f Frame) {
	switch f.Cmd {
	case 0x01:
		var secret string
		err := g.DB.QueryRow(`SELECT secret FROM piles WHERE device_no=?`, f.DeviceNo).Scan(&secret)
		if err != nil || secret != string(f.Data) {
			_, _ = conn.Write(buildFrame(0x90, f.DeviceNo, []byte("AUTH_FAIL")))
			return
		}
		_, _ = g.DB.Exec(`UPDATE piles SET status='online',last_seen_at=? WHERE device_no=?`, time.Now().Format(time.RFC3339), f.DeviceNo)
		g.mu.Lock()
		g.conns[f.DeviceNo] = conn
		g.mu.Unlock()
		_, _ = conn.Write(buildFrame(0x81, f.DeviceNo, []byte("OK")))
		g.flushPending(f.DeviceNo)
	case 0x02:
		_, _ = g.DB.Exec(`UPDATE piles SET status='online',last_seen_at=? WHERE device_no=?`, time.Now().Format(time.RFC3339), f.DeviceNo)
		_, _ = conn.Write(buildFrame(0x82, f.DeviceNo, nil))
	case 0x03:
		g.handleTelemetry(f)
		_, _ = conn.Write(buildFrame(0x83, f.DeviceNo, nil))
	default:
		_, _ = conn.Write(buildFrame(0x80|f.Cmd, f.DeviceNo, nil))
	}
}

func (g *Gateway) handleTelemetry(f Frame) {
	if len(f.Data) < 5 {
		return
	}
	gunNo := int(f.Data[0])
	energyWh := binary.BigEndian.Uint32(f.Data[1:5])
	var gunID int64
	if err := g.DB.QueryRow(`SELECT g.id FROM guns g JOIN piles p ON p.id=g.pile_id
		WHERE p.device_no=? AND g.gun_no=?`, f.DeviceNo, gunNo).Scan(&gunID); err != nil {
		return
	}
	_, _ = g.DB.Exec(`UPDATE orders SET energy_wh=?,last_report_at=? WHERE gun_id=? AND status='charging'`,
		int64(energyWh), time.Now().Format(time.RFC3339), gunID)
}

func commandByte(command string) byte {
	switch strings.ToUpper(command) {
	case "START":
		return 0x10
	case "STOP":
		return 0x11
	case "PARAM":
		return 0x12
	case "REBOOT":
		return 0x13
	default:
		return 0x10
	}
}

func (g *Gateway) SendCommand(deviceNo string, gunNo int, command, payload string) bool {
	if deviceNo == "" {
		return false
	}
	data := append([]byte{byte(gunNo)}, []byte(payload)...)
	g.mu.Lock()
	conn := g.conns[deviceNo]
	g.mu.Unlock()
	if conn != nil {
		_, err := conn.Write(buildFrame(commandByte(command), deviceNo, data))
		if err == nil {
			return true
		}
		g.mu.Lock()
		delete(g.conns, deviceNo)
		g.mu.Unlock()
	}
	_, _ = g.DB.Exec(`INSERT INTO device_commands(device_no,gun_no,command,payload,status,created_at)
		VALUES(?,?,?,?,?,?)`, deviceNo, gunNo, strings.ToUpper(command), payload, "pending", time.Now().Format(time.RFC3339))
	return false
}

func (g *Gateway) flushPending(deviceNo string) {
	rows, err := g.DB.Query(`SELECT id,gun_no,command,payload FROM device_commands WHERE device_no=? AND status='pending' ORDER BY id`, deviceNo)
	if err != nil {
		return
	}
	type cmd struct {
		id, gun int64
		name, payload string
	}
	var cmds []cmd
	for rows.Next() {
		var c cmd
		if rows.Scan(&c.id, &c.gun, &c.name, &c.payload) == nil {
			cmds = append(cmds, c)
		}
	}
	rows.Close()
	g.mu.Lock()
	conn := g.conns[deviceNo]
	g.mu.Unlock()
	if conn == nil {
		return
	}
	for _, c := range cmds {
		data := append([]byte{byte(c.gun)}, []byte(c.payload)...)
		if _, err = conn.Write(buildFrame(commandByte(c.name), deviceNo, data)); err != nil {
			return
		}
		_, _ = g.DB.Exec(`UPDATE device_commands SET status='sent',acked_at=? WHERE id=?`, time.Now().Format(time.RFC3339), c.id)
	}
}

func (g *Gateway) Logf(format string, args ...any) {
	g.log.Printf(fmt.Sprintf(format, args...))
}
