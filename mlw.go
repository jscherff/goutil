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

package goutil

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

	syslogs struct {
		System *srslog.Writer
		Access *srslog.Writer
		Error *srslog.Writer
	}

	Options struct {

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

		UseFlags struct {
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
		}

		RecoveryStack bool
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

func NewMultiLoggerWriter(cf ...string) *MultiLoggerWriter {

	this := &MultiLoggerWriter{}

	if len(cf) == 0 {
		return this
	}

	if fh, err := os.Open(cf[0]); err == nil {
		defer fh.Close()
		jd := json.NewDecoder(fh)
		err = jd.Decode(&this)
	} else {
		log.Printf("Error opening %q. Using default object.", cf[0])
	}

	return this
}

func (this *MultiLoggerWriter) Init() *MultiLoggerWriter {

	var (
		sw, aw, ew []io.Writer

		lFlags int

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


	if this.Options.LogFiles.System {
		if f, err := newfl(this.Config.LogFiles.System); err == nil {
			sw = append(sw, f)
		}
	}

	if this.Options.LogFiles.Access {
		if f, err := newfl(this.Config.LogFiles.Access); err == nil {
			aw = append(aw, f)
		}
	}

	if this.Options.LogFiles.Error {
		if f, err := newfl(this.Config.LogFiles.Error); err == nil {
			ew = append(ew, f)
		}
	}

	if this.Options.Console.System {
		sw = append(sw, os.Stdout)
	}

	if this.Options.Console.Access {
		aw = append(aw, os.Stdout)
	}

	if this.Options.Console.Error {
		ew = append(ew, os.Stderr)
	}

	if this.Options.Syslog.System {
		if s, err := newsl(SyslogPriInfo); err == nil {
			sw = append(sw, s)
		}
	}

	if this.Options.Syslog.Access {
		if s, err := newsl(SyslogPriInfo); err == nil {
			aw = append(aw, s)
		}
	}

	if this.Options.Syslog.Error {
		if s, err := newsl(SyslogPriErr); err == nil {
			ew = append(ew, s)
		}
	}

	if len(sw) == 0 {
		sw = append(sw, ioutil.Discard)
	}

	if len(aw) == 0 {
		aw = append(aw, ioutil.Discard)
	}

	if len(ew) == 0 {
		ew = append(ew, ioutil.Discard)
	}

	// Configure log flag options.

	this.Config.LogFlags.System = 0
	this.Config.LogFlags.Access = 0
	this.Config.LogFlags.Error = 0

	if this.Options.LogFlags.Standard {
		lFlags = log.LstdFlags
	}

	if this.Options.LogFlags.UTC {
		lFlags |= log.LUTC
	}

	if this.Options.LogFlags.Date {
		lFlags |= log.Ldate
	}

	if this.Options.LogFlags.Time {
		lFlags |= log.Ltime
	}

	if this.Options.LogFlags.ShortFile {
		lFlags |= log.Lshortfile
	}

	if this.Options.LogFlags.LongFile {
		lFlags |= log.Llongfile
	}

	if this.Options.UseFlags.System {
		this.Config.LogFlags.System = lFlags
	}

	if this.Options.UseFlags.Access {
		this.Config.LogFlags.Access = lFlags
	}

	if this.Options.UseFlags.Error {
		this.Config.LogFlags.Error = lFlags
	}

	// Create io.Writers

	this.writers.System = io.MultiWriter(sw...)
	this.writers.Access = io.MultiWriter(aw...)
	this.writers.Error = io.MultiWriter(ew...)

	// Create bufio.Writers

	this.bufWriters.System = bufio.NewWriter(this.writers.System)
	this.bufWriters.Access = bufio.NewWriter(this.writers.Access)
	this.bufWriters.Error = bufio.NewWriter(this.writers.Error)

	// Create log.Loggers

	this.Config.LogTags.System = strings.TrimSpace(this.Config.LogTags.System) + ` `
	this.Config.LogTags.Access = strings.TrimSpace(this.Config.LogTags.Access) + ` `
	this.Config.LogTags.Error = strings.TrimSpace(this.Config.LogTags.Error) + ` `

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

	return this
}

func (this *MultiLoggerWriter) GetConfig() (b []byte, err error) {

	return json.MarshalIndent(this, "", "\t")
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

func (this *MultiLoggerWriter) GetSystemWriter() io.Writer {
	return this.writers.System
}

func (this *MultiLoggerWriter) GetAccessWriter() io.Writer {
	return this.writers.Access
}

func (this *MultiLoggerWriter) GetErrorWriter() io.Writer {
	return this.writers.Error
}

// Getters for BufWriters.

func (this *MultiLoggerWriter) GetSystemBufWriter() io.Writer {
	return this.bufWriters.System
}

func (this *MultiLoggerWriter) GetAccessBufWriter() io.Writer {
	return this.bufWriters.Access
}

func (this *MultiLoggerWriter) GetErrorBufWriter() io.Writer {
	return this.bufWriters.Error
}

// Getters for Loggers.

func (this *MultiLoggerWriter) GetSystemLogger() *log.Logger {
	return this.loggers.System
}

func (this *MultiLoggerWriter) GetAccessLogger() *log.Logger {
	return this.loggers.Access
}

func (this *MultiLoggerWriter) GetErrorLogger() *log.Logger {
	return this.loggers.Error
}

// Setters.

func (this *MultiLoggerWriter) EnableSystem(b bool) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration is locked`)}
	this.Options.LogFiles.System = b
	this.Options.Console.System = b
	this.Options.Syslog.System = b
	return this
}

func (this *MultiLoggerWriter) EnableAccess(b bool) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration is locked`)}
	this.Options.LogFiles.Access = b
	this.Options.Console.Access = b
	this.Options.Syslog.Access = b
	return this
}

