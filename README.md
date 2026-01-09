git clone https://github.com/afeldman/errorbrain.git
go build ./cmd/fluxbrain
go test ./...
go test ./...
go test ./internal/state/...
go test ./internal/context/...
go test -cover ./...
# Fluxbrain


Fluxbrain ist ein Adapter für FluxCD.
Er sammelt ausschließlich Fakten (Status, Events, Logs, Metadaten) und exportiert sie gemäß der errorbrain-Spec.
Fluxbrain trifft keine Entscheidungen, erzeugt keine Verdicts und enthält keine AI/LLM-Logik.
Analyse und Ableitung erfolgen ausschließlich in errorbrain.

**Fakten = Kontext, keine Interpretation!**
Fluxbrain liefert nur das, was beobachtbar ist – keine Bewertung, keine Ableitung, keine Priorisierung.
Das hilft errorbrain, weil alle Kontextdaten sauber und nachvollziehbar sind.

**Erlaubte Zusatzfelder (Beispiele):**
- Flux-Objekt: `kind`, `namespace`, `name`
- Reconcile-Infos: `attempt`, `lastTransitionTime`
- Revision: `gitRevision`, `helmChartVersion`
- Controller-Quelle: `controllerName`

🔒 **Regel:** Alles bleibt beobachtbar, nichts wird bewertet.

---


## Leitplanken

- Nur Fakten und Rohsignale (Status, Events, Logs, Metadaten)
- Keine Ursachenbewertung, keine Priorisierung, keine menschlichen Learnings
- Errorbrain (extern) übernimmt Analyse und LLM-Auswahl
- Output muss errorbrain-spec-konform sein; Fehler im Export dürfen den Collector nicht blockieren
- **Jedes Feld muss aus dem beobachtbaren Kontext stammen – keine Interpretation!**

---

## Aktueller Stand

- Collector: `FluxEventCollector` sammelt Kubernetes `Warning` Events für Flux-Kustomizations. Der `KubernetesEventLister` ist noch ein Placeholder (client-go muss verdrahtet werden).
- Context Builder: baut deterministischen JSON-Kontext aus Status-, Event- und Log-Signalen (`internal/context`).
- Analyzer: `MockAnalyzer` als Platzhalter, bis das errorbrain SDK verfügbar ist. Keine eigene Analyse-Logik.
- Notifier: Slack-, Webhook- und GitHub-Issue-Notifier (`internal/notify`).
- State: In-Memory-Fingerprinting und Backoff, um Notification-Spam zu verhindern (`internal/state`).

Aktuelle Verantwortlichkeiten:

| Fluxbrain | Errorbrain |
|-----------|------------|
| FluxCD-Signale sammeln | Analyse, Reasoning, Verdicts |
| Kontext aufbauen | Prompting & Modellwahl |
| Dedupe & Backoff | Retry-Safety, Klassifikation |
| Notifications zustellen | Bewertung & Empfehlung |

---

## Architektur (Faktenfluss)

1. Collector liest Events → erzeugt `[]ErrorContext`.
2. Fingerprint per SHA256 → Backoff-Check (`state.MemoryStore`).
3. Analyzer (Platzhalter) → `AnalysisResult` (später errorbrain-Adapter).
4. Notifier senden den Kontext + Resultate weiter (Slack/Webhook/GitHub).
5. Backoff-Status wird aktualisiert (Failure → längerer Backoff, Success → Reset).

Geplante Erweiterungen: echter Kubernetes-EventLister via client-go, weitere Flux-Ressourcen (HelmRelease, GitRepository), optionale Log-Signale, persistenter State.

---

## Konfiguration

Pflicht:

| Variable | Beschreibung |
|----------|--------------|
| `FLUXBRAIN_CLUSTER` | Cluster-Name für Kontext (wird in jedem Event mitgegeben) |

Optional:

| Variable | Default | Beschreibung |
|----------|---------|--------------|
| `FLUXBRAIN_RUN_MODE` | `continuous` | `once` für CronJobs, sonst Continuous Mode |
| `FLUXBRAIN_REQUEUE_INTERVAL` | `5m` | Intervall im Continuous Mode |
| `FLUXBRAIN_FLUX_NAMESPACE` | `flux-system` | Namespace, in dem Flux-Events gelesen werden |
| `FLUXBRAIN_SLACK_WEBHOOK` | - | Slack Incoming Webhook |
| `FLUXBRAIN_WEBHOOK_URL` | - | Beliebiger HTTP-Webhook (liefert Kontext + Result) |
| `FLUXBRAIN_GITHUB_OWNER` | - | Owner für GitHub-Issues |
| `FLUXBRAIN_GITHUB_REPO` | - | Repo für GitHub-Issues |
| `FLUXBRAIN_GITHUB_TOKEN` | - | Token für GitHub-Issues |

Run Modes:

- `once`: einmalige Ausführung (CronJob, kein Ticker)
- `continuous` (Default): Ticker-basiert mit `FLUXBRAIN_REQUEUE_INTERVAL`

---

## Deployment (Beispiele)

### CronJob (once)

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: fluxbrain
  namespace: flux-system
spec:
  schedule: "*/5 * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: fluxbrain
            image: ghcr.io/afeldman/fluxbrain:latest
            env:
            - name: FLUXBRAIN_RUN_MODE
              value: "once"
            - name: FLUXBRAIN_CLUSTER
              value: "prod-eu-west-1"
          restartPolicy: OnFailure
```

### Deployment (continuous)

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: fluxbrain
  namespace: flux-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: fluxbrain
  template:
    metadata:
      labels:
        app: fluxbrain
    spec:
      containers:
      - name: fluxbrain
        image: ghcr.io/afeldman/fluxbrain:latest
        env:
        - name: FLUXBRAIN_CLUSTER
          value: "prod-eu-west-1"
        - name: FLUXBRAIN_REQUEUE_INTERVAL
          value: "5m"
```

---

## Build & Release

- Lokal: `go build -o fluxbrain ./cmd/fluxbrain`
- Docker: `docker build -t fluxbrain:local .`
- GoReleaser (Snapshot): `goreleaser release --snapshot --clean` → Artefakte in `dist/`

`.goreleaser.yaml` baut für `linux`/`darwin` auf `amd64` und `arm64` mit statischem Binary (`CGO_ENABLED=0`).

---

## Entwicklung

- Kubernetes-Events werden aktuell nicht aus einem echten Cluster gelesen. Verdrahtung von `KubernetesEventLister` mit client-go steht noch aus.
- errorbrain-SDK fehlt noch; der `MockAnalyzer` füllt nur die Schnittstelle, trifft aber keine Entscheidungen.
- Fingerprinting basiert auf Cluster, Namespace, Kind, Name, Reason, Git-Revision; Backoff default: 30s pro Fehler, gedeckelt auf 1h.
- Deterministisches JSON: `internal/context.MarshalErrorContext` nutzt stabiles Encoding ohne HTML-Escaping.

Empfohlene Tests:

```bash
go test ./...
```

---

## Hinweise

- Fluxbrain bleibt ein Sensor. Sobald errorbrain verfügbar ist, wird die Analyzer-Implementierung durch einen reinen Adapter ersetzt.
- Keine neue Logik in Fluxbrain hinzufügen, die über Fakten hinausgeht. Weniger Logik ist korrekt.
