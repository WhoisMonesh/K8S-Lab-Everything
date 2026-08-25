#!/usr/bin/env bash
# K8S-Lab-Everything fully automated setup (macOS / Linux)
# Installs Go, Docker, kubectl, and a cluster provider (kind) automatically.
# Usage: ./scripts/setup.sh [--install-cluster-provider kind|k3d|minikube]

set -euo pipefail

# ──────────────────────────── Colors ────────────────────────────
GREEN='\033[0;32m'; RED='\033[0;31m'; YELLOW='\033[0;33m'; CYAN='\033[0;36m'; BOLD='\033[1m'; NC='\033[0m'

# ──────────────────────────── Helpers ───────────────────────────
info()    { printf "${CYAN}[INFO]${NC}    %s\n" "$*"; }
ok()      { printf "${GREEN}[OK]${NC}      %s\n" "$*"; }
warn()    { printf "${YELLOW}[WARN]${NC}    %s\n" "$*"; }
fail()    { printf "${RED}[FAIL]${NC}    %s\n" "$*"; }
step()    { echo; printf "${BOLD}${CYAN}==> %s${NC}\n" "$*"; }
have()    { command -v "$1" >/dev/null 2>&1; }

# ──────────────────────── Detect OS / Arch ──────────────────────
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
    x86_64)  ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
esac

if [ "$OS" = "darwin" ]; then
    PLATFORM="macOS"
else
    PLATFORM="Linux"
fi

# ──────────────────────── Parse args ────────────────────────────
CLUSTER_PROVIDER="kind"
while [ $# -gt 0 ]; do
    case "$1" in
        --install-cluster-provider) CLUSTER_PROVIDER="$2"; shift 2 ;;
        -h|--help)
            echo "Usage: $0 [--install-cluster-provider kind|k3d|minikube]"
            echo "  Default cluster provider: kind"
            exit 0 ;;
        *) warn "Unknown option: $1"; shift ;;
    esac
done

# ──────────────────────── Detect package manager ────────────────
detect_pkg_manager() {
    if have brew; then
        echo "brew"
    elif have apt-get; then
        echo "apt"
    elif have dnf; then
        echo "dnf"
    elif have yum; then
        echo "yum"
    elif have pacman; then
        echo "pacman"
    elif have apk; then
        echo "apk"
    elif have zypper; then
        echo "zypper"
    else
        echo "unknown"
    fi
}

PKG_MGR=$(detect_pkg_manager)

# ──────────────────────── Install via package manager ───────────
pkg_install() {
    local pkg="$1"
    case "$PKG_MGR" in
        brew)   brew install "$pkg" ;;
        apt)    sudo apt-get update -qq && sudo apt-get install -y -qq "$pkg" ;;
        dnf)    sudo dnf install -y -q "$pkg" ;;
        yum)    sudo yum install -y -q "$pkg" ;;
        pacman) sudo pacman -S --noconfirm "$pkg" ;;
        apk)    sudo apk add --no-cache "$pkg" ;;
        zypper) sudo zypper install -y "$pkg" ;;
        *)      fail "No supported package manager found. Install '$pkg' manually."; return 1 ;;
    esac
}

# ──────────────────────── Homebrew installer ────────────────────
ensure_homebrew() {
    if have brew; then return 0; fi
    info "Installing Homebrew..."
    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
    # Add brew to PATH for Apple Silicon Macs
    if [ "$ARCH" = "arm64" ] && [ "$OS" = "darwin" ]; then
        eval "$(/opt/homebrew/bin/brew shellenv)"
    else
        eval "$(/usr/local/bin/brew shellenv)"
    fi
}

# ──────────────────────── Install Go ────────────────────────────
install_go() {
    if have go; then
        ok "Go $(go version | awk '{print $3}') already installed"
        return
    fi

    step "Installing Go"
    if [ "$OS" = "darwin" ]; then
        ensure_homebrew
        brew install go
    elif [ "$PKG_MGR" = "apt" ]; then
        # Use official tarball for latest version
        GO_VERSION="1.24.7"
        curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${ARCH}.tar.gz" -o /tmp/go.tar.gz
        sudo rm -rf /usr/local/go
        sudo tar -C /usr/local -xzf /tmp/go.tar.gz
        rm /tmp/go.tar.gz
        export PATH="/usr/local/go/bin:$PATH"
        # Add to shell profile
        for profile in "$HOME/.bashrc" "$HOME/.zshrc" "$HOME/.profile" "$HOME/.bash_profile"; do
            if [ -f "$profile" ]; then
                grep -q '/usr/local/go/bin' "$profile" 2>/dev/null || echo 'export PATH="/usr/local/go/bin:$PATH"' >> "$profile"
            fi
        done
    else
        pkg_install go
    fi
    ok "Go installed"
}

# ──────────────────────── Install Docker ────────────────────────
install_docker() {
    if have docker; then
        ok "Docker already installed"
        return
    fi

    step "Installing Docker"
    if [ "$OS" = "darwin" ]; then
        ensure_homebrew
        brew install --cask docker
        # Open Docker Desktop to complete setup
        open -a Docker 2>/dev/null || true
        warn "Docker Desktop installed. Please open Docker Desktop and complete the setup wizard."
        warn "Press Enter once Docker Desktop is running..."
        read -r
    elif [ "$PKG_MGR" = "apt" ]; then
        # Official Docker install script
        curl -fsSL https://get.docker.com | sudo sh
        sudo usermod -aG docker "$USER" 2>/dev/null || true
        warn "You may need to log out and back in for docker group permissions."
    else
        pkg_install docker
        sudo systemctl enable --now docker 2>/dev/null || true
    fi
    ok "Docker installed"
}