func (this *MultiLoggerWriter) EnableError(b bool) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration is locked`)}
	this.Options.LogFiles.Error = b
	this.Options.Console.Error = b
	this.Options.Syslog.Error = b
	return this
}

func (this *MultiLoggerWriter) EnableLogFiles(b bool) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration is locked`)}
	this.Options.LogFiles.System = b
	this.Options.LogFiles.Access = b
	this.Options.LogFiles.Error = b
	return this
}

func (this *MultiLoggerWriter) EnableConsole(b bool) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration is locked`)}
	this.Options.Console.System = b
	this.Options.Console.Access = b
	this.Options.Console.Error = b
	return this
}

func (this *MultiLoggerWriter) EnableSyslog(b bool) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration is locked`)}
	this.Options.Syslog.System = b
	this.Options.Syslog.Access = b
	this.Options.Syslog.Error = b
	return this
}

func (this *MultiLoggerWriter) SystemUseFlags(b bool) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration is locked`)}
	this.Options.UseFlags.System = b
	return this
}

func (this *MultiLoggerWriter) AccessUseFlags(b bool) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration is locked`)}
	this.Options.UseFlags.Access = b
	return this
}

func (this *MultiLoggerWriter) ErrorUseFlags(b bool) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration is locked`)}
	this.Options.UseFlags.Error = b
	return this
}

func (this *MultiLoggerWriter) FlagsUTC(b bool) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration is locked`)}
	this.Options.LogFlags.UTC = b
	return this
}

func (this *MultiLoggerWriter) FlagsDate(b bool) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration is locked`)}
	this.Options.LogFlags.Date = b
	return this
}

func (this *MultiLoggerWriter) FlagsTime(b bool) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration is locked`)}
	this.Options.LogFlags.Time = b
	return this
}

func (this *MultiLoggerWriter) FlagsLongFile(b bool) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration is locked`)}
	if b {this.Options.LogFlags.ShortFile = false}
	this.Options.LogFlags.LongFile = b
	return this
}

func (this *MultiLoggerWriter) FlagsShortFile(b bool) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration is locked`)}
	if b {this.Options.LogFlags.LongFile = false}
	this.Options.LogFlags.ShortFile = b
	return this
}

func (this *MultiLoggerWriter) FlagsStandard(b bool) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration is locked`)}
	this.Options.LogFlags.Standard = b
	return this
}

func (this *MultiLoggerWriter) RecoveryStack(b bool) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration is locked`)}
	this.Options.RecoveryStack = b
	return this
}

