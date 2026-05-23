# linbench

`linbench` est un outil de benchmark matériel en ligne de commande pour Linux, écrit en Go.

Il mesure les performances principales d'une machine sans interface graphique :

- RAM : lecture, écriture, copie et latence
- caches CPU L1, L2 et L3 : débit et latence
- CPU : score entier, score flottant, SHA-256 et compression gzip
- disque : lecture/écriture séquentielle, IOPS aléatoires et latence

Le projet vise une première version simple, maintenable et utilisable sans privilèges root pour les tests RAM, cache et CPU. Le benchmark disque travaille uniquement sur un fichier temporaire dans un dossier choisi, jamais sur un périphérique brut.

## Installation

Prérequis :

- Go 1.24 ou plus récent
- Linux pour exécuter les benchmarks matériels complets

Construire le binaire :

```bash
go build -o linbench ./cmd/linbench
```

Construire explicitement pour Linux amd64 depuis une autre plateforme :

```bash
GOOS=linux GOARCH=amd64 go build -o linbench ./cmd/linbench
```

## Utilisation

Lancer tous les benchmarks :

```bash
./linbench
```

Équivalent explicite :

```bash
./linbench --all
```

Lancer un seul groupe de tests :

```bash
./linbench --ram
./linbench --cache
./linbench --cpu
./linbench --disk
```

Produire du JSON :

```bash
./linbench --json
```

Écrire la sortie dans un fichier :

```bash
./linbench --json --output results.json
```

Configurer les tailles et la durée :

```bash
./linbench --duration 5s --memory-size 512M --disk-size 1G
```

Tester un chemin disque précis :

```bash
./linbench --disk --disk-path /mnt/data
```

## Options CLI

```text
--all                 Lance tous les benchmarks
--ram                 Lance uniquement le benchmark RAM
--cache               Lance uniquement le benchmark cache CPU
--cpu                 Lance uniquement le benchmark CPU
--disk                Lance uniquement le benchmark disque

--json                Sortie JSON
--output FILE         Écrit le résultat dans un fichier

--duration DURATION   Durée approximative de chaque test, ex: 5s
--threads N           Nombre de threads CPU, ou auto

--memory-size SIZE    Taille du buffer RAM, ex: 512M, 1G
--disk-path PATH      Dossier où créer le fichier de test disque
--disk-size SIZE      Taille du fichier de test disque, ex: 1G
--keep-test-file      Ne supprime pas le fichier de test disque
--no-sync             Ne force pas fsync pendant les écritures disque

--verbose             Affiche les détails techniques sur stderr
--version             Affiche la version
```

## Sorties

La sortie texte est lisible directement dans un terminal et regroupe les mesures par section.

La sortie JSON expose une structure stable avec les blocs suivants :

- `system`
- `memory`
- `cache`
- `cpu`
- `disk`
- `errors`, présent uniquement si une section a échoué sans bloquer les autres

Exemple :

```bash
./linbench --json --cpu
```

## Fonctionnement du code

Le code est organisé en packages internes pour séparer les responsabilités :

```text
cmd/linbench/main.go        Parsing CLI, orchestration des benchmarks, écriture de sortie
internal/model/             Structures de résultats partagées
internal/benchmark/         Implémentation des benchmarks RAM, cache, CPU et disque
internal/system/            Collecte et parsing des informations système Linux
internal/output/            Formatage texte et JSON
internal/unit/              Parsing des tailles et constantes d'unité
```

### RAM

Le benchmark RAM alloue un grand buffer, puis mesure :

- lecture séquentielle
- écriture séquentielle
- copie de buffer
- latence via pointer chasing pseudo-aléatoire

Les boucles alimentent une variable globale `Sink` afin d'éviter que le compilateur élimine le travail mesuré.

### Cache CPU

Les tailles de cache sont lues depuis :

```text
/sys/devices/system/cpu/cpu0/cache/index*
```

Chaque niveau est mesuré avec un buffer dimensionné pour tenir dans le cache ciblé. Si les informations Linux ne sont pas disponibles, le code utilise des valeurs de repli.

### CPU

Le benchmark CPU utilise par défaut `runtime.NumCPU()` goroutines, configurable avec `--threads`.

Il mesure :

- opérations entières
- opérations flottantes
- débit SHA-256 avec `crypto/sha256`
- débit gzip avec `compress/gzip`

### Disque

Le benchmark disque crée un fichier temporaire dans `--disk-path`, mesure des accès séquentiels et aléatoires, puis supprime le fichier par défaut.

Le fichier est conservé uniquement avec :

```bash
./linbench --disk --keep-test-file
```

`--no-sync` permet de ne pas forcer `fsync`, ce qui mesure davantage l'effet du cache OS que le débit réellement flushé.

## Développement

Formater :

```bash
gofmt -w cmd internal
```

Lint standard Go :

```bash
go vet ./...
```

Tests :

```bash
go test ./...
```

Build Linux amd64 :

```bash
GOOS=linux GOARCH=amd64 go build ./...
```

Si l'environnement empêche Go d'écrire dans le cache utilisateur, utiliser un cache local temporaire :

```bash
GOCACHE=/private/tmp/go-build-cache go test ./...
```

## Limites

Les résultats peuvent varier selon la fréquence CPU, le gouverneur d'alimentation, la température, la charge système, NUMA, la virtualisation et le cache disque Linux.

Le projet fournit des mesures pratiques et rapides, pas un protocole de benchmark scientifique strict.


