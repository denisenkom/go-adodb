package mssql

import (
	"context"
	"database/sql/driver"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net"
	"testing"
)

// tests performance of the most common operations

func runTestServer(t *testing.B, handler func(net.Conn)) *Conn {
	addr := &net.TCPAddr{IP: net.IP{127, 0, 0, 1}}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		t.Fatal("Cannot start a listener", err)
	}
	addr = listener.Addr().(*net.TCPAddr)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			t.Log("Failed to accept connection", err)
			return
		}
		handler(conn)
		_ = conn.Close()
	}()
	connStr := fmt.Sprintf("host=%s;port=%d", addr.IP.String(), addr.Port)
	conn, err := driverInstance.open(context.Background(), connStr)
	if err != nil {
		// should not fail here
		t.Fatal("Open connection failed:", err.Error())
	}
	return conn
}

func testConnClose(t *testing.B, conn *Conn) {
	err := conn.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func BenchmarkSelect(b *testing.B) {
	conn := runTestServer(b, func(conn net.Conn) {
		tdsBuf := newTdsBuffer(defaultPacketSize, conn)

		// read prelogin request
		packetType, err := tdsBuf.BeginRead()
		if err != nil {
			b.Fatal("Failed to read PRELOGIN request", err)
		}
		if packetType != packPrelogin {
			b.Fatal("Client sent non PRELOGIN request packet type", packetType)
		}

		// write prelogin response
		fields := map[uint8][]byte{
			preloginENCRYPTION: {encryptNotSup},
		}
		err = writePrelogin(packReply, tdsBuf, fields)
		if err != nil {
			b.Fatal("Writing PRELOGIN packet failed", err)
		}

		// read login request
		packetType, err = tdsBuf.BeginRead()
		if err != nil {
			b.Fatal("Failed to read LOGIN request", err)
		}
		if packetType != packLogin7 {
			b.Fatal("Client sent non LOGIN request packet type", packetType)
		}
		_, err = ioutil.ReadAll(tdsBuf)
		if err != nil {
			b.Fatal(err)
		}

		// send login response
		tdsBuf.BeginPacket(packReply, false)
		buf := make([]byte, 1 + 2 + 2 + 8)
		buf[0] = byte(tokenDone)
		binary.LittleEndian.PutUint16(buf[1:], 0)
		binary.LittleEndian.PutUint16(buf[3:], 0)
		binary.LittleEndian.PutUint64(buf[5:], 0)
		_, err = tdsBuf.Write(buf)
		if err != nil {
			b.Log("writing login reply failed", err)
			return
		}
		err = tdsBuf.FinishPacket()
		if err != nil {
			b.Log("writing login reply failed", err)
			return
		}

		for requests := 0; ; requests += 1{
			// read request
			_, err = tdsBuf.BeginRead()
			if err != nil {
				b.Log(err)
				return
			}
			_, err = ioutil.ReadAll(tdsBuf)
			if err != nil {
				b.Log(err)
				return
			}

			// send response
			tdsBuf.BeginPacket(packReply, false)
			buf, err = base64.StdEncoding.DecodeString("gQEAAAAAACAAOADRAQAAAP0QAMEAAQAAAAAAAAA=")
			if err != nil {
				b.Log(err)
				return
			}
			_, err = tdsBuf.Write(buf)
			if err != nil {
				b.Log("writing login reply failed", err)
				return
			}
			err = tdsBuf.FinishPacket()
			if err != nil {
				b.Log("writing login reply failed", err)
				return
			}
		}
	})
	defer testConnClose(b, conn)

	values := make([]driver.Value, 1)
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		stmt, err := conn.prepareContext(ctx, "select 1")
		if err != nil {
			b.Fatal(err)
		}
		rows, err := stmt.queryContext(ctx, nil)
		if err != nil {
			b.Fatal(err)
		}

		err = rows.Next(values)
		if err != nil {
			b.Fatal(err)
		}

		err = rows.Close()
		if err != nil {
			b.Fatal(err)
		}
		err = stmt.Close()
		if err != nil {
			b.Fatal(err)
		}
	}
}