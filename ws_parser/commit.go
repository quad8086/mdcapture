package ws_parser

import (
	"time"
	"os"
	"io"
	"log"
	"fmt"
	"sort"
	"encoding/csv"
	"path/filepath"
	"github.com/valyala/fasttemplate"
)

type Committer struct {
	raw_fd *os.File
	raw_count int64
	directory string
	headers map[string][]string
	writers map[string]*csv.Writer
	counts map[string]int64
}

func NewCommitter(directory string) (*Committer) {
	if len(directory)>0 {
		tpl := fasttemplate.New(directory, "{", "}")
		t := time.Now()
		y,m,d := t.Date()
		vars := map[string]interface{}{
			"y": fmt.Sprintf("%04d", y),
			"m": fmt.Sprintf("%02d", m),
			"d": fmt.Sprintf("%02d", d),
			"ymd": fmt.Sprintf("%04d%02d%02d", y, m, d),
		}
		directory = tpl.ExecuteString(vars)
		log.Printf("committer: directory=%v\n", directory)
		err := os.MkdirAll(directory, 0755)
		if err != nil {
			log.Panicf("Committer: cannot create directory=%v: %v\n", directory, err)
		}
	}
	c := &Committer{nil, 0, directory, make(map[string][]string), make(map[string]*csv.Writer), make(map[string]int64)}
	return c
}

func (c *Committer) createOutputFilename(content_type string, extension string) string {
	t := time.Now()
	y,m,d := t.Date()
	fname := fmt.Sprintf("%04d%02d%02d.%s.%s", y, m, d, content_type, extension)
	if len(c.directory)>0 {
		fname = filepath.Join(c.directory, fname)
	}

	return fname
}

func (c *Committer) getRawOutputWriter() (*os.File) {
	if c.raw_fd != nil {
		return c.raw_fd
	}

	fname := c.createOutputFilename("raw", "json")
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

	c.raw_count += 1
}

func (c *Committer) RegisterTable(name string, header []string) {
	c.headers[name] = header
	c.counts[name] = 0
}

func (c *Committer) CommitRecord(ts time.Time, name string, record []string) {
	writer, exists := c.writers[name]
	if !exists {
		fname := c.createOutputFilename(name, "csv")
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

	c.counts[name] = c.counts[name] + 1
	writer.Write(record)
	writer.Flush()
}

func (c *Committer) Status() (string) {
	if c.raw_fd != nil {
		offset, _ := c.raw_fd.Seek(0, io.SeekCurrent)
		return fmt.Sprintf("raw_count=%v raw_offset=%v", c.raw_count, offset)
	}

	var keys []string
	for k,_ := range c.counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var res string
	for _,k := range keys {
		res += fmt.Sprintf("%v=%v ", k, c.counts[k])
	}
	return res
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
