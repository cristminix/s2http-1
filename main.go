package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/elazarl/goproxy"
	"golang.org/x/net/proxy"
)

func main() {
	socks := flag.String("socks", "127.0.0.1:1081", "socks url (default)")
	port := flag.String("port", "8080", "port to listen")
	pure := flag.Bool("pure", false, "pure http/https proxy without socks")
	verbose := flag.Bool("v", true, "verbose")
	version := flag.Bool("version", false, "show version.")
	keepalive := flag.Bool("keepalive", true, "enable keep-alive connections")
	keepaliveTimeout := flag.Int("keepalive-timeout", 60, "keep-alive timeout in seconds")
	maxIdleConns := flag.Int("max-idle-conns", 100, "maximum idle connections")
	dummyPacket := flag.Bool("dummy-packet", true, "send dummy packets to prevent connection reset errors")
	dummyPacketInterval := flag.Int("dummy-packet-interval", 30, "interval in seconds to send dummy packets")
	dummyPacketSize := flag.Int("dummy-packet-size", 1, "size of dummy packet in bytes")
	errorHandling := flag.String("error-handling", "retry", "error handling strategy: retry, ignore, or fatal")
	minPacketInterval := flag.Int("min-packet-interval", 5, "minimum interval in seconds to send dummy packets")
	readTimeout := flag.Int("read-timeout", 5, "read timeout in seconds")
	writeTimeout := flag.Int("write-timeout", 5, "write timeout in seconds")

	flag.Parse()

	// Display version info.
	if *version {
		fmt.Println("version=1.1, 2017-9-5, by chinglinwen")
		os.Exit(0)
	}

	dialer, err := proxy.SOCKS5("tcp", *socks, nil, proxy.Direct)
	if err != nil {
		fmt.Fprintln(os.Stderr, "can't connect to the proxy:", err)
		os.Exit(1)
	}

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = *verbose

	// Terapkan timeout
	readTimeoutDuration := time.Duration(*readTimeout) * time.Second
	writeTimeoutDuration := time.Duration(*writeTimeout) * time.Second

	if !*pure {
		// Buat dialer kustom dengan timeout
		proxy.Tr.Dial = func(network, addr string) (net.Conn, error) {
			// Coba hubungkan dengan mekanisme retry
			var conn net.Conn
			var err error

			// Maksimal 3 percobaan
			for i := 0; i < 30000; i++ {
				conn, err = dialer.Dial(network, addr)
				if err == nil {
					break
				}
				// Tunggu sebentar sebelum mencoba lagi
				time.Sleep(time.Duration(i+1) * time.Second)
			}

			if err != nil {
				return nil, err
			}

			// Atur timeout untuk koneksi
			conn.SetDeadline(time.Now().Add(readTimeoutDuration))

			// Jika dummy packet diaktifkan, bungkus koneksi dengan dummyPacketConn
			if *dummyPacket {
				dummyConn := &dummyPacketConn{
					Conn:             conn,
					sendDummy:        true,
					interval:         time.Duration(*dummyPacketInterval) * time.Second,
					minInterval:      time.Duration(*minPacketInterval) * time.Second,
					packetSize:       *dummyPacketSize,
					errorHandling:    *errorHandling,
					lastActivityTime: time.Now(),
				}
				// Mulai timer untuk mengirim paket null secara berkala
				dummyConn.startNullPacketTimer()
				return dummyConn, nil
			}
			return conn, nil
		}
	} else {
		// Untuk proxy murni, atur timeout pada dialer default
		proxy.Tr.Dial = func(network, addr string) (net.Conn, error) {
			d := net.Dialer{Timeout: readTimeoutDuration}

			// Coba hubungkan dengan mekanisme retry
			var conn net.Conn
			var err error

			// Maksimal 3 percobaan
			for i := 0; i < 30000; i++ {
				conn, err = d.Dial(network, addr)
				if err == nil {
					break
				}
				// Tunggu sebentar sebelum mencoba lagi
				time.Sleep(time.Duration(i+1) * time.Second)
			}

			if err != nil {
				return nil, err
			}

			// Jika dummy packet diaktifkan, bungkus koneksi dengan dummyPacketConn
			if *dummyPacket {
				dummyConn := &dummyPacketConn{
					Conn:             conn,
					sendDummy:        true,
					interval:         time.Duration(*dummyPacketInterval) * time.Second,
					minInterval:      time.Duration(*minPacketInterval) * time.Second,
					packetSize:       *dummyPacketSize,
					errorHandling:    *errorHandling,
					lastActivityTime: time.Now(),
				}
				// Mulai timer untuk mengirim paket null secara berkala
				dummyConn.startNullPacketTimer()
				return dummyConn, nil
			}
			return conn, nil
		}
	}

	// Atur timeout untuk server HTTP
	server := &http.Server{
		Addr:              ":" + *port,
		Handler:           proxy,
		ReadTimeout:       readTimeoutDuration,
		WriteTimeout:      writeTimeoutDuration,
		IdleTimeout:       time.Duration(*keepaliveTimeout) * time.Second,
		ReadHeaderTimeout: readTimeoutDuration,
	}

	// Konfigurasi keep-alive
	server.SetKeepAlivesEnabled(*keepalive)

	// Untuk mengatur max idle connections, kita perlu membuat Transport kustom
	proxy.Tr = &http.Transport{
		Dial:                proxy.Tr.Dial,
		MaxIdleConns:        *maxIdleConns,
		IdleConnTimeout:     time.Duration(*keepaliveTimeout) * time.Second,
		TLSHandshakeTimeout: readTimeoutDuration,
	}

	log.Fatal(server.ListenAndServe())
}

