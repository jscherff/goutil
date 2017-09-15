// Copyright 2017 John Scherff
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package goutils

import (
	`bufio`
	`encoding/json`
	`log`
	`io`
	`io/ioutil`
	`os`
	`path/filepath`
	`strings`
	`github.com/RackSec/srslog`
)

type MultiLoggerWriter struct {

	logFlags int
	isLocked bool

	loggers struct {
		System *log.Logger
		Access *log.Logger
		Error *log.Logger
	}

	writers struct {
		System io.Writer
		Access io.Writer
		Error io.Writer
	}

	bufWriters struct {
		System io.Writer
		Access io.Writer
		Error io.Writer
	}

	Options struct {

		RecoveryStack bool

		LogFiles struct {
			System bool
			Access bool
			Error bool
		}

		Console struct {
			System bool
			Access bool
			Error bool
		}

		Syslog struct {
			System bool
			Access bool
			Error bool
		}

		LogFlags struct {
			UTC bool
			Date bool
			Time bool
			LongFile bool
			ShortFile bool
			Standard bool
			System bool
			Access bool
			Error bool
		}
	}

	Config struct {

		AppName string

		AppDir string
		LogDir string

		LogFiles struct {
			System string
			Access string
			Error string
		}

		LogFlags struct {
			System int
			Access int
			Error int
		}

		LogTags struct {
			System string
			Access string
			Error string
		}

		Syslog struct {
			Prot string
			Host string
			Port string
			Tag string
		}
	}
}

func NewMultiLoggerWriter(cf ...string) (this *MultiLoggerWriter, err error) {

	if len(cf) != 0 {

		fh, err := os.Open(cf[0])

		if err != nil {
			return nil, err
		}

		defer fh.Close()

		this = &MultiLoggerWriter{}
		jd := json.NewDecoder(fh)
		err = jd.Decode(&this)

	} else {

		return &MultiLoggerWriter{}, nil
	}

	return this, err
}

func (this *MultiLoggerWriter) Init() {

	var (
		sw, aw, ew []io.Writer

		slProt = this.Config.Syslog.Prot
		slHost = this.Config.Syslog.Host
		slPort = this.Config.Syslog.Port
		slTag = this.Config.Syslog.Tag
		slRaddr = strings.Join([]string{slHost, slPort}, `:`)
	)

	this.isLocked = true
	this.Config.AppDir = filepath.Dir(os.Args[0])

	if len(this.Config.LogDir) == 0 {
		this.Config.LogDir = `log`
	}

	if filepath.Dir(this.Config.LogDir) == `.` {
		this.Config.LogDir = filepath.Join(this.Config.AppDir, this.Config.LogDir)
	}

	var newfl = func(f string) (h *os.File, err error) {

		if h, err = os.OpenFile(f, FileFlags, FileMode); err != nil {
			log.Printf(`%v`, ErrorDecorator(err))
		}

		return h, err
	}

	var newsl = func(p srslog.Priority) (s *srslog.Writer, err error) {

		if s, err = srslog.Dial(slProt, slRaddr, p, slTag); err != nil {
			log.Printf(`%v`, ErrorDecorator(err))
		}

		return s, err
	}

	switch true {

	case this.Options.LogFiles.System:
		if f, err := newfl(this.Config.LogFiles.System); err == nil {
			sw = append(sw, f)
		}
		fallthrough

	case this.Options.LogFiles.Access:
		if f, err := newfl(this.Config.LogFiles.Access); err == nil {
			aw = append(aw, f)
		}
		fallthrough

	case this.Options.LogFiles.Error:
		if f, err := newfl(this.Config.LogFiles.Error); err == nil {
			ew = append(ew, f)
		}
		fallthrough

	case this.Options.Console.System:
		sw = append(sw, os.Stdout)
		fallthrough

	case this.Options.Console.Access:
		aw = append(aw, os.Stdout)
		fallthrough

	case this.Options.Console.Error:
		ew = append(ew, os.Stderr)
		fallthrough

	case this.Options.Syslog.System:
		if s, err := newsl(SyslogPriInfo); err == nil {
			sw = append(sw, s)
		}
		fallthrough

	case this.Options.Syslog.Access:
		if s, err := newsl(SyslogPriInfo); err == nil {
			aw = append(aw, s)
		}
		fallthrough

	case this.Options.Syslog.Error:
		if s, err := newsl(SyslogPriErr); err == nil {
			ew = append(ew, s)
		}

	case len(sw) == 0:
		sw = append(sw, ioutil.Discard)
		fallthrough

	case len(aw) == 0:
		aw = append(aw, ioutil.Discard)
		fallthrough

	case len(ew) == 0:
		ew = append(ew, ioutil.Discard)
	}

	// Configure log flag options.

	switch {

	case this.Options.LogFlags.Standard:
		this.logFlags |= log.LstdFlags
		break

	case this.Options.LogFlags.UTC:
		this.logFlags |= log.LUTC
		fallthrough

	case this.Options.LogFlags.Date:
		this.logFlags |= log.Ldate
		fallthrough

	case this.Options.LogFlags.Time:
		this.logFlags |= log.Ltime
		fallthrough

	case this.Options.LogFlags.ShortFile:
		this.logFlags |= log.Lshortfile
		break

	case this.Options.LogFlags.LongFile:
		this.logFlags |= log.Llongfile

	}

	this.Config.LogFlags.System = this.logFlags
	this.Config.LogFlags.Access = this.logFlags
	this.Config.LogFlags.Error = this.logFlags

	switch {

	case !this.Options.LogFlags.System:
		this.Config.LogFlags.System = 0
		fallthrough

	case !this.Options.LogFlags.Access:
		this.Config.LogFlags.Access= 0
		fallthrough

	case !this.Options.LogFlags.Error:
		this.Config.LogFlags.Error = 0
	}

	this.writers.System = io.MultiWriter(sw...)
	this.writers.Access = io.MultiWriter(aw...)
	this.writers.Error = io.MultiWriter(ew...)

	this.bufWriters.System = bufio.NewWriter(this.writers.System)
	this.bufWriters.Access = bufio.NewWriter(this.writers.Access)
	this.bufWriters.Error = bufio.NewWriter(this.writers.Error)

	this.loggers.System = log.New(
		this.writers.System,
		this.Config.LogTags.System,
		this.Config.LogFlags.System,
	)

	this.loggers.Access = log.New(
		this.writers.Access,
		this.Config.LogTags.Access,
		this.Config.LogFlags.Access,
	)

	this.loggers.Error = log.New(
		this.writers.Error,
		this.Config.LogTags.Error,
		this.Config.LogFlags.Error,
	)
}

