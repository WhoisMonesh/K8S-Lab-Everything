#!/bin/bash
set -e

# CKA Lab Runner - Complete Demo Script
# This script demonstrates the full workflow of cka-lab-runner

echo "════════════════════════════════════════════════════════"
echo "  CKA Lab Runner - Complete Demo"
echo "════════════════════════════════════════════════════════"
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

demo_step() {
    echo ""
    echo -e "${BLUE}>>> $1${NC}"
    echo ""
}

demo_command() {
    echo -e "${YELLOW}$ $1${NC}"
    eval "$1"
}

# Step 1: Initialize configuration
demo_step "Step 1: Initialize configuration"
demo_command "cka-lab-runner init"
echo ""
echo "Generated config file:"
demo_command "cat cka-lab-runner.yaml"

# Step 2: Create the cluster
demo_step "Step 2: Create local Kubernetes cluster"
echo "This will create a kind cluster with the configured settings..."
demo_command "cka-lab-runner up"

# Step 3: List available labs
demo_step "Step 3: List available labs"
demo_command "cka-lab-runner lab list"

# Step 4: Filter labs by category
demo_step "Step 4: Filter labs by category (control-plane)"
demo_command "cka-lab-runner lab list --category control-plane"

# Step 5: Filter labs by difficulty
demo_step "Step 5: Filter labs by difficulty (easy)"
demo_command "cka-lab-runner lab list --difficulty easy"

# Step 6: Run a lab
demo_step "Step 6: Run the CoreDNS broken config lab"
demo_command "cka-lab-runner lab run coredns_broken_config"

# Step 7: Debug the issue (manual step)
demo_step "Step 7: Now you would debug the issue using kubectl"
echo "Example debugging commands:"
echo "  kubectl get pods -n kube-system"
echo "  kubectl logs -n kube-system -l k8s-app=kube-dns"
echo "  kubectl describe configmap coredns -n kube-system"
echo ""
echo "Press Enter when you want to see the solution..."
read

# Step 8: View solution
demo_step "Step 8: View the canonical solution"
demo_command "cka-lab-runner lab solution coredns_broken_config"

# Step 9: Test another lab
demo_step "Step 9: Run a random lab with fixed seed"
demo_command "cka-lab-runner lab random --seed 42"

# Step 10: View another solution
demo_step "Step 10: View solution for the etcd lab"
demo_command "cka-lab-runner lab solution etcd_wrong_ip"

# Step 11: Clean up
demo_step "Step 11: Clean up the cluster"
demo_command "cka-lab-runner down"

echo ""
echo -e "${GREEN}════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}  Demo Complete!${NC}"
echo -e "${GREEN}════════════════════════════════════════════════════════${NC}"
echo ""
echo "Next steps:"
echo "  - Try running different labs"
echo "  - Practice fixing issues without looking at solutions"
echo "  - Create your own custom labs"
echo ""
