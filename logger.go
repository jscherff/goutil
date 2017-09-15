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
	`log`
	`io`
	`io/ioutil`
	`os`
	`strings`
	`github.com/RackSec/srslog`
	`github.com/jscherff/goutils`
)

type MultiLogger struct {

	logFlags int
	isLocked bool

	loggers struct {
		System *log.Logger
		Access *log.Logger
		Error *log.Logger
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

		Syslog struct {
			Prot string
			Host string
			Port string
			Tag string
		}

		Tags struct {
			System string
			Access string
			Error string
		}
	}
}

func NewMultiLogger(cf ...string) (this *MultiLogger, err error) {

	if len(cf) != 0 {

		fh, err := os.Open(cf[0])

		if err != nil {
			return nil, err
		}

		defer fh.Close()

		this = &MultiLogger{}
		jd := json.NewDecoder(fh)
		err = jd.Decode(&this)

	} else {

		return &MultiLogger{}, nil
	}

	return this, err
}

func (this *MultiLogger) Init() (err error) {

	this.isLocked = true

	var (
		sw, aw, ew []io.Writer
		slProt = this.Config.Syslog.Prot
		slHost = this.Config.Syslog.Host
		slPort = this.Config.Syslog.Port
		slTag = this.Config.Syslog.Tag
		slRaddr = strings.Join([]string{slAddr, slPort}, `:`)
	)

	var newfl = func(f string) (h *os.File, err error) {

		if h, err = os.OpenFile(f, FileFlags, FileMode); err != nil {
			log.Printf(`%v`, goutils.ErrorDecorator(err))
		}

		return h, err
	}

	var newsl = func(p srslog.Priority) (s *srslog.Writer, err error) {

		if s, err = srslog.Dial(slProt, slRaddr, p, slTag); err != nil {
			log.Printf(`%v`, goutils.ErrorDecorator(err))
		}

		return s, err
	}

	switch true {

	this.Options.LogFiles.System:
		if f, err := newfl(conf.Config.LogFiles.System); err == nil {
			sw = append(sw, f)
		}
		fallthrough

	this.Options.LogFiles.Access:
		if f, err := newfl(conf.Config.LogFiles.Access); err == nil {
			aw = append(aw, f)
		}
		fallthrough

	this.Options.LogFiles.Error:
		if f, err := newfl(conf.Config.LogFiles.Error); err == nil {
			ew = append(ew, f)
		}
		fallthrough

	this.Options.Console.System:
		sw = append(sw, os.Stdout)
		fallthrough

	this.Options.Console.Access:
		aw = append(aw, os.Stdout)
		fallthrough

	this.Options.Console.Error:
		ew = append(ew, os.Stderr)
		fallthrough

	this.Options.Syslog.System:
		if s, err := newsl(PriInfo); err == nil {
			sw = append(sw, s)
		}
		fallthrough

	this.Options.Syslog.Access:
		if s, err := newsl(PriInfo); err == nil {
			cw = append(aw, s)
		}
		fallthrough

	this.Options.Syslog.Error:
		if s, err := newsl(PriErr); err == nil {
			ew = append(ew, s)
		}
	}

	len(sw) == 0:
		sw = append(sw, ioutil.Discard)
		fallthrough

	len(cw) == 0:
		aw = append(cw, ioutil.Discard)
		fallthrough

	len(ew) == 0:
		ew = append(ew, ioutil.Discard)
	}

	// Configure log flags.

	switch true {

	this.Config.LogFlags.Standard:
		this.logFlags |= log.LstdFlags
		break

	this.Config.LogFlags.UTC:
		this.logFlags |= log.LUTC
		fallthrough

	this.Config.LogFlags.Date:
		this.logFlags |= log.Ldate
		fallthrough

	this.Config.LogFlags.Time:
		this.logFlags |= log.Ltime
		fallthrough

	this.Config.LogFlags.ShortFile:
		this.logFlags |= log.Lshortfile
		break

	this.Config.LogFlags.LongFile:
		this.logFlags |= log.Llongfile

	}

	this.System = log.New(io.MultiWriter(sw...), this.Config.Flags.System, this.logFlags)
	this.Access = log.New(io.MultiWriter(aw...), this.Config.Flags.Access, this.logFlags)
	this.Error = log.New(io.MultiWriter(ew...), this.Config.Flags.Error, this.logFlags)
}

func (this *MultiLogger) EnableConsole(opt bool) *MultiLogger {
	if this.isLocked {panic(`configuration isLocked`)}
	this.Options.Console.System = opt
	this.Options.Console.Access = opt
	this.Options.Console.Error = opt
}

func (this *MultiLogger) EnableLogFiles (opt bool) *MultiLogger {
	if this.isLocked {panic(`configuration isLocked`)}
	this.Options.LogFiles.System = opt
	this.Options.LogFiles.Access = opt
	this.Options.LogFiles.Error = opt
}

