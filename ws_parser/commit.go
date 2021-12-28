package ws_parser

import (
	"time"
	"os"
	"io"
	"log"
	"fmt"
	"encoding/csv"
)

type Committer struct {
	raw_fd *os.File
	headers map[string][]string
	writers map[string]*csv.Writer
}

func NewCommitter() (*Committer) {
	c := &Committer{nil, make(map[string][]string), make(map[string]*csv.Writer)}
	return c
}

func createOutputFilename(content_type string, extension string) string {
	t := time.Now()
	y,m,d := t.Date()
	return fmt.Sprintf("%04d%02d%02d.%s.%s", y, m, d, content_type, extension)
}

func (c *Committer) getRawOutputWriter() (*os.File) {
	if c.raw_fd != nil {
		return c.raw_fd
	}

	fname := createOutputFilename("raw", "json")
	log.Printf("createOutputWriter: output=%v\n", fname)
	fd, err := os.OpenFile(fname, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("cannot open output file=%v: %v", fname, err)
	}

	c.raw_fd = fd
	return c.raw_fd
}

func (c *Committer) commitRawPayload(ts time.Time, msg []byte) {
	fd := c.getRawOutputWriter()
	header, err := ts.MarshalText()
	if err != nil {
		log.Fatal("commitRawPayload: cannot marshal timestamp")
	}
	header = append(header, ' ')
	fd.Write(header)

	msg = append(msg, '\n')
	fd.Write(msg)
}

func (c *Committer) RegisterTable(name string, header []string) {
	c.headers[name] = header
}

func (c *Committer) CommitRecord(ts time.Time, name string, record []string) {
	writer, exists := c.writers[name]
	if !exists {
		fname := createOutputFilename(name, "csv")
		log.Printf("createOutputWriter: name=%v fname=%v\n", name, fname)
		fd, err := os.OpenFile(fname, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("cannot open output file=%v: %v", fname, err)
		}
		writer = csv.NewWriter(fd)
		c.writers[name] = writer
		offset, err := fd.Seek(0, io.SeekCurrent)
		if err != nil {
			log.Fatal("cannot ftell:", err)
		}
		if offset==0 {
			writer.Write(c.headers[name])
		}
	}

	writer.Write(record)
	writer.Flush()
}

func (c *Committer) Close() {
	if c.raw_fd != nil {
		c.raw_fd.Close()
		c.raw_fd = nil
	}

	for _, w := range c.writers {
		w.Flush()
	}

	c.writers = map[string]*csv.Writer{}
}
