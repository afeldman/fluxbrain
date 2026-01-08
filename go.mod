module github.com/afeldman/fluxbrain

go 1.22

require (
	k8s.io/api v0.29.0
	k8s.io/apimachinery v0.29.0
	k8s.io/client-go v0.29.0
)

// Errorbrain als externe Dependency (auskommentiert für lokale Entwicklung)
// require github.com/afeldman/errorbrain v0.0.0

// Lokale Entwicklung: Errorbrain im Schwester-Verzeichnis
// Für Produktion: replace-Direktive entfernen und require aktivieren
// replace github.com/afeldman/errorbrain => ../errorbrain
