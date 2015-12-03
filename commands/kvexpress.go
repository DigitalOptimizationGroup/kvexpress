package commands

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"
)

func init() {
	// Nothing happens here.
}

// LengthCheck makes sure a string has at least minLength lines.
func LengthCheck(data string, minLength int) bool {
	length := LineCount(data)
	Log(fmt.Sprintf("length='%d' minLength='%d'", length, minLength), "debug")
	if length >= minLength {
		return true
	}
	return false
}

// ReadURL grabs a URL and returns the string from the body.
func ReadURL(url string, dogstatsd bool) string {
	resp, err := http.Get(url)
	if err != nil {
		Log(fmt.Sprintf("function='ReadURL' panic='true' url='%s'", url), "info")
		if dogstatsd {
			StatsdPanic(url, "read_url")
		}
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return string(body)
}

// LineCount splits a string by linebreak and returns the number of lines.
func LineCount(data string) int {
	var length int
	if strings.ContainsAny(data, "\n") {
		length = strings.Count(data, "\n")
	} else {
		length = 1
	}
	return length
}

// ComputeChecksum takes a string and computes a SHA256 checksum.
func ComputeChecksum(data string) string {
	dataBytes := []byte(data)
	computedChecksum := sha256.Sum256(dataBytes)
	finalChecksum := fmt.Sprintf("%x\n", computedChecksum)
	Log(fmt.Sprintf("computedChecksum='%s'", finalChecksum), "debug")
	return finalChecksum
}

// ChecksumCompare takes a string, generates a SHA256 checksum and compares
// against the passed checksum to see if they match.
func ChecksumCompare(data string, checksum string) bool {
	computedChecksum := ComputeChecksum(data)
	Log(fmt.Sprintf("checksum='%s' computedChecksum='%s'", checksum, computedChecksum), "debug")
	if strings.TrimSpace(computedChecksum) == strings.TrimSpace(checksum) {
		return true
	}
	return false
}

// UnixDiff runs diff to generate text for the Datadog events.
func UnixDiff(old, new string) string {
	diff, _ := exec.Command("diff", "-u", old, new).Output()
	text := string(diff)
	finalText := removeLines(text, 3)
	return finalText
}

// removeLines trims the top n number of lines from a string.
func removeLines(text string, number int) string {
	lines := strings.Split(text, "\n")
	var cleaned []string
	cleaned = append(cleaned, lines[number:]...)
	finalText := strings.Join(cleaned, "\n")
	return finalText
}

// RunCommand runs a cli command with arguments.
func RunCommand(command string) bool {
	parts := strings.Fields(command)
	cli := parts[0]
	args := parts[1:len(parts)]
	cmd := exec.Command(cli, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		Log(fmt.Sprintf("exec='error' message='%v'", err), "info")
		return false
	}
	return true
}
