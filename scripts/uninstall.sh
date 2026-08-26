#!/usr/bin/env bash
set -euo pipefail

# Colors
RED='\033[31m'
GREEN='\033[32m'
YELLOW='\033[33m'
CYAN='\033[36m'
BOLD='\033[1m'
DIM='\033[2m'
RESET='\033[0m'

echo ""
printf "  ${CYAN}${BOLD}┌─────────────────────────────────────────────────────────────┐${RESET}\n"
printf "  ${CYAN}${BOLD}│${RESET}  ${BOLD}K8S-Lab-Everything — Uninstaller${RESET}                          ${CYAN}${BOLD}│${RESET}\n"
printf "  ${CYAN}${BOLD}└─────────────────────────────────────────────────────────────┘${RESET}\n"
echo ""

# Items to remove
ITEMS=(
    "/usr/local/bin/cka-lab-runner"
    "$HOME/.cka-lab-runner"
    "$HOME/Library/Application Support/cka-lab-runner"
)

DRY_RUN=false
FORCE=false

# Parse args
while [[ $# -gt 0 ]]; do
    case $1 in
        --dry-run) DRY_RUN=true; shift ;;
        --force|-f) FORCE=true; shift ;;
        --help|-h)
            echo "  Usage: ./uninstall.sh [OPTIONS]"
            echo ""
            echo "  Options:"
            echo "    --dry-run    Show what would be removed without deleting"
            echo "    --force, -f  Skip confirmation prompt"
            echo "    --help, -h   Show this help"
            echo ""
            exit 0
            ;;
        *) echo "  Unknown option: $1"; exit 1 ;;
    esac
done

# Find things to remove
FOUND=()

# Binary
if [[ -f /usr/local/bin/cka-lab-runner ]]; then
    FOUND+=("/usr/local/bin/cka-lab-runner")
fi

# Config directory
if [[ -d "$HOME/.cka-lab-runner" ]]; then
    FOUND+=("$HOME/.cka-lab-runner")
fi

# macOS app support
if [[ -d "$HOME/Library/Application Support/cka-lab-runner" ]]; then
    FOUND+=("$HOME/Library/Application Support/cka-lab-runner")
fi

# Linux config
if [[ -d "$HOME/.config/cka-lab-runner" ]]; then
    FOUND+=("$HOME/.config/cka-lab-runner")
fi

# Local progress files (search current and common dirs)
while IFS= read -r f; do
    FOUND+=("$f")
done < <(find "$HOME" -maxdepth 3 -name ".lab-progress.json" -o -name "cka-lab-runner.yaml" 2>/dev/null | head -20)

if [[ ${#FOUND[@]} -eq 0 ]]; then
    printf "\n  ${GREEN}✔${RESET}  ${BOLD}Nothing to remove — cka-lab-runner is not installed${RESET}\n\n"
    exit 0
fi

# Show what will be removed
printf "\n  ${BOLD}Items to remove:${RESET}\n\n"
for item in "${FOUND[@]}"; do
    if [[ -f "$item" ]]; then
        printf "    ${DIM}file${RESET}   %s\n" "$item"
    elif [[ -d "$item" ]]; then
        printf "    ${DIM}dir${RESET}    %s\n" "$item"
    fi
done
echo ""

# Dry run exit
if [[ "$DRY_RUN" == true ]]; then
    printf "  ${YELLOW}${BOLD}Dry run — nothing was deleted${RESET}\n\n"
    exit 0
fi

# Confirm
if [[ "$FORCE" == false ]]; then
    printf "  ${YELLOW}${BOLD}⚠  Remove all listed items? [y/N]${RESET} "
    read -r ans
    if [[ ! "$ans" =~ ^[Yy]$ ]]; then
        printf "\n  ${DIM}Cancelled.${RESET}\n\n"
        exit 0
    fi
fi

echo ""

# Remove
for item in "${FOUND[@]}"; do
    if [[ -f "$item" || -L "$item" ]]; then
        if sudo rm -f "$item" 2>/dev/null; then
            printf "  ${GREEN}✔${RESET}  Removed %s\n" "$item"
        else
            printf "  ${RED}✖${RESET}  Failed to remove %s (try with sudo)\n" "$item"
        fi
    elif [[ -d "$item" ]]; then
        if sudo rm -rf "$item" 2>/dev/null; then
            printf "  ${GREEN}✔${RESET}  Removed %s\n" "$item"
        else
            printf "  ${RED}✖${RESET}  Failed to remove %s (try with sudo)\n" "$item"
        fi
    fi
done

echo ""
printf "  ${GREEN}${BOLD}✔  cka-lab-runner has been uninstalled${RESET}\n\n"
