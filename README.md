# WebApp Operator 개발 및 빌드 가이드

## 1. 프로젝트 구조

현재 프로젝트는 다음과 같은 구조로 구성되어 있습니다.

```text
~/operator/
├── dev/
│   └── Dockerfile
├── docs/
├── manifests/
└── webapp-operator/
    ├── Dockerfile
    ├── Makefile
    ├── PROJECT
    ├── api/
    ├── bin/
    ├── cmd/
    ├── config/
    ├── internal/
    ├── test/
    ├── go.mod
    ├── go.sum
    └── ...
```

구성은 다음과 같습니다.

* `dev/Dockerfile`

  * Operator 개발 및 빌드에 필요한 도구가 포함된 개발용 컨테이너 이미지
* `webapp-operator/`

  * 실제 Kubebuilder / Operator SDK 프로젝트
* `webapp-operator/Dockerfile`

  * Kubernetes에 배포할 Operator 실행 이미지 빌드용 Dockerfile
* `webapp-operator/Makefile`

  * Operator 빌드, Docker 이미지 빌드, 배포 등의 작업 정의

---

## 2. 개발용 Docker 이미지 빌드

호스트에서 프로젝트 루트 디렉터리로 이동합니다.

```bash
cd ~/operator
```

`dev/Dockerfile`을 사용하여 개발용 이미지를 빌드합니다.

```bash
docker build \
  -t operator-dev:1.0 \
  -f dev/Dockerfile \
  dev
```

빌드된 이미지를 확인합니다.

```bash
docker images | grep operator-dev
```

예상 결과:

```text
operator-dev    1.0    ...
```

---

## 3. 개발 컨테이너 실행

개발용 Docker 컨테이너를 실행합니다.

```bash
docker run -it --rm \
  -v ~/operator/webapp-operator:/workspace \
  -v ~/.kube:/root/.kube:ro \
  -v /var/run/docker.sock:/var/run/docker.sock \
  operator-dev:1.0
```

각 볼륨의 용도는 다음과 같습니다.

| 호스트                          | 컨테이너                   | 용도                        |
| ---------------------------- | ---------------------- | ------------------------- |
| `~/operator/webapp-operator` | `/workspace`           | Operator 프로젝트 소스 코드       |
| `~/.kube`                    | `/root/.kube:ro`       | Kubernetes 클러스터 접근 설정     |
| `/var/run/docker.sock`       | `/var/run/docker.sock` | 호스트 Docker 데몬을 이용한 이미지 빌드 |

`docker.sock`을 마운트하기 때문에 개발 컨테이너 내부에서 `docker build`를 실행하면 실제 Docker 이미지는 **호스트의 Docker 데몬**에 생성됩니다.

---

## 4. 컨테이너 내부에서 프로젝트 확인

컨테이너에 진입한 후 `/workspace`로 이동합니다.

```bash
cd /workspace
```

현재 위치를 확인합니다.

```bash
pwd
```

결과:

```text
/workspace
```

프로젝트 파일을 확인합니다.

```bash
ls -la
```

또는:

```bash
tree
```

---

## 5. Go 프로젝트 빌드

Operator 소스 코드가 정상적으로 동작하는지 먼저 Go 빌드를 수행합니다.

```bash
go build ./...
```

문제가 없다면 별도의 출력 없이 종료됩니다.

빌드 결과를 확인하려면 다음 명령을 사용할 수 있습니다.

```bash
echo $?
```

결과가:

```text
0
```

이면 빌드가 성공한 것입니다.

---

## 6. Unit Test 실행

Operator의 테스트 코드를 실행합니다.

```bash
go test ./...
```

보다 상세한 테스트 결과를 확인하려면:

```bash
go test -v ./...
```

---

## 7. Operator 바이너리 빌드

프로젝트의 Makefile을 이용하여 Operator 바이너리를 빌드합니다.

```bash
make build
```

정상적으로 빌드되면 일반적으로 `bin/manager`가 생성됩니다.

확인:

```bash
ls -lh bin/manager
```

---

## 8. Docker Operator 이미지 빌드

Kubernetes에 배포할 Operator 이미지를 빌드합니다.

Nexus Registry 주소:

```text
nexus.sysproto.com
```

이미지 경로:

```text
docker-hosted/operator-lab/webapp-operator:1.0
```

컨테이너 내부에서 다음 명령을 실행합니다.

