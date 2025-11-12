package logger

type Options struct {
    Filename   string
    Level      string
    ToConsole  bool
    Encoding   string
    AddCaller  bool
    TimeKey    string
	
}