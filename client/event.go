/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package client

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
)

// Event headers
var (
	headerID    = []byte("id:")
	headerData  = []byte("data:")
	headerEvent = []byte("event:")
	headerMeta  = []byte("meta:")
)

// Event holds all of the event source fields
type StreamEvent struct {
	ID    []byte
	Data  []byte
	Event []byte
	Meta  []byte
}

func (e *StreamEvent) hasContent() bool {
	return len(e.ID) > 0 || len(e.Data) > 0 || len(e.Event) > 0 || len(e.Meta) > 0
}

// EventStreamReader scans an io.Reader looking for EventStream messages.
type EventStreamReader struct {
	scanner *bufio.Scanner
}

// NewEventStreamReader creates an instance of EventStreamReader.
func NewEventStreamReader(eventStream io.Reader, maxBufferSize int) *EventStreamReader {
	scanner := bufio.NewScanner(eventStream)
	initBufferSize := minPosInt(4096, maxBufferSize)
	scanner.Buffer(make([]byte, initBufferSize), maxBufferSize)
	split := func(data []byte, atEOF bool) (int, []byte, error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		// We have a full event payload to parse.
		if i, nlen := containsDoubleNewline(data); i >= 0 {
			return i + nlen, data[0:i], nil
		}
		// If we're at EOF, we have all of the data.
		if atEOF {
			return len(data), data, nil
		}
		// Request more data.
		return 0, nil, nil
	}
	// Set the split function for the scanning operation.
	scanner.Split(split)
	return &EventStreamReader{
		scanner: scanner,
	}
}

// ReadEvent scans the EventStream for events.
func (e *EventStreamReader) ReadEvent() (*StreamEvent, error) {
	if e.scanner.Scan() {
		eventBody := e.scanner.Bytes()
		event, err := parseEvent(eventBody)
		if err != nil {
			return nil, err
		}
		return event, nil
	}
	if err := e.scanner.Err(); err != nil {
		if err == context.Canceled {
			return nil, io.EOF
		}
		return nil, err
	}
	return nil, io.EOF
}

// Returns a tuple containing the index of a double newline, and the number of bytes
// represented by that sequence. If no double newline is present, the first value
// will be negative.
func containsDoubleNewline(data []byte) (int, int) {
	// Search for each potentially valid sequence of newline characters
	crcr := bytes.Index(data, []byte("\r\r"))
	lflf := bytes.Index(data, []byte("\n\n"))
	crlflf := bytes.Index(data, []byte("\r\n\n"))
	lfcrlf := bytes.Index(data, []byte("\n\r\n"))
	crlfcrlf := bytes.Index(data, []byte("\r\n\r\n"))
	// Find the earliest position of a double newline combination
	minPos := minPosInt(crcr, minPosInt(lflf, minPosInt(crlflf, minPosInt(lfcrlf, crlfcrlf))))
	// Detemine the length of the sequence
	nlen := 2
	if minPos == crlfcrlf {
		nlen = 4
	} else if minPos == crlflf || minPos == lfcrlf {
		nlen = 3
	}
	return minPos, nlen
}

// Returns the minimum non-negative value out of the two values. If both
// are negative, a negative value is returned.
func minPosInt(a, b int) int {
	if a < 0 {
		return b
	}
	if b < 0 {
		return a
	}
	if a > b {
		return b
	}
	return a
}

func parseEvent(msg []byte) (*StreamEvent, error) {
	var event StreamEvent
	if len(msg) < 1 {
		return nil, errors.New("event message was empty")
	}
	// Normalize the crlf to lf to make it easier to split the lines.
	// Split the line by "\n" or "\r", per the spec.
	for _, line := range bytes.FieldsFunc(msg, func(r rune) bool { return r == '\n' || r == '\r' }) {
		switch {
		case bytes.HasPrefix(line, headerID):
			event.ID = append([]byte(nil), trimHeader(len(headerID), line)...)
		case bytes.HasPrefix(line, headerData):
			// The spec allows for multiple data fields per event, concatenated them with "\n".
			event.Data = append(event.Data[:], append(trimHeader(len(headerData), line), byte('\n'))...)
			// The spec says that a line that simply contains the string "data" should be treated as a data field with an empty body.
		case bytes.Equal(line, bytes.TrimSuffix(headerData, []byte(":"))):
			event.Data = append(event.Data, byte('\n'))
		case bytes.HasPrefix(line, headerEvent):
			event.Event = append([]byte(nil), trimHeader(len(headerEvent), line)...)
		case bytes.HasPrefix(line, headerMeta):
			event.Meta = append([]byte(nil), trimHeader(len(headerMeta), line)...)
		default:
			// Ignore any garbage that doesn't match what we're looking for.
		}
	}
	// Trim the last "\n" per the spec.
	event.Data = bytes.TrimSuffix(event.Data, []byte("\n"))
	return &event, nil
}

func trimHeader(size int, data []byte) []byte {
	if data == nil || len(data) < size {
		return data
	}
	data = data[size:]
	// Remove optional leading whitespace
	if len(data) > 0 && data[0] == 32 {
		data = data[1:]
	}
	// Remove trailing new line
	if len(data) > 0 && data[len(data)-1] == 10 {
		data = data[:len(data)-1]
	}
	return data
}
