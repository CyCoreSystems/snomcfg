package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
)

// ErrComment indicates that the line was ignored due to it being commented out
var ErrComment = errors.New("Comment line")

var dir string
var debug bool

var phoneTypeRegex = regexp.MustCompile(`snom\d{3}`)

func main() {
	addr := flag.String("l", ":8080", "Listen address")
	flag.StringVar(&dir, "d", "/etc/asterisk/snom", "Source directory for configuration")
	flag.BoolVar(&debug, "debug", false, "enable debug logging")

	flag.Parse()

	// Set handlers
	http.HandleFunc("/config", config)
	http.HandleFunc("/firmware", firmware)

	log.Fatal(http.ListenAndServe(*addr, nil))
}

func firmware(w http.ResponseWriter, r *http.Request) {
	ptype := getPhoneType(r.UserAgent())
	if ptype == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid phone type"))
		return
	}
	f, err := os.Open(path.Join(dir, fmt.Sprintf("%s-firmware.cfg", ptype)))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Failed to find firmware config"))
		return
	}
	defer f.Close()

	io.Copy(w, f)
	return
}

func config(w http.ResponseWriter, r *http.Request) {
	cfg := make(map[string]string)

	files := []string{"snom-passwd.cfg"}

	// Find the phone type
	ptype := getPhoneType(r.UserAgent())
	if ptype != "" {
		log.Println("Matched phone type", ptype)
		files = append(files, ptype+".cfg")
	}

	// Find the MAC-specific config
	mac := r.FormValue("mac")
	if mac != "" {
		log.Println("Matched MAC address", mac)
		files = append(files, fmt.Sprintf("snom-%s.cfg", mac))
	}

	for _, fn := range files {
		f, err := os.Open(path.Join(dir, fn))
		if err != nil {
			log.Println("Failed to open file", fn, err)
		}
		err = readLines(f, cfg)
		if err != nil {
			log.Println("Failed to read file", fn, err)
		}
		f.Close()
	}

	// Write the configuration to the ResponseWriter
	for k, v := range cfg {
		if debug {
			fmt.Printf("%s: %s\n", k, v)
		}
		fmt.Fprintf(w, "%s: %s\n", k, v)
	}
	return
}

func getPhoneType(ua string) (ret string) {
	return phoneTypeRegex.FindString(ua)
}

func parseLine(ln string) (key string, val string, err error) {
	if strings.HasPrefix(ln, "#") {
		err = ErrComment
		return
	}

	// Split the line by the ':' character
	pieces := strings.SplitN(ln, ":", 2)
	if len(pieces) < 2 {
		err = fmt.Errorf("Failed to parse line: %s", ln)
		return
	}

	return strings.TrimSpace(pieces[0]), strings.TrimSpace(pieces[1]), nil
}

func readLines(in io.Reader, out map[string]string) error {
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		k, v, err := parseLine(scanner.Text())
		if err == nil {
			out[k] = v
			continue
		}
		if err != ErrComment {
			return err
		}
	}
	return nil
}