```bash
make docker-build \
  IMG=nexus.sysproto.com/docker-hosted/operator-lab/webapp-operator:1.0
```

전체 이미지 이름:

```text
nexus.sysproto.com/docker-hosted/operator-lab/webapp-operator:1.0
```

빌드가 완료되면 이미지를 확인합니다.

```bash
docker images | grep webapp-operator
```

또는:

```bash
docker image ls \
  nexus.sysproto.com/docker-hosted/operator-lab/webapp-operator
```

> 주의:
>
> 프로젝트 이름은 `webapp-operator`입니다.
>
> 이전에 사용했던 `webapp-oprator`는 `operator`의 오타입니다.
>
> 권장 이미지 이름:
>
> ```text
> nexus.sysproto.com/docker-hosted/operator-lab/webapp-operator:1.0
> ```

---

## 9. Nexus Registry 로그인

Nexus에 이미 로그인되어 있지 않다면 로그인합니다.

```bash
docker login nexus.sysproto.com
```

Nexus 계정 정보를 입력합니다.

로그인 상태를 확인하려면:

```bash
cat ~/.docker/config.json
```

---

## 10. Nexus Registry에 이미지 Push

Operator 이미지 빌드가 완료되면 Nexus Registry에 Push합니다.

```bash
docker push \
  nexus.sysproto.com/docker-hosted/operator-lab/webapp-operator:1.0
```

Push가 완료되면 Nexus에서 다음 이미지가 존재하는지 확인합니다.

```text
nexus.sysproto.com/docker-hosted/operator-lab/webapp-operator:1.0
```

---

## 11. 전체 빌드 과정

### 11.1 호스트에서 개발 이미지 빌드

호스트:

```bash
cd ~/operator

docker build \
  -t operator-dev:1.0 \
  -f dev/Dockerfile \
  dev
```

### 11.2 개발 컨테이너 실행

호스트:

```bash
docker run -it --rm \
  -v ~/operator/webapp-operator:/workspace \
  -v ~/.kube:/root/.kube:ro \
  -v /var/run/docker.sock:/var/run/docker.sock \
  operator-dev:1.0
```

### 11.3 컨테이너 내부에서 프로젝트 디렉터리 이동

컨테이너:

```bash
cd /workspace
```

### 11.4 Go 빌드

컨테이너:

```bash
go build ./...
```

### 11.5 테스트

컨테이너:

```bash
go test ./...
```

### 11.6 Operator 바이너리 빌드

컨테이너:

```bash
make build
```

### 11.7 Operator Docker 이미지 빌드

컨테이너:

```bash
make docker-build \
  IMG=nexus.sysproto.com/docker-hosted/operator-lab/webapp-operator:1.0
```

### 11.8 이미지 확인

컨테이너:

```bash
docker images | grep webapp-operator
```

### 11.9 Nexus 로그인

컨테이너:

```bash
docker login nexus.sysproto.com
```

### 11.10 Nexus Push

컨테이너:

```bash
docker push \
  nexus.sysproto.com/docker-hosted/operator-lab/webapp-operator:1.0
```

---

## 12. Kubernetes에 Operator 배포

이미지가 Nexus에 Push된 후 Kubernetes에 Operator를 배포합니다.

프로젝트 루트:

```bash
cd /workspace
```

기본 배포:

```bash
make deploy \
  IMG=nexus.sysproto.com/docker-hosted/operator-lab/webapp-operator:1.0
```

배포 상태 확인:

```bash
kubectl get pods -n webapp-operator-system
```

Deployment 확인:

```bash
kubectl get deployment -n webapp-operator-system
```

Operator 로그 확인:

```bash
kubectl logs \
  -n webapp-operator-system \
  deployment/webapp-operator-controller-manager \
  -f
```

---

## 13. CRD 확인

WebApp CRD가 설치되었는지 확인합니다.

```bash
kubectl get crd webapps.apps.sysproto.com
```

CRD 상세 정보:

```bash
kubectl describe crd webapps.apps.sysproto.com
```

---

## 14. WebApp Custom Resource 배포

샘플 WebApp 리소스를 배포합니다.

```bash
kubectl apply -f config/samples/apps_v1alpha1_webapp.yaml
```

또는 프로젝트 루트에 있는 샘플 파일을 사용하는 경우:

```bash
kubectl apply -f webapp-sample.yaml
```

WebApp 리소스 확인:

```bash
kubectl get webapps
```

상세 정보:

