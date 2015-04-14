package main

import (
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

/* slower, by we can print/log everything */
func myrawcopy(dst, src net.Conn) (written int64, err error) {

	buf := make([]byte, 32*1024)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {

			fmt.Printf("%v ms : %s -> %s (%v + %v)\n", time.Now().UnixNano()/1000000, src.RemoteAddr(), dst.RemoteAddr(), written, nr)

			b := buf[0:nr]

			if !strings.Contains(src.RemoteAddr().String(), "192.168.1.22:443") {

				if nr > 100 {
					bs := string(b[:100])
					if strings.Contains(bs, "HTTP") {
						i := strings.Index(bs, "\n")
						if i == -1 {
							fmt.Printf("%s\n\n", bs)
						} else {
							fmt.Printf("%s\n\n", bs[:i])
						}
					}
				}
			}

			nw, ew := dst.Write(b)
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er == io.EOF {
			break
		}
		if er != nil {
			err = er
			break
		}
	}
	return written, err
}

func myiocopy(dst net.Conn, src net.Conn) {
	myrawcopy(dst, src)
	//io.Copy(dst,src);
	dst.Close()
	src.Close()
}

func handleclient(c net.Conn) {
	config := tls.Config{InsecureSkipVerify: true}
	conn, err := tls.Dial("tcp", "gigacard1.gigacloud.tw:443", &config)
	checkError(err)

	go myiocopy(conn, c)

	//io.Copy(c, conn)
	myrawcopy(c, conn)
	c.Close()
	conn.Close()
}

func main() {
	cert, err := tls.LoadX509KeyPair("cert.pem", "server.key")
	if err != nil {
		log.Fatalf("server: loadkeys: %s", err)
	}
	config := tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS10,
	}
	config.Rand = rand.Reader
	service := "0.0.0.0:443"
	listener, err := tls.Listen("tcp", service, &config)
	if err != nil {
		log.Fatalf("server: listen: %s", err)
	}
	log.Printf("server: listening on %s for https, connects to https://example.com:443", service)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("server: accept: %s", err)
			break
		}
		defer conn.Close()
		log.Printf("server: accepted from %s", conn.RemoteAddr())
		go handleclient(conn)
	}
}
