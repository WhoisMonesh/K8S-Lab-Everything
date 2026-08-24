#!/usr/bin/env bash
# K8S-Lab-Everything prerequisite checker (macOS / Linux)
# Usage: ./scripts/setup.sh

set -u

GREEN='\033[0;32m'; RED='\033[0;31m'; YELLOW='\033[0;33m'; CYAN='\033[0;36m'; NC='\033[0m'
missing=0

have() { command -v "$1" >/dev/null 2>&1; }

check() {
    local name="$1"; shift
    if have "$name"; then
        printf "${GREEN}[OK]${NC}      %s\n" "$name"
    else
        printf "${RED}[MISSING]${NC} %s\n" "$name"
        printf "${YELLOW}          Install:%s\n" ""
        for hint in "$@"; do
            printf "${YELLOW}            %s${NC}\n" "$hint"
        done
        missing=$((missing+1))
    fi
}

echo
printf "${CYAN}== K8S-Lab-Everything setup check ==${NC}\n"
echo

check "go"      "brew install go   |   or https://go.dev/dl/"
check "docker"  "brew install --cask docker (macOS)   |   curl -fsSL https://get.docker.com | sh (Linux)"
check "kubectl" "brew install kubectl   |   or https://kubernetes.io/docs/tasks/tools/"

if have kind || have k3d || have minikube; then
    printf "${GREEN}[OK]${NC}      cluster provider (kind / k3d / minikube)\n"
else
    printf "${RED}[MISSING]${NC} cluster provider - install at least one:\n"
    printf "${YELLOW}            brew install kind   |   brew install k3d   |   brew install minikube${NC}\n"
    missing=$((missing+1))
fi

echo

if [ "$missing" -eq 0 ]; then
    printf "${GREEN}All prerequisites installed! Next steps:${NC}\n"
    echo "  go install ./cmd/cka-lab-runner"
    echo "  cka-lab-runner init"
    echo "  cka-lab-runner up"
    echo "  cka-lab-runner lab list"
else
    printf "${YELLOW}%s prerequisite(s) missing - install them, then re-run this script.${NC}\n" "$missing"
fi
echo