```bash
kubectl get webapps -o wide
```

Operator가 생성한 Kubernetes 리소스를 확인합니다.

```bash
kubectl get all
```

---

## 15. 개발 작업 시 일반적인 작업 흐름

소스 코드를 수정한 후 다음 순서로 작업합니다.

```text
소스 코드 수정
    │
    ▼
개발 컨테이너 진입
    │
    ▼
cd /workspace
    │
    ▼
go test ./...
    │
    ▼
make docker-build IMG=...
    │
    ▼
docker push ...
    │
    ▼
Kubernetes Operator 이미지 업데이트
    │
    ▼
kubectl 확인
```

개발 컨테이너는 매번 다음 명령으로 실행할 수 있습니다.

```bash
docker run -it --rm \
  -v ~/operator/webapp-operator:/workspace \
  -v ~/.kube:/root/.kube:ro \
  -v /var/run/docker.sock:/var/run/docker.sock \
  operator-dev:1.0
```

컨테이너에 들어간 후:

```bash
cd /workspace

go test ./...

make docker-build \
  IMG=nexus.sysproto.com/docker-hosted/operator-lab/webapp-operator:1.0

docker push \
  nexus.sysproto.com/docker-hosted/operator-lab/webapp-operator:1.0
```

---

## 16. 주요 이미지 및 경로

| 항목                  | 값                                                                   |
| ------------------- | ------------------------------------------------------------------- |
| 프로젝트 루트             | `~/operator`                                                        |
| Operator 소스         | `~/operator/webapp-operator`                                        |
| 개발 Dockerfile       | `~/operator/dev/Dockerfile`                                         |
| Operator Dockerfile | `~/operator/webapp-operator/Dockerfile`                             |
| 컨테이너 프로젝트 경로        | `/workspace`                                                        |
| 개발 이미지              | `operator-dev:1.0`                                                  |
| Nexus Registry      | `nexus.sysproto.com`                                                |
| Operator 이미지        | `nexus.sysproto.com/docker-hosted/operator-lab/webapp-operator:1.0` |

---

## 17. 핵심 명령어 요약

### 개발 이미지 빌드

```yaml
FROM golang:1.24

ARG OPERATOR_SDK_VERSION=v1.41.1
ARG KUBECTL_VERSION=v1.34.1
ARG DOCKER_VERSION=29.6.2

RUN apt-get update && \
    apt-get install -y \
        git \
        make \
        curl \
        ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# kubectl
RUN curl -fsSL -o /usr/local/bin/kubectl \
    https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl && \
    chmod +x /usr/local/bin/kubectl

# operator-sdk
RUN curl -fsSL -o /usr/local/bin/operator-sdk \
    https://github.com/operator-framework/operator-sdk/releases/download/${OPERATOR_SDK_VERSION}/operator-sdk_linux_amd64 && \
    chmod +x /usr/local/bin/operator-sdk

# docker CLI only (no daemon)
RUN curl -fsSL -o /tmp/docker.tgz \
    https://download.docker.com/linux/static/stable/x86_64/docker-${DOCKER_VERSION}.tgz && \
    tar xzvf /tmp/docker.tgz -C /tmp && \
    mv /tmp/docker/docker /usr/local/bin/docker && \
    rm -rf /tmp/docker /tmp/docker.tgz

WORKDIR /workspace
```

```bash
cd ~/operator

docker build \
  -t operator-dev:1.0 \
  -f dev/Dockerfile \
  dev
```

### 개발 컨테이너 실행

```bash
docker run -it --rm \
  -v ~/operator/webapp-operator:/workspace \
  -v ~/.kube:/root/.kube:ro \
  -v /var/run/docker.sock:/var/run/docker.sock \
  operator-dev:1.0
```

### 컨테이너 내부에서 빌드

```bash
cd /workspace

go test ./...

make build

make docker-build \
  IMG=nexus.sysproto.com/docker-hosted/operator-lab/webapp-operator:1.0
```

### Nexus Push

```bash
docker login nexus.sysproto.com

docker push \
  nexus.sysproto.com/docker-hosted/operator-lab/webapp-operator:1.0
```

### Kubernetes 배포

```bash
make deploy \
  IMG=nexus.sysproto.com/docker-hosted/operator-lab/webapp-operator:1.0
```

### 상태 확인

```bash
kubectl get pods -n webapp-operator-system

kubectl get crd webapps.apps.sysproto.com

kubectl get webapps
```