// dummyPacketConn adalah wrapper untuk net.Conn yang mengirim paket null secara berkala
type dummyPacketConn struct {
	net.Conn
	sendDummy        bool
	timer            *time.Timer
	interval         time.Duration
	minInterval      time.Duration
	packetSize       int
	errorHandling    string
	stopCh           chan struct{}
	stopped          bool
	mu               sync.Mutex
	lastActivityTime time.Time
}

// Write menulis data ke koneksi dan mengirim paket null jika terjadi error dan fitur diaktifkan
func (c *dummyPacketConn) Write(b []byte) (n int, err error) {
	// Perbarui waktu aktivitas terakhir
	c.lastActivityTime = time.Now()

	n, err = c.Conn.Write(b)
	if err != nil && c.sendDummy {
		// Deteksi error "broken pipe" dan "network is unreachable" secara spesifik
		if strings.Contains(err.Error(), "broken pipe") || strings.Contains(err.Error(), "connection reset by peer") || strings.Contains(err.Error(), "network is unreachable") {
			// Kirim paket null untuk mencoba mempertahankan koneksi
			c.sendNullPacket()

			// Terapkan strategi penanganan error
			switch c.errorHandling {
			case "ignore":
				// Abaikan error dan kembalikan sukses
				return n, nil
			case "fatal":
				// Hentikan program jika terjadi error
				log.Fatal("Fatal error occurred: ", err)
			default:
				// Default adalah "retry" - kembalikan error seperti biasa
				return n, err
			}
		}

		// Untuk error lain, kirim paket null dan terapkan strategi penanganan error
		c.sendNullPacket()

		// Terapkan strategi penanganan error
		switch c.errorHandling {
		case "ignore":
			// Abaikan error dan kembalikan sukses
			return n, nil
		case "fatal":
			// Hentikan program jika terjadi error
			log.Fatal("Fatal error occurred: ", err)
		default:
			// Default adalah "retry" - kembalikan error seperti biasa
			return n, err
		}
	}
	return n, err
}

// Read membaca data dari koneksi dan mengirim paket null jika terjadi error dan fitur diaktifkan
func (c *dummyPacketConn) Read(b []byte) (n int, err error) {
	// Perbarui waktu aktivitas terakhir
	c.lastActivityTime = time.Now()

	n, err = c.Conn.Read(b)
	if err != nil && c.sendDummy {
		// Deteksi error "broken pipe" dan "network is unreachable" secara spesifik
		if strings.Contains(err.Error(), "broken pipe") || strings.Contains(err.Error(), "connection reset by peer") || strings.Contains(err.Error(), "network is unreachable") {
			// Kirim paket null untuk mencoba mempertahankan koneksi
			c.sendNullPacket()

			// Terapkan strategi penanganan error
			switch c.errorHandling {
			case "ignore":
				// Abaikan error dan kembalikan sukses
				return n, nil
			case "fatal":
				// Hentikan program jika terjadi error
				log.Fatal("Fatal error occurred: ", err)
			default:
				// Default adalah "retry" - kembalikan error seperti biasa
				return n, err
			}
		}

		// Untuk error lain, kirim paket null dan terapkan strategi penanganan error
		c.sendNullPacket()

		// Terapkan strategi penanganan error
		switch c.errorHandling {
		case "ignore":
			// Abaikan error dan kembalikan sukses
			return n, nil
		case "fatal":
			// Hentikan program jika terjadi error
			log.Fatal("Fatal error occurred: ", err)
		default:
			// Default adalah "retry" - kembalikan error seperti biasa
			return n, err
		}
	}
	return n, err
}

// sendNullPacket mengirim paket null ke koneksi
func (c *dummyPacketConn) sendNullPacket() error {
	if c.sendDummy {
		// Buat paket null dengan ukuran yang ditentukan
		nullData := make([]byte, c.packetSize)
		// Isi dengan null bytes
		for i := range nullData {
			nullData[i] = 0
		}
		_, err := c.Conn.Write(nullData)
		return err
	}
	return nil
}

// startNullPacketTimer memulai timer untuk mengirim paket null secara berkala
func (c *dummyPacketConn) startNullPacketTimer() {
	if !c.sendDummy {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.stopped {
		return
	}

	// Hitung interval yang sesuai
	interval := c.interval

	// Jika interval minimum lebih kecil dari interval default, gunakan interval minimum
	if c.minInterval < interval {
		interval = c.minInterval
	}

	// Reset timer jika sudah ada
	if c.timer != nil {
		c.timer.Reset(interval)
	} else {
		c.timer = time.AfterFunc(interval, func() {
			c.sendNullPacket()
			// Restart timer
			c.startNullPacketTimer()
		})
	}
}

// stopNullPacketTimer menghentikan timer untuk mengirim paket null
func (c *dummyPacketConn) stopNullPacketTimer() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.stopped = true
	if c.timer != nil {
		c.timer.Stop()
	}
}

// Close menutup koneksi dan menghentikan timer
func (c *dummyPacketConn) Close() error {
	c.stopNullPacketTimer()
	return c.Conn.Close()
}

// sendDummyPacket mengirim paket dummy ke koneksi untuk mencegah koneksi terputus
func sendDummyPacket(conn net.Conn) error {
	// Paket dummy sederhana
	dummyData := []byte("PING\r\n")
	_, err := conn.Write(dummyData)
	return err
}
