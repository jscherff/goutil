package main

import `github.com/jscherff/goutils`

func main() {

	mlw := goutils.NewMultiLoggerWriter().
		Defaults().
		EnableAccess(true).
		EnableConsole(true).
		EnableSyslog(false).
		Init()

	slog := mlw.GetSystemLogger()
	alog := mlw.GetAccessLogger()
	elog := mlw.GetErrorLogger()

	elog.Println("hello")
	alog.Println("wasshurname")
	slog.Println("goodbye")

	ew := mlw.GetErrorWriter()
	ew.Write([]byte("hello again\n"))

	/*
		GetSystemWriter()
		GetAccessWriter()
		GetErrorWriter()
		GetSystemBufWriter()
		GetAccessBufWriter()
		GetErrorBufWriter()
		GetSystemLogger()
		GetAccessLogger()
		GetErrorLogger()
	*/

	/*
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

		AppName(`appname`).
		AppDir(``).
		LogDir(`log`).
		SystemLog(`system.log`).
		AccessLog(`access.log`).
		ErrorLog(`error.log`).

		SyslogProt(``).
		SyslogHost(``).
		SyslogPort(``).
		SyslogTag(`appname`).

		SystemTag(`system`).
		AccessTag(`access`).
		ErrorTag(`error`).

		Defaults().
		SaveConfig(`config.json`)
	*/


	/*
		GetSystemWriter()
		GetAccessWriter()
		GetErrorWriter()
		GetSystemBufWriter()
		GetAccessBufWriter()
		GetErrorBufWriter()
		GetSystemLogger()
		GetAccessLogger()
		GetErrorLogger()
	*/
}

