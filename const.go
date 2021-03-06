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
	`log`
	`os`
	`github.com/RackSec/srslog`
)

const (
	SyslogPriInfo = srslog.LOG_LOCAL7|srslog.LOG_INFO
	SyslogPriErr = srslog.LOG_LOCAL7|srslog.LOG_ERR

	FileFlagsAppend = os.O_APPEND|os.O_CREATE|os.O_WRONLY
	FileModeDefault = 0640
	DirModeDefault = 0750

	LoggerFlags = log.LstdFlags
)
