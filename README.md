### kube-arena

**kube-arena** proporciona:

- **Un CLI en el host** (`ks-arena`) instalable como binario en Linux.
- **Una imagen Docker runner** con **KIND + Terraform + ClusterLoader2** que el CLI orquesta vía `docker run`.

### Requisitos

- Docker instalado en la máquina host.
- Para ejecutar KIND dentro del contenedor runner se recomienda:
  - Ejecutar el contenedor en modo `--privileged` **o**
  - Compartir el socket de Docker del host: `-v /var/run/docker.sock:/var/run/docker.sock`.

### Instalación del CLI (Linux)

En GitHub Releases (tags `v*`) se publican binarios:

- `ks-arena-linux-amd64`
- `ks-arena-linux-arm64`

Ejemplo (amd64):

```bash
sudo install -m 0755 ks-arena-linux-amd64 /usr/local/bin/ks-arena
ks-arena --help
```

### Build local de la imagen runner

```bash
docker build -t ks-arena:local .
```

### Uso básico

El CLI ejecuta el runner montando el directorio del YAML como `/workspace` dentro del contenedor.

```bash
ks-arena kind --image ks-arena:local examples/example.yaml

ks-arena test --image ks-arena:local examples/test-example.yaml
```

### Esquema de `example.yaml` (KIND + Terraform)

```yaml
kindConfigPath: kind-config.yaml   # opcional, configuración avanzada de KIND
clusterName: ks-arena-cluster      # nombre del cluster KIND
terraformDir: terraform            # directorio con los módulos de Terraform
variables:                         # variables opcionales para Terraform
  namespace: "perf-tests"
  replicas: "3"
```

### Esquema de `test-example.yaml` (ClusterLoader2)

```yaml
clusterName: ks-arena-cluster      # nombre del cluster KIND a usar
# kubeconfigPath: /root/.kube/config   # opcional, usa el default de KIND si se omite
testConfigPath: clusterloader2-config.yaml
repeat: 1
timeoutSeconds: 1800
```

### CI/CD con GitHub Actions y GHCR

Este repositorio incluye un workflow en `.github/workflows/build-and-publish.yml` que:

- Se ejecuta en `push` a `main` y en tags `v*`.
- Construye la imagen Docker a partir del `Dockerfile`.
- Publica la imagen en **GitHub Container Registry (GHCR)** como:
  - `ghcr.io/<owner>/<repo>/ks-arena:<tag>`
 - En tags `v*`, además publica binarios Linux del CLI en GitHub Releases.

No necesitas credenciales adicionales si usas `GITHUB_TOKEN` por defecto, ya que el workflow configura los permisos `packages: write`.

