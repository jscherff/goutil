package main

import `github.com/jscherff/goutils`

func main() {
	mlw, _ := goutils.NewMultiLoggerWriter("config_in.json")
	mlw.Init()
	mlw.SaveConfig("config_out.json")
}
