#!/bin/bash

# =============================================================================
# SSL CERTIFICATE EXTRACTION SCRIPT
# =============================================================================
# Extracts SSL certificates from nginx-proxy-manager container for web UI setup

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Default container runtime
CONTAINER_CMD="sudo podman"

print_header() {
    echo -e "${BLUE}======================================${NC}"
    echo -e "${BLUE}  SSL CERTIFICATE EXTRACTION TOOL${NC}"
    echo -e "${BLUE}======================================${NC}"
    echo ""
}

print_section() {
    echo -e "${CYAN}$1${NC}"
    echo -e "${CYAN}$(printf '=%.0s' $(seq 1 ${#1}))${NC}"
}

check_container() {
    echo -e "${YELLOW}Checking nginx-proxy-manager container...${NC}"
    
    if ! $CONTAINER_CMD container exists nginx-proxy-manager 2>/dev/null; then
        echo -e "${RED}Error: nginx-proxy-manager container not found!${NC}"
        echo "Please start the proxy services first:"
        echo "  ./scripts/start-services.sh -c proxy"
        exit 1
    fi
    
    local status=$($CONTAINER_CMD inspect --format='{{.State.Status}}' nginx-proxy-manager 2>/dev/null)
    if [[ "$status" != "running" ]]; then
        echo -e "${RED}Error: nginx-proxy-manager container is not running (status: $status)${NC}"
        echo "Please start the container first:"
        echo "  $CONTAINER_CMD start nginx-proxy-manager"
        exit 1
    fi
    
    echo -e "${GREEN}✓ Container is running${NC}"
    echo ""
}

check_certificates() {
    echo -e "${YELLOW}Checking for SSL certificates...${NC}"
    
    if ! $CONTAINER_CMD exec nginx-proxy-manager test -f /etc/letsencrypt/live/wildcard.test/fullchain.pem 2>/dev/null; then
        echo -e "${RED}Error: Certificate not found in container!${NC}"
        echo "Please run the SSL setup script first:"
        echo "  ./scripts/setup-ssl.sh -s -e"
        exit 1
    fi
    
    if ! $CONTAINER_CMD exec nginx-proxy-manager test -f /etc/letsencrypt/live/wildcard.test/privkey.pem 2>/dev/null; then
        echo -e "${RED}Error: Private key not found in container!${NC}"
        echo "Please run the SSL setup script first:"
        echo "  ./scripts/setup-ssl.sh -s -e"
        exit 1
    fi
    
    echo -e "${GREEN}✓ Certificates found${NC}"
    echo ""
}

extract_and_display() {
    print_section "PRIVATE KEY (Certificate Key Field)"
    echo -e "${YELLOW}Copy everything between and including the BEGIN/END lines:${NC}"
    echo ""
    echo -e "${GREEN}--- START COPYING FROM HERE ---${NC}"
    $CONTAINER_CMD exec nginx-proxy-manager cat /etc/letsencrypt/live/wildcard.test/privkey.pem 2>/dev/null
    echo -e "${GREEN}--- STOP COPYING HERE ---${NC}"
    echo ""
    echo ""
    
    print_section "CERTIFICATE (Certificate Field)"
    echo -e "${YELLOW}Copy everything between and including the BEGIN/END lines:${NC}"
    echo ""
    echo -e "${GREEN}--- START COPYING FROM HERE ---${NC}"
    $CONTAINER_CMD exec nginx-proxy-manager cat /etc/letsencrypt/live/wildcard.test/fullchain.pem 2>/dev/null
    echo -e "${GREEN}--- STOP COPYING HERE ---${NC}"
    echo ""
}

save_to_files() {
    local save_dir="$1"
    mkdir -p "$save_dir"
    
    echo -e "${YELLOW}Saving certificates to files...${NC}"
    
    # Save private key
    $CONTAINER_CMD exec nginx-proxy-manager cat /etc/letsencrypt/live/wildcard.test/privkey.pem > "$save_dir/privkey.pem" 2>/dev/null
    if [[ $? -eq 0 ]]; then
        echo -e "${GREEN}✓ Private key saved to: $save_dir/privkey.pem${NC}"
    else
        echo -e "${RED}✗ Failed to save private key${NC}"
    fi
    
    # Save certificate
    $CONTAINER_CMD exec nginx-proxy-manager cat /etc/letsencrypt/live/wildcard.test/fullchain.pem > "$save_dir/fullchain.pem" 2>/dev/null
    if [[ $? -eq 0 ]]; then
        echo -e "${GREEN}✓ Certificate saved to: $save_dir/fullchain.pem${NC}"
    else
        echo -e "${RED}✗ Failed to save certificate${NC}"
    fi
    
    # Create a combined file for convenience
    if [[ -f "$save_dir/privkey.pem" && -f "$save_dir/fullchain.pem" ]]; then
        cat "$save_dir/fullchain.pem" "$save_dir/privkey.pem" > "$save_dir/combined.pem"
        echo -e "${GREEN}✓ Combined certificate saved to: $save_dir/combined.pem${NC}"
    fi
    
    echo ""
}

