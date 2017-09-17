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
	`io`
	`log`
	`os`
	`path/filepath`
)

type MultiLogger struct {
	*log.Logger
	writers	[]io.Writer
	files	[]*os.File
	stdout	bool
	stderr	bool
}

func NewMultiLogger(prefix string, flags int, stdout, stderr bool, files ...string) *MultiLogger {

	this := new(MultiLogger)

	for _, fn := range files {
		if fh, err := this.mkDirOpen(fn); err != nil {
			log.Println(err)
		} else {
			this.files = append(this.files, fh)
		}
	}

	this.
		SetPrefix(prefix).
		SetFlags(flags).
		SetStdout(stdout).
		SetStderr(stderr)

	return this
}

func (this *MultiLogger) AddFile(fn string) (err error) {
	if fh, err := this.mkDirOpen(fn); err == nil {
		this.files = append(this.files, fh)
		this.refreshWriters()
	}
	return err
}

func (this *MultiLogger) AddWriter(writer io.Writer) *MultiLogger {
	this.writers = append(this.writers, writer)
	this.refreshWriters()
	return this
}

func (this *MultiLogger) SetStdout(opt bool) *MultiLogger {
	this.stdout = opt
	this.refreshWriters()
	return this
}

func (this *MultiLogger) SetStderr(opt bool) *MultiLogger {
	this.stderr = opt
	this.refreshWriters()
	return this
}

func (this *MultiLogger) SetFlags(flags int) *MultiLogger {
	this.Logger.SetFlags(flags)
	return this
}

func (this *MultiLogger) SetOutput(writer io.Writer) *MultiLogger {
	this.Logger.SetOutput(writer)
	return this
}

func (this *MultiLogger) SetPrefix(prefix string) *MultiLogger {
	this.Logger.SetPrefix(prefix)
	return this
}

func (this *MultiLogger) mkDirOpen(fn string) (*os.File, error) {
	const (
		fFlags = os.O_APPEND|os.O_CREATE|os.O_WRONLY
		fMode = 0640
		dMode = 0750
	)
	if err := os.MkdirAll(filepath.Dir(fn), dMode); err != nil {
		return nil, err
	}
	return os.OpenFile(fn, fFlags, fMode)
}

func (this *MultiLogger) refreshWriters() {

	var writers []io.Writer

	if this.stdout {
		writers = append(writers, os.Stdout)
	}
	if this.stderr {
		writers = append(writers, os.Stderr)
	}
	for _, w := range this.writers {
		writers = append(writers, w)
	}
	for _, f := range this.files {
		writers = append(writers, f)
	}

	this.SetOutput(io.MultiWriter(writers...))
}
