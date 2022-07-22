package logging

type Log interface {
	Print(args ...any)
	Printf(format string, args ...any)
	Println(args ...any)
}
