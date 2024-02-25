module github.com/LinusGates/linusgates

go 1.22

require (
	github.com/JoshuaDoes/json v0.0.0-20200726213358-ec3860544ac0
	github.com/LinusGates/osmgr v0.0.0
	github.com/LinusGates/reghive v0.0.0
	github.com/MarinX/keylogger v0.0.0-20210528193429-a54d7834cc1a
	github.com/spf13/pflag v1.0.5
)

require (
	github.com/gabriel-samfira/go-hivex v0.0.0-20190725123041-b40bc95a7ced // indirect
	golang.org/x/sync v0.6.0 // indirect
	seehuhn.de/go/ncurses v0.2.1-0.20231214110636-c694e8edeef5 // indirect
)

require (
	github.com/otiai10/copy v1.14.0
	golang.org/x/sys v0.17.0 // indirect
)

replace github.com/LinusGates/reghive v0.0.0 => ../reghive

replace github.com/LinusGates/osmgr v0.0.0 => ../osmgr
