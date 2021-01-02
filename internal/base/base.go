// Package base ...
package base

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unsafe"
)

var (
	seedInt64    = int64(10000)
	base64String = "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789" +
		"+/"
)

// IsNil ...
func IsNil(val interface{}) (ret bool) {
	defer func() {
		if e := recover(); e != nil {
			ret = false
		}
	}()

	if val == nil {
		return true
	}

	return reflect.ValueOf(val).IsNil()
}

// MinInt ...
func MinInt(v1 int, v2 int) int {
	if v1 < v2 {
		return v1
	}

	return v2
}

// MaxInt ...
func MaxInt(v1 int, v2 int) int {
	if v1 < v2 {
		return v2
	}

	return v1
}

// StringToBytesUnsafe ...
func StringToBytesUnsafe(s string) (ret []byte) {
	bytesHeader := (*reflect.SliceHeader)(unsafe.Pointer(&ret))
	stringHeader := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bytesHeader.Len = stringHeader.Len
	bytesHeader.Cap = stringHeader.Len
	bytesHeader.Data = stringHeader.Data
	return
}

// BytesToStringUnsafe ...
func BytesToStringUnsafe(bytes []byte) (ret string) {
	bytesHeader := (*reflect.SliceHeader)(unsafe.Pointer(&bytes))
	stringHeader := (*reflect.StringHeader)(unsafe.Pointer(&ret))
	stringHeader.Len = bytesHeader.Len
	stringHeader.Data = bytesHeader.Data
	return
}

// IsUTF8Bytes ...
func IsUTF8Bytes(bytes []byte) bool {
	idx := 0
	length := len(bytes)

	for idx < length {
		c := bytes[idx]
		if c < 128 {
			idx++
		} else if c < 224 {
			if (idx+2 > length) ||
				(bytes[idx+1]&0xC0 != 0x80) {
				return false
			}
			idx += 2
		} else if c < 240 {
			if (idx+3 > length) ||
				(bytes[idx+1]&0xC0 != 0x80) ||
				(bytes[idx+2]&0xC0 != 0x80) {
				return false
			}
			idx += 3
		} else if c < 248 {
			if (idx+4 > length) ||
				(bytes[idx+1]&0xC0 != 0x80) ||
				(bytes[idx+2]&0xC0 != 0x80) ||
				(bytes[idx+3]&0xC0 != 0x80) {
				return false
			}
			idx += 4
		} else {
			return false
		}
	}

	return idx == length
}

// GetSeed get int64 seed, it is goroutine safety
func GetSeed() int64 {
	return atomic.AddInt64(&seedInt64, 1)
}

// GetRandString get random string
func GetRandString(strLen int) string {
	sb := NewStringBuilder()
	defer sb.Release()

	for strLen > 0 {
		rand64 := rand.Uint64()
		for used := 0; used < 10 && strLen > 0; used++ {
			sb.AppendByte(base64String[rand64%64])
			rand64 = rand64 / 64
			strLen--
		}
	}
	return sb.String()
}

// AddPrefixPerLine ...
func AddPrefixPerLine(text string, prefix string) string {
	sb := NewStringBuilder()
	defer sb.Release()

	first := true
	array := strings.Split(text, "\n")
	for idx, v := range array {
		if first {
			first = false
		} else {
			sb.AppendByte('\n')
		}

		if v != "" || idx == 0 || idx != len(array)-1 {
			sb.AppendString(prefix)
			sb.AppendString(v)
		}
	}
	return sb.String()
}

// ConcatString ...
func ConcatString(args ...string) string {
	sb := NewStringBuilder()
	defer sb.Release()
	for _, v := range args {
		sb.AppendString(v)
	}
	return sb.String()
}

// GetFileLine ...
func GetFileLine(skip uint) string {
	return AddFileLine("", skip+1)
}

// AddFileLine ...
func AddFileLine(header string, skip uint) string {
	sb := NewStringBuilder()
	defer sb.Release()

	if _, file, line, ok := runtime.Caller(int(skip) + 1); ok && line > 0 {
		if header != "" {
			sb.AppendString(header)
			sb.AppendByte(' ')
		}

		sb.AppendString(file)
		sb.AppendByte(':')
		sb.AppendString(strconv.Itoa(line))
	} else {
		sb.AppendString(header)
	}

	return sb.String()
}

// ConvertOrdinalToString ...
func ConvertOrdinalToString(n uint) string {
	if n == 0 {
		return ""
	}

	switch n {
	case 1:
		return "1st"
	case 2:
		return "2nd"
	case 3:
		return "3rd"
	default:
		return strconv.Itoa(int(n)) + "th"
	}
}

// IsTCPPortOccupied ...
func IsTCPPortOccupied(port uint16) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 0)
	defer func() {
		if conn != nil {
			_ = conn.Close()
		}
	}()
	return err == nil
}

// ReadFromFile ...
func ReadFromFile(filePath string) (string, error) {
	ret, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	// for windows, remove \r
	return strings.Replace(string(ret), "\r", "", -1), nil
}

// WaitAtLeastDurationWhenRunning ...
func WaitAtLeastDurationWhenRunning(
	startNS int64,
	isRunning func() bool,
	duration time.Duration,
) {
	sleepTime := 100 * time.Millisecond
	runNS := TimeNow().UnixNano() - startNS
	sleepCount := (duration - time.Duration(runNS) + sleepTime/2) / sleepTime
	for isRunning() && sleepCount > 0 {
		time.Sleep(sleepTime)
		sleepCount--
	}
}

// GetTLSServerConfig ...
func GetTLSServerConfig(certFile string, keyFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)

	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		// Causes servers to use Go's default ciphersuite preferences,
		// which are tuned to avoid attacks. Does nothing on clients.
		PreferServerCipherSuites: true,
		// Only use curves which have assembly implementations
		CurvePreferences: []tls.CurveID{
			tls.CurveP256,
			tls.X25519, // Go 1.8 only
		},
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305, // Go 1.8 only
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,   // Go 1.8 only
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,

			// Best disabled, as they don't provide Forward Secrecy,
			// but might be necessary for some clients
			// tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			// tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
		},
	}, nil
}

// GetTLSClientConfig ...
func GetTLSClientConfig(
	verifyServerCert bool,
	caFiles []string,
) (*tls.Config, error) {
	caPool := (*x509.CertPool)(nil)

	if len(caFiles) > 0 {
		caPool = x509.NewCertPool()

		for _, caFile := range caFiles {
			if caCert, e := ioutil.ReadFile(caFile); e == nil {
				caPool.AppendCertsFromPEM(caCert)
			} else {
				return nil, e
			}
		}
	}

	return &tls.Config{
		InsecureSkipVerify: !verifyServerCert,
		RootCAs:            caPool,
	}, nil
}
