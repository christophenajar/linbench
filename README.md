# linbench

`linbench` est un outil de benchmark matériel en ligne de commande pour Linux, écrit en Go.

Version actuelle : `0.2.1`.

Il mesure les performances principales d'une machine sans interface graphique :

- RAM : lecture, écriture, copie et latence
- caches CPU L1, L2 et L3 : débit et latence
- CPU : score entier, score flottant, SHA-256 et compression gzip
- disque : lecture/écriture séquentielle, IOPS aléatoires et latence
- réseau : audit NIC, PCIe, NUMA, RDMA/RoCE et IRQ affinity

Le projet vise une première version simple, maintenable et utilisable sans privilèges root pour les tests RAM, cache et CPU. Le benchmark disque travaille uniquement sur un fichier temporaire dans un dossier choisi, jamais sur un périphérique brut.

## Installation

Prérequis :

- Go 1.24 ou plus récent
- Linux pour exécuter les benchmarks matériels complets

Construire le binaire :

```bash
go build -o linbench ./cmd/linbench
```

Construire un binaire Linux amd64 pure-Go depuis une autre plateforme :

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags '-s -w' -o bin/linbench-linux-amd64 ./cmd/linbench
```

Construire sur Linux avec les noyaux C optimisés pour les débits RAM/cache :

```bash
CGO_ENABLED=1 go build -trimpath -ldflags '-s -w' -o bin/linbench-linux-amd64 ./cmd/linbench
```

Le build cgo demande un compilateur C local, par exemple `gcc`. Il est recommandé de le compiler directement sur la machine Debian cible. Le build `CGO_ENABLED=0` reste supporté et utilise les boucles Go de repli.

### Build optimise cgo sur Debian

Depuis `0.1.2`, les benchmarks de debit RAM/cache peuvent utiliser des noyaux C via cgo. Ce mode reduit le cout des boucles Go et donne des debits cache plus proches de ce que le materiel peut faire.

Sur Debian, installer les outils de compilation :

```bash
sudo apt update
sudo apt install -y golang build-essential git
```

Recuperer le projet et compiler avec cgo :

```bash
git clone https://github.com/christophenajar/linbench.git
cd linbench
mkdir -p bin
CGO_ENABLED=1 go build -trimpath -ldflags '-s -w' -o bin/linbench-linux-amd64 ./cmd/linbench
```

Si le binaire est compile et execute sur la meme machine, il est possible de demander au compilateur C d'utiliser les optimisations propres au CPU local :

```bash
CGO_ENABLED=1 CGO_CFLAGS='-O3 -march=native' go build -trimpath -ldflags '-s -w' -o bin/linbench-linux-amd64 ./cmd/linbench
```

Ne pas utiliser `-march=native` pour construire un binaire destine a etre copie sur des machines differentes.

Verifier la version :

```bash
./bin/linbench-linux-amd64 --version
```

Verifier que le binaire utilise bien cgo :

```bash
ldd ./bin/linbench-linux-amd64
```

Un binaire cgo affiche des bibliotheques dynamiques comme `libc.so.6`. Un binaire pure-Go compile avec `CGO_ENABLED=0` est generalement indique comme `statically linked` par `file`.

Tester uniquement RAM/cache avec les noyaux C :

```bash
./bin/linbench-linux-amd64 --ram --cache --duration 5s --memory-size 1G
```

Pour un binaire portable sans dependance libc dynamique, utiliser le fallback pure-Go :

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags '-s -w' -o bin/linbench-linux-amd64 ./cmd/linbench
```

Ce binaire reste plus simple a distribuer, mais les debits RAM/cache seront limites par les boucles Go.

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
./linbench --network
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
--network             Lance uniquement l'audit réseau/RDMA/NUMA
--network-all         Inclut toutes les interfaces réseau dans l'audit

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
- `benchmark_engine`
- `numa`
- `memory`
- `cache`
- `cpu`
- `disk`
- `network`
- `warnings`, présent uniquement si les conditions de test peuvent fausser fortement les résultats
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

La sortie indique le moteur de benchmark utilise :

