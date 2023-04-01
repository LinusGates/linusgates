module github.com/LinusGates/linusgates

go 1.20

require (
	github.com/JoshuaDoes/json v0.0.0-20200726213358-ec3860544ac0
	github.com/LinusGates/osmgr v0.0.0
	github.com/LinusGates/reghive v0.0.0
	github.com/MarinX/keylogger v0.0.0-20210528193429-a54d7834cc1a
	github.com/spf13/pflag v1.0.5
)

require github.com/gabriel-samfira/go-hivex v0.0.0-20190725123041-b40bc95a7ced // indirect

require (
	github.com/otiai10/copy v1.9.0
	golang.org/x/sys v0.4.0 // indirect
)

replace github.com/LinusGates/reghive v0.0.0 => ../reghive

replace github.com/LinusGates/osmgr v0.0.0 => ../osmgr
