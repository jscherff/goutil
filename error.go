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
	`fmt`
	`runtime`
	`path/filepath`
)

// ErrorDecorator prepends function filename, line number, and function name
// to error messages.
func ErrorDecorator(err error) (error) {

	// Need space only for PC of caller.
	pc := make([]uintptr, 1)

	// Skip PC of Callers and decorator function.
	n := runtime.Callers(2, pc)

	// Return original error if no PCs available.
	if n == 0 { return err }

	// Obtain the caller frame.
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	// Decorate error with caller information and return.
	return fmt.Errorf(`%s:%d: %s(): %v`,
		filepath.Base(frame.File),
		frame.Line,
		frame.Function,
		err,
	)
}

// PrintLine prints the line on which the caller executes it.
func PrintLine() {

	// Need space only for PC of caller.
	pc := make([]uintptr, 1)

	// Skip PC of Callers and decorator function.
	n := runtime.Callers(2, pc)

	// Return if no PCs available.
	if n == 0 { return }

	// Obtain the caller frame.
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	// Print file and line number, then return.
	fmt.Printf("%s:%d\n",
		filepath.Base(frame.File),
		frame.Line,
	)
}