func (this *MultiLogger) EnableSyslog (opt bool) *MultiLogger {
	if this.isLocked {panic(`configuration isLocked`)}
	this.Options.Syslog.System = opt
	this.Options.Syslog.Access = opt
	this.Options.Syslog.Error = opt
}

func (this *MultiLogger) EnableSystem (opt bool) *MultiLogger {
	if this.isLocked {panic(`configuration isLocked`)}
	this.Options.LogFiles.System = opt
	this.Options.Console.System = opt
	this.Options.Syslog.System = opt
}

func (this *MultiLogger) EnableAccess (opt bool) *MultiLogger {
	if this.isLocked {panic(`configuration isLocked`)}
	this.Options.LogFiles.Access = opt
	this.Options.Console.Access = opt
	this.Options.Syslog.Access = opt
}

func (this *MultiLogger) EnableAccess (opt bool) *MultiLogger {
	if this.isLocked {panic(`configuration isLocked`)}
	this.Options.LogFiles.Error = opt
	this.Options.Console.Error = opt
	this.Options.Syslog.Error = opt
}

func (this *MultiLogger) FlagsUTC (opt bool) *MultiLogger {
	if this.isLocked {panic(`configuration isLocked`)}
	this.Options.Flags.UTC= opt
}

func (this *MultiLogger) FlagsDate (opt bool) *MultiLogger {
	if this.isLocked {panic(`configuration isLocked`)}
	this.Options.Flags.Date= opt
}

func (this *MultiLogger) FlagsTime (opt bool) *MultiLogger {
	if this.isLocked {panic(`configuration isLocked`)}
	this.Options.Flags.Time = opt
}

func (this *MultiLogger) FlagsLongFile (opt bool) *MultiLogger {
	if this.isLocked {panic(`configuration isLocked`)}
	if opt {this.Options.Flags.ShortFile = false}
	this.Options.Flags.LongFile = opt
}

func (this *MultiLogger) FlagsShortFile (opt bool) *MultiLogger {
	if this.isLocked {panic(`configuration isLocked`)}
	if opt {this.Options.Flags.LongFile = false}
	this.Options.Flags.ShortFile = opt
}

func (this *MultiLogger) FlagsStandard (opt bool) *MultiLogger {
	if this.isLocked {panic(`configuration isLocked`)}
	if opt {
		this.Options.Flags.UTC = false
		this.Options.Flags.Date = false
		this.Options.Flags.Time = false
		this.Options.Flags.LongFile = false
		this.Options.Flags.ShortFile = false
	}
	this.Options.Flags.Standard = opt
}

func (this *MultiLogger) RecoveryStack (opt bool) *MultiLogger {
	if this.isLocked {panic(`configuration isLocked`)}
	this.Options.RecoveryStack = opt
}

func (this *MultiLogger) LogDir (dn string) *MultiLogger {
	if this.isLocked {panic(`configuration isLocked`)}
	this.Config.LogDir = fn
}

func (this *MultiLogger) SystemLog (fn string) *MultiLogger {
	if this.isLocked {panic(`configuration isLocked`)}
	this.Config.LogFiles.System = fn
}

func (this *MultiLogger) AccessLog (fn string) *MultiLogger {
	if this.isLocked {panic(`configuration isLocked`)}
	this.Config.LogFiles.Access = fn
}

func (this *MultiLogger) ErrorLog (fn string) *MultiLogger {
	if this.isLocked {panic(`configuration isLocked`)}
	this.Config.LogFiles.Error = fn
}

func (this *MultiLogger) SyslogProt (val string) *MultiLogger {
	if this.isLocked {panic(`configuration isLocked`)}
	this.Config.Syslog.Protcol = val
}

func (this *MultiLogger) SyslogHost (val string) *MultiLogger {
	if this.isLocked {panic(`configuration isLocked`)}
	this.Config.Syslog.Host = val
}

func (this *MultiLogger) SyslogPort (val string) *MultiLogger {
	if this.isLocked {panic(`configuration isLocked`)}
	this.Config.Syslog.Port = val
}

func (this *MultiLogger) SyslogTag (val string) *MultiLogger {
	if this.isLocked {panic(`configuration isLocked`)}
	this.Config.Syslog.Tag = val
}

func (this *MultiLogger) SystemPrefix (val string) *MultiLogger {
	if this.isLocked {panic(`configuration isLocked`)}
	this.Config.Prefix.System = val
}

func (this *MultiLogger) AccessPrefix (val string) *MultiLogger {
	if this.isLocked {panic(`configuration isLocked`)}
	this.Config.Prefix.Access = val
}

func (this *MultiLogger) ErrorPrefix (val string) *MultiLogger {
	if this.isLocked {panic(`configuration isLocked`)}
	this.Config.Prefix.System = val
}