- `Engine : go` avec un build `CGO_ENABLED=0`
- `Engine : cgo` avec un build `CGO_ENABLED=1`

Avec cgo, les noyaux C utilisent des boucles batchees. Quand le compilateur expose `__AVX2__` via `-march=native`, le noyau d'ecriture utilise un chemin AVX2 conditionnel. La copie reste volontairement deleguee a `memcpy`, car la libc choisit souvent une routine mieux optimisee que les intrinsics manuels selon le CPU et la taille du buffer.

Sur les machines multi-socket ou multi-node NUMA, les resultats RAM/cache peuvent varier selon le placement NUMA. `linbench` affiche un warning si plusieurs nodes NUMA sont detectes.

### NUMA

`linbench` lit les informations NUMA depuis :

```text
/sys/devices/system/node/node*/cpulist
/sys/devices/system/node/node*/meminfo
/sys/devices/system/node/node*/distance
```

La sortie affiche :

- nombre de nodes NUMA
- CPUs rattaches a chaque node
- memoire totale et libre par node
- matrice de distance NUMA
- commandes `numactl` recommandees pour comparer acces local et distant

Exemple de comparaison :

```bash
numactl --cpunodebind=0 --membind=0 ./bin/linbench-linux-amd64 --ram --cache --duration 5s --memory-size 1G
numactl --cpunodebind=0 --membind=1 ./bin/linbench-linux-amd64 --ram --cache --duration 5s --memory-size 1G
```

### CPU

Le benchmark CPU utilise par défaut `runtime.NumCPU()` goroutines, configurable avec `--threads`.

Il mesure :

- opérations entières
- opérations flottantes
- débit SHA-256 avec `crypto/sha256`
- débit gzip avec `compress/gzip`

### Disque

Le benchmark disque crée un fichier temporaire dans `--disk-path`, mesure des accès séquentiels et aléatoires, puis supprime le fichier par défaut.

Par défaut, `linbench` utilise `~/tmp` si ce dossier existe. Sinon il utilise le dossier temporaire du système.

Depuis `0.1.1`, le benchmark tente de réduire l'effet du cache de page Linux avant les lectures avec `posix_fadvise(..., DONTNEED)`. Ce n'est pas une garantie absolue : le cache OS, le contrôleur disque et le filesystem peuvent encore influencer les résultats.

Pour éviter de mesurer surtout la RAM ou le cache OS, préférer un chemin sur un vrai disque et un fichier plus grand :

```bash
./linbench --disk --disk-path /var/tmp --disk-size 4G --duration 5s
```

`linbench` affiche des warnings si `--disk-path` pointe vers `/tmp`, un filesystem mémoire comme `tmpfs`, si `--disk-size` est inférieur à `1G`, ou si `--no-sync` est utilisé.

Le fichier est conservé uniquement avec :

```bash
./linbench --disk --keep-test-file
```

`--no-sync` permet de ne pas forcer `fsync`, ce qui mesure davantage l'effet du cache OS que le débit réellement flushé.

### Network / RDMA

Depuis `0.2.0`, `linbench` inclut un audit réseau passif pour les serveurs haute performance, notamment Mellanox / NVIDIA ConnectX et RoCE.

Lancer uniquement l'audit réseau :

```bash
./bin/linbench-linux-amd64 --network
```

Par défaut, `--network` se concentre sur les interfaces Mellanox/NVIDIA et RDMA afin d'éviter le bruit des interfaces de management, Broadcom 1 GbE, iDRAC, etc. Pour afficher toutes les interfaces physiques détectées :

```bash
./bin/linbench-linux-amd64 --network-all
```

Le module lit les informations depuis Linux, sans générer de trafic réseau :

```text
/sys/class/net
/sys/class/net/<iface>/device
/sys/class/infiniband
/proc/irq
/proc/interrupts
/proc/self/status
```

Il vérifie :