show_npm_instructions() {
    print_section "NGINX PROXY MANAGER SETUP INSTRUCTIONS"
    echo ""
    echo -e "${YELLOW}1. Open Nginx Proxy Manager in your browser:${NC}"
    echo "   http://localhost:81"
    echo ""
    echo -e "${YELLOW}2. Login with default credentials:${NC}"
    echo "   Email: admin@example.com"
    echo "   Password: changeme"
    echo "   (You'll be prompted to change these on first login)"
    echo ""
    echo -e "${YELLOW}3. Add SSL Certificate:${NC}"
    echo "   • Go to 'SSL Certificates' tab"
    echo "   • Click 'Add SSL Certificate'"
    echo "   • Select 'Custom'"
    echo "   • Name: wildcard.test"
    echo "   • Certificate Key: Paste the PRIVATE KEY from above"
    echo "   • Certificate: Paste the CERTIFICATE from above"
    echo "   • Click 'Save'"
    echo ""
    echo -e "${YELLOW}4. Create Proxy Hosts (example for Grafana):${NC}"
    echo "   • Go to 'Hosts' → 'Proxy Hosts'"
    echo "   • Click 'Add Proxy Host'"
    echo "   • Domain Names: grafana.test"
    echo "   • Forward Hostname/IP: grafana"
    echo "   • Forward Port: 3000"
    echo "   • Go to 'SSL' tab"
    echo "   • SSL Certificate: Select 'wildcard.test'"
    echo "   • Force SSL: ✓ Enable"
    echo "   • HTTP/2 Support: ✓ Enable"
    echo "   • Click 'Save'"
    echo ""
    echo -e "${YELLOW}5. Test your setup:${NC}"
    echo "   curl -k https://grafana.test"
    echo ""
}

show_quick_services() {
    print_section "QUICK SERVICE SETUP"
    echo ""
    echo -e "${YELLOW}Common services to set up with wildcard.test certificate:${NC}"
    echo ""
    cat << 'EOF'
| Service     | Domain          | Forward Host | Port |
|-------------|-----------------|--------------|------|
| Grafana     | grafana.test    | grafana      | 3000 |
| Metabase    | metabase.test   | metabase     | 3000 |
| Adminer     | adminer.test    | adminer      | 8080 |
| phpMyAdmin  | phpmyadmin.test | phpmyadmin   | 80   |
| Mongo Exp   | mongodb.test    | mongo-express| 8081 |
| n8n         | n8n.test        | n8n          | 5678 |
| Langflow    | langflow.test   | langflow     | 7860 |
| Kibana      | kibana.test     | kibana       | 5601 |
| Mailpit     | mailpit.test    | mailpit      | 8025 |
| Gitea       | gitea.test      | gitea        | 3000 |
| NocoDB      | nocodb.test     | nocodb       | 8080 |
| pgAdmin     | pgadmin.test    | pgadmin      | 80   |
EOF
    echo ""
}

usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "OPTIONS:"
    echo "  -s, --save DIR    Save certificates to directory"
    echo "  -d, --docker      Use docker instead of podman"
    echo "  -q, --quiet       Only show certificates (no instructions)"
    echo "  -h, --help        Show this help"
    echo ""
    echo "EXAMPLES:"
    echo "  $0                      # Display certificates and instructions"
    echo "  $0 -s ~/ssl-certs       # Save certificates to ~/ssl-certs"
    echo "  $0 -q                   # Only show certificates"
    echo "  $0 -d                   # Use docker instead of podman"
}

# Parse command line arguments
SAVE_DIR=""
QUIET=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -s|--save)
            SAVE_DIR="$2"
            shift 2
            ;;
        -d|--docker)
            CONTAINER_CMD="sudo docker"
            shift
            ;;
        -q|--quiet)
            QUIET=true
            shift
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Main execution
main() {
    if [[ "$QUIET" == "false" ]]; then
        print_header
    fi
    
    check_container
    check_certificates
    extract_and_display
    
    if [[ -n "$SAVE_DIR" ]]; then
        save_to_files "$SAVE_DIR"
    fi
    
    if [[ "$QUIET" == "false" ]]; then
        show_npm_instructions
        show_quick_services
        
        echo -e "${BLUE}======================================${NC}"
        echo -e "${GREEN}Certificate extraction completed!${NC}"
        echo -e "${BLUE}======================================${NC}"
    fi
}

main "$@"