func (this *MultiLoggerWriter) SaveConfig(cf string) (err error) {

	fh, err := os.Create(cf)

	if err != nil {
		return err
	}

	defer fh.Close()

	je := json.NewEncoder(fh)
	je.SetIndent("", "\t")
	err = je.Encode(&this)

	return err
}

// Getters for Writers.

func (this *MultiLoggerWriter) SystemWriter() io.Writer {
	return this.writers.System
}

func (this *MultiLoggerWriter) AccessWriter() io.Writer {
	return this.writers.Access
}

func (this *MultiLoggerWriter) ErrorWriter() io.Writer {
	return this.writers.Error
}

// Getters for BufWriters.

func (this *MultiLoggerWriter) SystemBufWriter() io.Writer {
	return this.bufWriters.System
}

func (this *MultiLoggerWriter) AccessBufWriter() io.Writer {
	return this.bufWriters.Access
}

func (this *MultiLoggerWriter) ErrorBufWriter() io.Writer {
	return this.bufWriters.Error
}

// Getters for Loggers.

func (this *MultiLoggerWriter) SystemLogger() *log.Logger {
	return this.loggers.System
}

func (this *MultiLoggerWriter) AccessLogger() *log.Logger {
	return this.loggers.Access
}

func (this *MultiLoggerWriter) ErrorLogger() *log.Logger {
	return this.loggers.Error
}

// Setters.

func (this *MultiLoggerWriter) EnableConsole(b bool) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration isLocked`)}
	this.Options.Console.System = b
	this.Options.Console.Access = b
	this.Options.Console.Error = b
	return this
}

func (this *MultiLoggerWriter) EnableLogFiles (b bool) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration isLocked`)}
	this.Options.LogFiles.System = b
	this.Options.LogFiles.Access = b
	this.Options.LogFiles.Error = b
	return this
}

