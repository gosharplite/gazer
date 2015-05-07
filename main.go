package main

import (
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

/* slower, by we can print/log everything */
func myrawcopy(dst, src net.Conn) (written int64, err error) {

	buf := make([]byte, 32*1024)

	for {

		go logIP(src.RemoteAddr().String(), dst.RemoteAddr().String())

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

	//io.Copy(dst,src);
	myrawcopy(dst, src)

	dst.Close()
	src.Close()
}

func handleclient(c net.Conn) {

	config := tls.Config{InsecureSkipVerify: true}

	// Set "192.168.1.22    gigacard1.gigacloud.tw" in /etc/hosts.
	conn, err := tls.Dial("tcp", "gigacard1.gigacloud.tw:443", &config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}

	// gazer to IIS
	go myiocopy(conn, c)

	// IIS to gazer
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

		log.Printf("server: accepted from %s", conn.RemoteAddr())

		go handleclient(conn)
	}
}

func logIP(src, dst string) {

	_, err := http.PostForm("http://192.168.1.32:8080/gazer/logip",
		url.Values{"src": {src}, "dst": {dst}})

	if err != nil {
		log.Printf("err: http.PostForm: %v", err)
	}
}