- interfaces réseau et driver
- interfaces Mellanox / NVIDIA `mlx5`
- mapping netdev vers adresse PCI
- NUMA node du périphérique PCIe
- CPU list locale du périphérique
- vitesse et largeur PCIe courantes/max
- MTU, état du lien, vitesse du lien
- device RDMA associé, par exemple `mlx5_0`
- disponibilité RoCE via les GID types RDMA
- IRQ MSI/MSI-X et `smp_affinity_list`
- affinité CPU du processus courant
- présence des outils RDMA perftest : `ib_write_bw`, `ib_read_bw`, `ib_send_bw`, `ib_write_lat`, `ib_read_lat`, `ib_send_lat`

Le module signale notamment :

- process exécuté sur des CPUs hors du NUMA node local de la NIC
- IRQs mlx5 autorisées sur des CPUs distants
- lien réseau down ou vitesse inférieure à 25 Gb/s sur carte Mellanox
- MTU inférieur à 9000 quand RoCE est détecté
- lien PCIe dégradé par rapport à la capacité max
- device RDMA manquant
- outils perftest manquants

Sur une machine NUMA, l'audit recommande une commande alignant CPU et mémoire sur le node local de la NIC :

```bash
numactl --cpunodebind=0 --membind=0 ./bin/linbench-linux-amd64 --network
```

Les tests actifs RDMA, comme `ib_write_bw` ou `ib_read_bw`, ne sont pas lancés automatiquement en `0.2.0`, car ils nécessitent un pair distant et une configuration RoCE correcte. Le module se limite à vérifier que la machine locale est cohérente avant de lancer ces outils manuellement.

## Changements 0.1.1

- Boucles RAM et cache en `uint64` au lieu de parcours byte-par-byte.
- Lecture RAM/cache sur tous les mots du buffer, avec comptage des octets réellement lus.
- Latence cache corrigée avec un working set dimensionné au niveau ciblé.
- Tentative de purge du cache de page Linux avant lecture disque via `posix_fadvise`.
- Warnings visibles en texte et JSON pour `/tmp`, `tmpfs`, petites tailles disque et `--no-sync`.

## Changements 0.1.2

- Noyaux C optionnels via cgo pour les débits RAM/cache.
- Fallback Go automatique quand `CGO_ENABLED=0`.
- Chemin disque par défaut changé vers `~/tmp` quand ce dossier existe.

## Changements 0.1.3

- Noyaux C RAM/cache batchés pour réduire le coût de `clock_gettime` sur les petits buffers L1/L2.
- Boucles C de lecture et écriture déroulées avec plusieurs accumulateurs indépendants.
- Documentation du build `CGO_CFLAGS='-O3 -march=native'` pour les builds locaux optimisés.

## Changements 0.1.4

- Affichage du moteur de benchmark : `go` ou `cgo`.
- Détection du nombre de sockets CPU dans la sortie système.
- Warning NUMA sur machines multi-socket.
- Chemins AVX2 conditionnels pour les noyaux C write/copy quand le binaire est compilé avec `-march=native` sur CPU compatible.

## Changements 0.1.5

- Retour a `memcpy` pour le noyau C copy, afin de laisser la libc choisir la meilleure implementation.
- Conservation du chemin AVX2 conditionnel pour le noyau C write.

## Changements 0.1.6

- Ajout d'une section NUMA dans la sortie texte et JSON.
- Lecture des nodes, CPU lists, meminfo et distances depuis `/sys/devices/system/node`.
- Recommandations automatiques de commandes `numactl` quand plusieurs nodes NUMA sont detectes.

## Changements 0.2.0

- Ajout de `--network`.
- Audit passif des interfaces réseau, PCIe, NUMA, RDMA/RoCE et IRQ affinity.
- Détection Mellanox / NVIDIA `mlx5`.
- Mapping netdev vers PCI address et NUMA node.
- Détection des outils RDMA perftest installés.
- Warnings et recommandations `numactl` pour aligner process, mémoire et NIC sur le même node NUMA.

## Changements 0.2.1

- `--network` filtre par défaut les interfaces non Mellanox/RDMA pour réduire le bruit.
- Ajout de `--network-all` pour afficher toutes les interfaces.
- Déduplication des recommandations `numactl`.
- Meilleure gestion des interfaces Mellanox en bond RDMA, par exemple `mlx5_bond_0`.

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