func (this *MultiLoggerWriter) AppName(s string) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration is locked`)}
	this.Config.AppName = s
	return this
}

func (this *MultiLoggerWriter) AppDir(s string) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration is locked`)}
	this.Config.AppDir = s
	return this
}

func (this *MultiLoggerWriter) LogDir(s string) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration is locked`)}
	this.Config.LogDir = s
	return this
}

func (this *MultiLoggerWriter) SystemLog(s string) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration is locked`)}
	this.Config.LogFiles.System = s
	return this
}

func (this *MultiLoggerWriter) AccessLog(s string) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration is locked`)}
	this.Config.LogFiles.Access = s
	return this
}

func (this *MultiLoggerWriter) ErrorLog(s string) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration is locked`)}
	this.Config.LogFiles.Error = s
	return this
}

func (this *MultiLoggerWriter) SyslogProt(s string) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration is locked`)}
	this.Config.Syslog.Prot = s
	return this
}

func (this *MultiLoggerWriter) SyslogHost(s string) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration is locked`)}
	this.Config.Syslog.Host = s
	return this
}

func (this *MultiLoggerWriter) SyslogPort(s string) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration is locked`)}
	this.Config.Syslog.Port = s
	return this
}

func (this *MultiLoggerWriter) SyslogTag(s string) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration is locked`)}
	this.Config.Syslog.Tag = s
	return this
}

func (this *MultiLoggerWriter) SystemTag(s string) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration is locked`)}
	this.Config.LogTags.System = s
	return this
}

func (this *MultiLoggerWriter) AccessTag(s string) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration is locked`)}
	this.Config.LogTags.Access = s
	return this
}

func (this *MultiLoggerWriter) ErrorTag(s string) *MultiLoggerWriter {
	if this.isLocked {panic(`configuration is locked`)}
	this.Config.LogTags.Error = s
	return this
}

func (this *MultiLoggerWriter) Defaults() *MultiLoggerWriter {

	if this.isLocked {panic(`configuration is locked`)}

	return this.

		EnableSystem(true).
		EnableAccess(false).
		EnableError(true).

		SystemUseFlags(true).
		AccessUseFlags(false).
		ErrorUseFlags(true).

		EnableConsole(false).
		EnableSyslog(false).

		FlagsUTC(false).
		FlagsDate(false).
		FlagsTime(false).
		FlagsLongFile(false).
		FlagsShortFile(true).
		FlagsStandard(true).

		RecoveryStack(false).

		AppName(``).

		AppDir(``).
		LogDir(`log`).

		SystemLog(`system.log`).
		AccessLog(`access.log`).
		ErrorLog(`error.log`).

		SyslogProt(``).
		SyslogHost(``).
		SyslogPort(``).
		SyslogTag(``).

		SystemTag(`system`).
		AccessTag(`access`).
		ErrorTag(`error`)
}

func (this *MultiLoggerWriter) DefaultsInit() *MultiLoggerWriter {
	if this.isLocked {panic(`configuration is locked`)}
	return this.Defaults().Init()
}

/* Sample configuration file with defaults:

{
	"Options": {
		"LogFiles": {
			"System": true,
			"Access": false,
			"Error": true
		},
		"Console": {
			"System": false,
			"Access": false,
			"Error": false
		},
		"Syslog": {
			"System": false,
			"Access": false,
			"Error": false
		},
		"LogFlags": {
			"UTC": false,
			"Date": false,
			"Time": false,
			"LongFile": false,
			"ShortFile": false,
			"Standard": true
		},
		"UseFlags": {
			"System": true,
			"Access": false,
			"Error": true
		},
		"RecoveryStack": false
	},
	"Config": {
		"AppName": "",
		"AppDir": "",
		"LogDir": "",
		"LogFiles": {
			"System": "system.log",
			"Access": "access.log",
			"Error": "error.log"
		},
		"LogFlags": {
			"System": 3,
			"Access": 0,
			"Error": 3
		},
		"LogTags": {
			"System": "system",
			"Access": "access",
			"Error": "error"
		},
		"Syslog": {
			"Prot": "",
			"Host": "",
			"Port": "",
			"Tag": ""
		}
	}
}
*/
