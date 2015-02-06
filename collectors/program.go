package collectors

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/oliveagle/go-collectors/datapoint"
	// "github.com/oliveagle/go-collectors/metadata"
	"github.com/oliveagle/go-collectors/slog"
	"github.com/oliveagle/go-collectors/util"
)

type ProgramCollector struct {
	Path     string
	Interval time.Duration
}

func InitPrograms(cpath string) {
	cdir, err := os.Open(cpath)
	if err != nil {
		slog.Infoln(err)
		return
	}
	idirs, err := cdir.Readdir(0)
	if err != nil {
		slog.Infoln(err)
		return
	}
	for _, idir := range idirs {
		i, err := strconv.Atoi(idir.Name())
		if err != nil || i < 0 {
			slog.Infoln("invalid collector folder name:", idir.Name())
			continue
		}
		interval := time.Second * time.Duration(i)
		dir, err := os.Open(filepath.Join(cdir.Name(), idir.Name()))
		if err != nil {
			slog.Infoln(err)
			continue
		}
		files, err := dir.Readdir(0)
		if err != nil {
			slog.Infoln(err)
			continue
		}
		for _, file := range files {
			if !isExecutable(file) {
				continue
			}
			collectors = append(collectors, &ProgramCollector{
				Path:     filepath.Join(dir.Name(), file.Name()),
				Interval: interval,
			})
		}
	}
}

func isExecutable(f os.FileInfo) bool {
	switch runtime.GOOS {
	case "windows":
		exts := strings.Split(os.Getenv("PATHEXT"), ";")
		fileExt := filepath.Ext(strings.ToUpper(f.Name()))
		for _, ext := range exts {
			if filepath.Ext(strings.ToUpper(ext)) == fileExt {
				return true
			}
		}
		return false
	default:
		return f.Mode()&0111 != 0
	}
}

func (c *ProgramCollector) Run(dpchan chan<- *datapoint.DataPoint) {
	if c.Interval == 0 {
		for {
			next := time.After(DefaultFreq)
			if err := c.runProgram(dpchan); err != nil {
				slog.Infoln(err)
			}
			<-next
			slog.Infoln("restarting", c.Path)
		}
	} else {
		for {
			next := time.After(c.Interval)
			c.runProgram(dpchan)
			<-next
		}
	}
}

func (c *ProgramCollector) Init() {
}

func (c *ProgramCollector) runProgram(dpchan chan<- *datapoint.DataPoint) (progError error) {
	cmd := exec.Command(c.Path)
	pr, pw := io.Pipe()
	s := bufio.NewScanner(pr)
	cmd.Stdout = pw
	er, ew := io.Pipe()
	cmd.Stderr = ew
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() {
		progError = cmd.Wait()
		pw.Close()
		ew.Close()
	}()
	go func() {
		es := bufio.NewScanner(er)
		for es.Scan() {
			line := strings.TrimSpace(es.Text())
			slog.Error(line)
		}
	}()
Loop:
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		sp := strings.Fields(line)
		if len(sp) < 3 {
			slog.Errorf("bad line in program %s: %s", c.Path, line)
			continue
		}
		ts, err := strconv.ParseInt(sp[1], 10, 64)
		if err != nil {
			slog.Errorf("bad timestamp in program %s: %s", c.Path, sp[1])
			continue
		}
		val, err := strconv.ParseFloat(sp[2], 64)
		if err != nil {
			slog.Errorf("bad value in program %s: %s", c.Path, sp[2])
			continue
		}
		if !datapoint.ValidTag(sp[0]) {
			slog.Errorf("bad metric in program %s: %s", c.Path, sp[0])
		}
		dp := datapoint.DataPoint{
			Metric:    sp[0],
			Timestamp: ts,
			Value:     val,
			Tags:      datapoint.TagSet{"host": util.Hostname},
		}
		for _, tag := range sp[3:] {
			tags, err := datapoint.ParseTags(tag)
			if v, ok := tags["host"]; ok && v == "" {
				delete(dp.Tags, "host")
			} else if err != nil {
				slog.Errorf("bad tag in program %s, metric %s: %v: %v", c.Path, sp[0], tag, err)
				continue Loop
			} else {
				dp.Tags.Merge(tags)
			}
		}
		dp.Tags = AddTags.Copy().Merge(dp.Tags)
		dpchan <- &dp
	}
	if err := s.Err(); err != nil {
		return err
	}
	return
}

func (c *ProgramCollector) Name() string {
	return c.Path
}
