package main

import "github.com/jscherff/goutil"

func main() {

	ml := NewMultiLogger("test", flags int, stdout, stderr bool, files ...string) *MultiLogger {
func (this *MultiLogger) AddFile(fn string) (err error) {
func (this *MultiLogger) AddWriter(writer io.Writer) *MultiLogger {
func (this *MultiLogger) SetStdout(opt bool) *MultiLogger {
func (this *MultiLogger) SetStderr(opt bool) *MultiLogger {
func (this *MultiLogger) SetFlags(flags int) *MultiLogger {
func (this *MultiLogger) SetOutput(writer io.Writer) *MultiLogger {
func (this *MultiLogger) SetPrefix(prefix string) *MultiLogger {
func (this *MultiLogger) mkDirOpen(fn string) (*os.File, error) {
func (this *MultiLogger) refreshWriters() {