func (this *MultiLoggerWriter) EnableSyslog (b bool) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration isLocked`)}
	this.Options.Syslog.System = b
	this.Options.Syslog.Access = b
	this.Options.Syslog.Error = b
	return this
}

func (this *MultiLoggerWriter) EnableSystem (b bool) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration isLocked`)}
	this.Options.LogFiles.System = b
	this.Options.Console.System = b
	this.Options.Syslog.System = b
	return this
}

func (this *MultiLoggerWriter) EnableAccess (b bool) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration isLocked`)}
	this.Options.LogFiles.Access = b
	this.Options.Console.Access = b
	this.Options.Syslog.Access = b
	return this
}

func (this *MultiLoggerWriter) EnableError (b bool) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration isLocked`)}
	this.Options.LogFiles.Error = b
	this.Options.Console.Error = b
	this.Options.Syslog.Error = b
	return this
}

func (this *MultiLoggerWriter) SystemLogFlags (b bool) *MultiLoggerWriter {
	this.Options.LogFlags.System = b
	return this
}

func (this *MultiLoggerWriter) AccessLogFlags (b bool) *MultiLoggerWriter {
	this.Options.LogFlags.Access = b
	return this
}

func (this *MultiLoggerWriter) ErrorLogFlags (b bool) *MultiLoggerWriter {
	this.Options.LogFlags.Error = b
	return this
}

func (this *MultiLoggerWriter) FlagsUTC (b bool) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration isLocked`)}
	this.Options.LogFlags.UTC = b
	return this
}

func (this *MultiLoggerWriter) FlagsDate (b bool) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration isLocked`)}
	this.Options.LogFlags.Date = b
	return this
}

func (this *MultiLoggerWriter) FlagsTime (b bool) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration isLocked`)}
	this.Options.LogFlags.Time = b
	return this
}

func (this *MultiLoggerWriter) FlagsLongFile (b bool) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration isLocked`)}
	if b {this.Options.LogFlags.ShortFile = false}
	this.Options.LogFlags.LongFile = b
	return this
}

func (this *MultiLoggerWriter) FlagsShortFile (b bool) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration isLocked`)}
	if b {this.Options.LogFlags.LongFile = false}
	this.Options.LogFlags.ShortFile = b
	return this
}

func (this *MultiLoggerWriter) FlagsStandard (b bool) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration isLocked`)}
	if b {
		this.Options.LogFlags.UTC = false
		this.Options.LogFlags.Date = false
		this.Options.LogFlags.Time = false
		this.Options.LogFlags.LongFile = false
		this.Options.LogFlags.ShortFile = false
	}
	this.Options.LogFlags.Standard = b
	return this
}

func (this *MultiLoggerWriter) RecoveryStack (b bool) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration isLocked`)}
	this.Options.RecoveryStack = b
	return this
}

func (this *MultiLoggerWriter) LogDir (d string) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration isLocked`)}
	this.Config.LogDir = d
	return this
}

func (this *MultiLoggerWriter) SystemLog (f string) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration isLocked`)}
	this.Config.LogFiles.System = f
	return this
}

func (this *MultiLoggerWriter) AccessLog (f string) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration isLocked`)}
	this.Config.LogFiles.Access = f
	return this
}

func (this *MultiLoggerWriter) ErrorLog (f string) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration isLocked`)}
	this.Config.LogFiles.Error = f
	return this
}

func (this *MultiLoggerWriter) SyslogProt (v string) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration isLocked`)}
	this.Config.Syslog.Prot = v
	return this
}

func (this *MultiLoggerWriter) SyslogHost (v string) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration isLocked`)}
	this.Config.Syslog.Host = v
	return this
}

func (this *MultiLoggerWriter) SyslogPort (v string) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration isLocked`)}
	this.Config.Syslog.Port = v
	return this
}

func (this *MultiLoggerWriter) SyslogTag (v string) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration isLocked`)}
	this.Config.Syslog.Tag = v
	return this
}

func (this *MultiLoggerWriter) SystemTag (v string) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration isLocked`)}
	this.Config.LogTags.System = v
	return this
}

func (this *MultiLoggerWriter) AccessTag (v string) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration isLocked`)}
	this.Config.LogTags.Access = v
	return this
}

func (this *MultiLoggerWriter) ErrorTag (v string) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration isLocked`)}
	this.Config.LogTags.System = v
	return this
}
