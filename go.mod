module github.com/afeldman/fluxbrain

go 1.22

// Errorbrain als externe Dependency (auskommentiert für lokale Entwicklung)
// require github.com/afeldman/errorbrain v0.0.0

// Lokale Entwicklung: Errorbrain im Schwester-Verzeichnis
// Für Produktion: replace-Direktive entfernen und require aktivieren
// replace github.com/afeldman/errorbrain => ../errorbrain

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/redis/go-redis/v9 v9.17.2 // indirect
)