# ──────────────────────── Install kubectl ───────────────────────
install_kubectl() {
    if have kubectl; then
        ok "kubectl already installed"
        return
    fi

    step "Installing kubectl"
    if [ "$OS" = "darwin" ]; then
        ensure_homebrew
        brew install kubectl
    elif [ "$PKG_MGR" = "apt" ]; then
        curl -fsSL "https://pkgs.k8s.io/core:/stable:/v1.30/deb/Release.key" | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg 2>/dev/null || true
        echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.30/deb/ /' | sudo tee /etc/apt/sources.list.d/kubernetes.list > /dev/null
        sudo apt-get update -qq
        sudo apt-get install -y -qq kubectl
    elif [ "$PKG_MGR" = "dnf" ]; then
        sudo rpm --import https://pkgs.k8s.io/core:/stable:/v1.30/rpm/repodata/repomd.xml 2>/dev/null || true
        curl -fsSL "https://pkgs.k8s.io/core:/stable:/v1.30/rpm/repodata/repomd.xml" | sudo tee /etc/yum.repos.d/kubernetes.repo > /dev/null
        sudo dnf install -y kubectl
    else
        # Fallback: curl binary
        curl -fsSL "https://dl.k8s.io/release/v1.30.0/bin/${OS}/${ARCH}/kubectl" -o /tmp/kubectl
        sudo install -o root -g root -m 0755 /tmp/kubectl /usr/local/bin/kubectl
        rm /tmp/kubectl
    fi
    ok "kubectl installed"
}

# ──────────────────────── Install cluster provider ──────────────
install_kind() {
    if have kind; then
        ok "kind already installed"
        return
    fi

    step "Installing kind"
    if [ "$OS" = "darwin" ]; then
        ensure_homebrew
        brew install kind
    elif [ "$PKG_MGR" = "apt" ] || [ "$PKG_MGR" = "dnf" ] || [ "$PKG_MGR" = "yum" ]; then
        curl -fsSL "https://kind.sigs.k8s.io/dl/v0.25.0/kind-${OS}-${ARCH}" -o /usr/local/bin/kind
        sudo chmod +x /usr/local/bin/kind
    else
        curl -fsSL "https://kind.sigs.k8s.io/dl/v0.25.0/kind-${OS}-${ARCH}" -o /usr/local/bin/kind
        sudo chmod +x /usr/local/bin/kind
    fi
    ok "kind installed"
}

install_k3d() {
    if have k3d; then
        ok "k3d already installed"
        return
    fi

    step "Installing k3d"
    if [ "$OS" = "darwin" ]; then
        ensure_homebrew
        brew install k3d
    else
        curl -fsSL https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
    fi
    ok "k3d installed"
}

install_minikube() {
    if have minikube; then
        ok "minikube already installed"
        return
    fi

    step "Installing minikube"
    if [ "$OS" = "darwin" ]; then
        ensure_homebrew
        brew install minikube
    elif [ "$PKG_MGR" = "apt" ]; then
        curl -fsSL https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64 -o /usr/local/bin/minikube
        sudo chmod +x /usr/local/bin/minikube
    else
        curl -fsSL https://storage.googleapis.com/minikube/releases/latest/minikube-${OS}-${ARCH} -o /usr/local/bin/minikube
        sudo chmod +x /usr/local/bin/minikube
    fi
    ok "minikube installed"
}

install_cluster_provider() {
    case "$CLUSTER_PROVIDER" in
        kind)     install_kind ;;
        k3d)      install_k3d ;;
        minikube) install_minikube ;;
        *)        fail "Unknown cluster provider: $CLUSTER_PROVIDER"; exit 1 ;;
    esac
}

# ──────────────────────── Build binary ──────────────────────────
build_binary() {
    if [ ! -f "./go.mod" ]; then
        warn "go.mod not found — skipping binary build"
        return
    fi

    step "Building cka-lab-runner"
    if have go; then
        mkdir -p bin
        go build -o bin/cka-lab-runner -v ./cmd/cka-lab-runner
        ok "Binary built: bin/cka-lab-runner"
    else
        warn "Go not available — cannot build binary"
        return
    fi

    step "Installing cka-lab-runner to /usr/local/bin"
    sudo cp bin/cka-lab-runner /usr/local/bin/cka-lab-runner
    sudo chmod +x /usr/local/bin/cka-lab-runner
    ok "Installed to /usr/local/bin/cka-lab-runner (available system-wide)"
}

# ──────────────────────── Main ──────────────────────────────────
echo
printf "${BOLD}${CYAN}══════════════════════════════════════════════════════════${NC}\n"
printf "${BOLD}${CYAN}  K8S-Lab-Everything — Automated Setup ($PLATFORM $ARCH)${NC}\n"
printf "${BOLD}${CYAN}══════════════════════════════════════════════════════════${NC}\n"
echo
info "Package manager: $PKG_MGR"
info "Cluster provider: $CLUSTER_PROVIDER"
echo

install_go
install_docker
install_kubectl
install_cluster_provider
build_binary

echo
printf "${BOLD}${GREEN}══════════════════════════════════════════════════════════${NC}\n"
printf "${BOLD}${GREEN}  All prerequisites installed!${NC}\n"
printf "${BOLD}${GREEN}══════════════════════════════════════════════════════════${NC}\n"
echo
echo "Next steps:"
echo "  cka-lab-runner init"
echo "  cka-lab-runner up"
echo "  cka-lab-runner lab list"
echo "  cka-lab-runner lab run pod_crashloop"
echo
