# s2http

Mengubah proxy socks menjadi proxy http(s)

# Prasyarat (jika menggunakan socks)

Proxy socks (kemungkinan client) sedang berjalan

Aplikasi ini mungkin tidak langsung terhubung ke server socks

## Penggunaan

```
Usage of ./s2http:
  -dummy-packet
    	kirim paket dummy untuk mencegah kesalahan reset koneksi (default true)
  -dummy-packet-interval int
    	interval dalam detik untuk mengirim paket dummy (default 30)
  -dummy-packet-size int
    	ukuran paket dummy dalam byte (default 1)
  -error-handling string
    	strategi penanganan kesalahan: retry, ignore, atau fatal (default "retry")
  -keepalive
    	aktifkan koneksi keep-alive (default true)
  -keepalive-timeout int
    	timeout keep-alive dalam detik (default 60)
  -max-idle-conns int
    	maksimum koneksi idle (default 100)
  -min-packet-interval int
    	interval minimum dalam detik untuk mengirim paket dummy (default 5)
  -port string
    	port untuk listening (default "8080")
  -pure
    	proxy http/https murni tanpa socks
  -read-timeout int
    	timeout baca dalam detik (default 5)
  -socks string
    	url socks (default) (default "127.0.0.1:1081")
  -v	verbose (default true)
  -version
    	tampilkan versi.
  -write-timeout int
    	timeout tulis dalam detik (default 5)
```

## Fitur

### Pengiriman Paket Dummy

Aplikasi dapat mengirim paket dummy (byte null) untuk menjaga koneksi tetap aktif dan mencegah kesalahan reset koneksi seperti "broken pipe", "connection reset by peer", dan "network is unreachable".

- `-dummy-packet`: Aktifkan/nonaktifkan pengiriman paket dummy (default: true)
- `-dummy-packet-size`: Ukuran paket dummy dalam byte (default: 1)
- `-dummy-packet-interval`: Interval dalam detik untuk mengirim paket dummy (default: 30)
- `-min-packet-interval`: Interval minimum dalam detik untuk mengirim paket dummy (default: 5)

### Penanganan Kesalahan

Strategi penanganan kesalahan yang fleksibel untuk kesalahan jaringan:

- `-error-handling`: Strategi untuk menangani kesalahan (retry, ignore, atau fatal) (default: "retry")

### Manajemen Koneksi

Fitur manajemen koneksi lanjutan:

- `-keepalive`: Aktifkan/nonaktifkan koneksi keep-alive (default: true)
- `-keepalive-timeout`: Timeout keep-alive dalam detik (default: 60)
- `-max-idle-conns`: Maksimum koneksi idle (default: 100)

### Konfigurasi Timeout

Konfigurasi timeout terpisah untuk operasi baca dan tulis:

- `-read-timeout`: Timeout baca dalam detik (default: 5)
- `-write-timeout`: Timeout tulis dalam detik (default: 5)

## Proxy murni

Jalankan sebagai proxy murni (tanpa socks)

    ./s2http -pure

## Contoh

### Penggunaan dasar

```
./s2http
```

### Port kustom dan proxy socks

```
./s2http -port 8081 -socks 127.0.0.1:1080
```

### Aktifkan mode verbose

```
./s2http -v
```

### Konfigurasi pengiriman paket dummy

```
./s2http -dummy-packet -dummy-packet-size 10 -dummy-packet-interval 10
```

### Atur strategi penanganan kesalahan

```
./s2http -error-handling ignore
```

### Konfigurasi timeout

```
./s2http -read-timeout 10 -write-timeout 10
```

### Proxy HTTP/HTTPS murni (tanpa SOCKS)

```
./s2http -pure
```

### Contoh konfigurasi lanjutan

```
./s2http -pure -port 8088 -socks 127.0.0.1:1080 -read-timeout=5000 -write-timeout=1000  -keepalive=true -keepalive-timeout=3000 -max-idle-conns=500 -dummy-packet=true -v -dummy-packet-size=128 -dummy-packet-interval=3
```

## Kredit

Terinspirasi oleh [GopherFromHell](https://www.reddit.com/user/GopherFromHell),
Karena dia menyediakan [contoh](https://play.golang.org/p/l0iLtkD1DV)

akhir.
