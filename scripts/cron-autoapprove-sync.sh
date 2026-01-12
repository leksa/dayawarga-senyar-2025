#!/bin/bash
# =============================================================================
# Dayawarga Auto-Approve & Sync Cron Script
# Script untuk auto-approve submissions di ODK Central dan sync ke PostgreSQL
#
# Setup crontab (run every 5 minutes):
#   */5 * * * * /path/to/cron-autoapprove-sync.sh >> /var/log/dayawarga-sync.log 2>&1
# =============================================================================

set -e

# Change to script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Configuration - load from environment or use defaults
ODK_CENTRAL_URL="${ODK_CENTRAL_URL:-https://data.dayawarga.com}"
ODK_PROJECT_ID="${ODK_PROJECT_ID:-3}"
ODK_EMAIL="${ODK_EMAIL:-}"
ODK_PASSWORD="${ODK_PASSWORD:-}"
API_BASE_URL="${API_BASE_URL:-https://api.dayawarga.com/api/v1}"
SYNC_API_KEY="${SYNC_API_KEY:-}"

# Forms to auto-approve
FORMS=("form_posko_v1" "form_feed_v1" "form_faskes_v1" "form_jembatan_v1")

# Log with timestamp
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

log_error() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] ERROR: $1" >&2
}

# Check required env vars
check_env() {
    if [ -z "$ODK_EMAIL" ] || [ -z "$ODK_PASSWORD" ]; then
        log_error "ODK_EMAIL and ODK_PASSWORD must be set"
        exit 1
    fi
    if [ -z "$SYNC_API_KEY" ]; then
        log_error "SYNC_API_KEY must be set"
        exit 1
    fi
}

# Auto-approve pending submissions for a form
approve_form() {
    local form_id="$1"
    log "Approving pending submissions for form: $form_id"

    # Run Python approve script
    if [ -f "$SCRIPT_DIR/approve_submissions.py" ]; then
        ODK_FORM_ID="$form_id" python3 "$SCRIPT_DIR/approve_submissions.py" --include-edited 2>&1 || {
            log_error "Failed to approve submissions for $form_id"
            return 1
        }
    else
        log_error "approve_submissions.py not found"
        return 1
    fi
}

# Sync data from ODK Central to PostgreSQL
sync_data() {
    local endpoint="$1"
    local name="$2"

    log "Syncing $name..."

    response=$(curl -sf -X POST "${API_BASE_URL}/sync/${endpoint}" \
        -H "X-API-Key: ${SYNC_API_KEY}" \
        -H "Content-Type: application/json" 2>&1) || {
        log_error "Failed to sync $name"
        return 1
    }

    if echo "$response" | jq -e '.success == true' > /dev/null 2>&1; then
        created=$(echo "$response" | jq -r '.data.created // 0')
        updated=$(echo "$response" | jq -r '.data.updated // 0')
        duration=$(echo "$response" | jq -r '.data.duration // "N/A"')
        log "$name sync completed: created=$created, updated=$updated, duration=$duration"
    else
        log_error "$name sync failed: $response"
        return 1
    fi
}

# Main execution
main() {
    log "=========================================="
    log "Starting auto-approve and sync process"
    log "=========================================="

    check_env

    # Step 1: Auto-approve all forms
    log "--- Step 1: Auto-approving submissions ---"
    for form in "${FORMS[@]}"; do
        approve_form "$form" || true  # Continue even if one form fails
    done

    # Step 2: Sync all data
    log "--- Step 2: Syncing data to PostgreSQL ---"
    sync_data "posko" "Posko"
    sync_data "feed" "Feed"
    sync_data "faskes" "Faskes"
    sync_data "infrastruktur" "Infrastruktur"

    # Step 3: Sync photos
    log "--- Step 3: Syncing photos ---"
    sync_data "photos" "Posko Photos"
    sync_data "feed-photos" "Feed Photos"
    sync_data "faskes-photos" "Faskes Photos"

    log "=========================================="
    log "Auto-approve and sync process completed"
    log "=========================================="
}

main "$@"
