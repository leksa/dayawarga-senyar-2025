#!/bin/bash
# =============================================================================
# Dayawarga Sync Script
# Script untuk sinkronisasi data dari ODK Central ke PostgreSQL
# =============================================================================

set -e

# Configuration
API_BASE_URL="${API_BASE_URL:-https://api.dayawarga.com/api/v1}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
print_header() {
    echo ""
    echo -e "${BLUE}============================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}============================================${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}→ $1${NC}"
}

check_api() {
    print_info "Checking API health..."
    if curl -sf "${API_BASE_URL%/api/v1}/health" > /dev/null 2>&1; then
        print_success "API is healthy"
        return 0
    else
        print_error "API is not responding at ${API_BASE_URL}"
        return 1
    fi
}

sync_posko() {
    print_header "Syncing Posko (Locations)"
    print_info "Fetching approved submissions from ODK Central..."

    response=$(curl -sf -X POST "${API_BASE_URL}/sync/posko" 2>&1) || {
        print_error "Failed to sync posko"
        echo "$response"
        return 1
    }

    if echo "$response" | jq -e '.success == true' > /dev/null 2>&1; then
        created=$(echo "$response" | jq -r '.data.created // 0')
        updated=$(echo "$response" | jq -r '.data.updated // 0')
        skipped=$(echo "$response" | jq -r '.data.skipped // 0')
        errors=$(echo "$response" | jq -r '.data.errors // 0')
        duration=$(echo "$response" | jq -r '.data.duration // "N/A"')

        print_success "Posko sync completed in ${duration}"
        echo "   Created: ${created}, Updated: ${updated}, Skipped: ${skipped}, Errors: ${errors}"
    else
        print_error "Posko sync failed"
        echo "$response" | jq .
        return 1
    fi
}

sync_feeds() {
    print_header "Syncing Feeds (Information Updates)"
    print_info "Fetching approved feed submissions from ODK Central..."

    response=$(curl -sf -X POST "${API_BASE_URL}/sync/feed" 2>&1) || {
        print_error "Failed to sync feeds"
        echo "$response"
        return 1
    }

    if echo "$response" | jq -e '.success == true' > /dev/null 2>&1; then
        created=$(echo "$response" | jq -r '.data.created // 0')
        updated=$(echo "$response" | jq -r '.data.updated // 0')
        skipped=$(echo "$response" | jq -r '.data.skipped // 0')
        errors=$(echo "$response" | jq -r '.data.errors // 0')
        duration=$(echo "$response" | jq -r '.data.duration // "N/A"')

        print_success "Feed sync completed in ${duration}"
        echo "   Created: ${created}, Updated: ${updated}, Skipped: ${skipped}, Errors: ${errors}"
    else
        print_error "Feed sync failed"
        echo "$response" | jq .
        return 1
    fi
}

sync_faskes() {
    print_header "Syncing Faskes (Health Facilities)"
    print_info "Fetching approved faskes submissions from ODK Central..."

    response=$(curl -sf -X POST "${API_BASE_URL}/sync/faskes" 2>&1) || {
        print_error "Failed to sync faskes"
        echo "$response"
        return 1
    }

    if echo "$response" | jq -e '.success == true' > /dev/null 2>&1; then
        created=$(echo "$response" | jq -r '.data.created // 0')
        updated=$(echo "$response" | jq -r '.data.updated // 0')
        skipped=$(echo "$response" | jq -r '.data.skipped // 0')
        errors=$(echo "$response" | jq -r '.data.errors // 0')
        duration=$(echo "$response" | jq -r '.data.duration // "N/A"')

        print_success "Faskes sync completed in ${duration}"
        echo "   Created: ${created}, Updated: ${updated}, Skipped: ${skipped}, Errors: ${errors}"
    else
        print_error "Faskes sync failed"
        echo "$response" | jq .
        return 1
    fi
}

sync_photos() {
    print_header "Syncing Photos (Posko)"
    print_info "Downloading uncached photos to S3..."

    response=$(curl -sf -X POST "${API_BASE_URL}/sync/photos" 2>&1) || {
        print_error "Failed to sync posko photos"
        echo "$response"
        return 1
    }

    if echo "$response" | jq -e '.success == true' > /dev/null 2>&1; then
        total=$(echo "$response" | jq -r '.data.total_found // 0')
        downloaded=$(echo "$response" | jq -r '.data.downloaded // 0')
        errors=$(echo "$response" | jq -r '.data.errors // 0')
        duration=$(echo "$response" | jq -r '.data.duration // "N/A"')

        print_success "Photo sync completed in ${duration}"
        echo "   Found: ${total}, Downloaded: ${downloaded}, Errors: ${errors}"
    else
        print_error "Photo sync failed"
        echo "$response" | jq .
        return 1
    fi
}

sync_feed_photos() {
    print_header "Syncing Feed Photos"
    print_info "Downloading uncached feed photos to S3..."

    response=$(curl -sf -X POST "${API_BASE_URL}/sync/feed-photos" 2>&1) || {
        print_error "Failed to sync feed photos"
        echo "$response"
        return 1
    }

    if echo "$response" | jq -e '.success == true' > /dev/null 2>&1; then
        total=$(echo "$response" | jq -r '.data.total_found // 0')
        downloaded=$(echo "$response" | jq -r '.data.downloaded // 0')
        errors=$(echo "$response" | jq -r '.data.errors // 0')
        duration=$(echo "$response" | jq -r '.data.duration // "N/A"')

        print_success "Feed photo sync completed in ${duration}"
        echo "   Found: ${total}, Downloaded: ${downloaded}, Errors: ${errors}"
    else
        print_error "Feed photo sync failed"
        echo "$response" | jq .
        return 1
    fi
}

sync_faskes_photos() {
    print_header "Syncing Faskes Photos"
    print_info "Downloading uncached faskes photos to S3..."

    response=$(curl -sf -X POST "${API_BASE_URL}/sync/faskes-photos" 2>&1) || {
        print_error "Failed to sync faskes photos"
        echo "$response"
        return 1
    }

    if echo "$response" | jq -e '.success == true' > /dev/null 2>&1; then
        total=$(echo "$response" | jq -r '.data.total_found // 0')
        downloaded=$(echo "$response" | jq -r '.data.downloaded // 0')
        errors=$(echo "$response" | jq -r '.data.errors // 0')
        duration=$(echo "$response" | jq -r '.data.duration // "N/A"')

        print_success "Faskes photo sync completed in ${duration}"
        echo "   Found: ${total}, Downloaded: ${downloaded}, Errors: ${errors}"
    else
        print_error "Faskes photo sync failed"
        echo "$response" | jq .
        return 1
    fi
}

# ============================================
# HARD SYNC FUNCTIONS (sync + delete orphans)
# ============================================

hard_sync_posko() {
    print_header "Hard Syncing Posko (Locations)"
    print_info "Syncing and deleting records not in ODK Central..."

    response=$(curl -sf -X POST "${API_BASE_URL}/sync/posko/hard" 2>&1) || {
        print_error "Failed to hard sync posko"
        echo "$response"
        return 1
    }

    if echo "$response" | jq -e '.success == true' > /dev/null 2>&1; then
        created=$(echo "$response" | jq -r '.data.created // 0')
        updated=$(echo "$response" | jq -r '.data.updated // 0')
        deleted=$(echo "$response" | jq -r '.data.deleted // 0')
        errors=$(echo "$response" | jq -r '.data.errors // 0')
        duration=$(echo "$response" | jq -r '.data.duration // "N/A"')

        print_success "Posko hard sync completed in ${duration}"
        echo "   Created: ${created}, Updated: ${updated}, Deleted: ${deleted}, Errors: ${errors}"
    else
        print_error "Posko hard sync failed"
        echo "$response" | jq .
        return 1
    fi
}

hard_sync_feeds() {
    print_header "Hard Syncing Feeds"
    print_info "Syncing and deleting records not in ODK Central..."

    response=$(curl -sf -X POST "${API_BASE_URL}/sync/feed/hard" 2>&1) || {
        print_error "Failed to hard sync feeds"
        echo "$response"
        return 1
    }

    if echo "$response" | jq -e '.success == true' > /dev/null 2>&1; then
        created=$(echo "$response" | jq -r '.data.created // 0')
        updated=$(echo "$response" | jq -r '.data.updated // 0')
        deleted=$(echo "$response" | jq -r '.data.deleted // 0')
        errors=$(echo "$response" | jq -r '.data.errors // 0')
        duration=$(echo "$response" | jq -r '.data.duration // "N/A"')

        print_success "Feed hard sync completed in ${duration}"
        echo "   Created: ${created}, Updated: ${updated}, Deleted: ${deleted}, Errors: ${errors}"
    else
        print_error "Feed hard sync failed"
        echo "$response" | jq .
        return 1
    fi
}

hard_sync_faskes() {
    print_header "Hard Syncing Faskes"
    print_info "Syncing and deleting records not in ODK Central..."

    response=$(curl -sf -X POST "${API_BASE_URL}/sync/faskes/hard" 2>&1) || {
        print_error "Failed to hard sync faskes"
        echo "$response"
        return 1
    }

    if echo "$response" | jq -e '.success == true' > /dev/null 2>&1; then
        created=$(echo "$response" | jq -r '.data.created // 0')
        updated=$(echo "$response" | jq -r '.data.updated // 0')
        deleted=$(echo "$response" | jq -r '.data.deleted // 0')
        errors=$(echo "$response" | jq -r '.data.errors // 0')
        duration=$(echo "$response" | jq -r '.data.duration // "N/A"')

        print_success "Faskes hard sync completed in ${duration}"
        echo "   Created: ${created}, Updated: ${updated}, Deleted: ${deleted}, Errors: ${errors}"
    else
        print_error "Faskes hard sync failed"
        echo "$response" | jq .
        return 1
    fi
}

show_status() {
    print_header "Sync Status"

    echo ""
    print_info "Posko sync status:"
    curl -sf "${API_BASE_URL}/sync/status" | jq '.data | {status, last_sync_time, total_records}' 2>/dev/null || echo "N/A"

    echo ""
    print_info "Feed sync status:"
    curl -sf "${API_BASE_URL}/sync/feed/status" | jq '.data | {status, last_sync_time, total_records}' 2>/dev/null || echo "N/A"

    echo ""
    print_info "Faskes sync status:"
    curl -sf "${API_BASE_URL}/sync/faskes/status" | jq '.data | {status, last_sync_time, total_records}' 2>/dev/null || echo "N/A"
}

show_help() {
    echo "Dayawarga Sync Script"
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  all             Sync everything (posko, feeds, faskes, and all photos)"
    echo "  posko           Sync posko/locations only"
    echo "  feeds           Sync feeds only"
    echo "  faskes          Sync faskes only"
    echo "  photos          Sync all photos (posko, feed, faskes)"
    echo "  photos-posko    Sync posko photos only"
    echo "  photos-feed     Sync feed photos only"
    echo "  photos-faskes   Sync faskes photos only"
    echo "  status          Show sync status"
    echo ""
    echo "Hard Sync Commands (sync + delete orphaned records):"
    echo "  hard            Hard sync all (posko, feeds, faskes) - DELETES orphans"
    echo "  hard-posko      Hard sync posko only - DELETES orphans"
    echo "  hard-feeds      Hard sync feeds only - DELETES orphans"
    echo "  hard-faskes     Hard sync faskes only - DELETES orphans"
    echo ""
    echo "  help            Show this help message"
    echo ""
    echo "Environment Variables:"
    echo "  API_BASE_URL  API base URL (default: https://api.dayawarga.com/api/v1)"
    echo ""
    echo "Examples:"
    echo "  $0 all                    # Sync everything"
    echo "  $0 feeds                  # Sync feeds only"
    echo "  $0 hard                   # Hard sync all (deletes orphaned records)"
    echo "  $0 hard-feeds             # Hard sync feeds only"
    echo "  API_BASE_URL=http://localhost:8080/api/v1 $0 all  # Use local API"
}

# Main
main() {
    echo ""
    echo -e "${BLUE}Dayawarga Sync Script${NC}"
    echo -e "API: ${API_BASE_URL}"
    echo ""

    # Check dependencies
    if ! command -v curl &> /dev/null; then
        print_error "curl is required but not installed"
        exit 1
    fi

    if ! command -v jq &> /dev/null; then
        print_error "jq is required but not installed"
        exit 1
    fi

    case "${1:-all}" in
        all)
            check_api || exit 1
            sync_posko
            sync_feeds
            sync_faskes
            sync_photos
            sync_feed_photos
            sync_faskes_photos
            print_header "Sync Complete!"
            ;;
        posko)
            check_api || exit 1
            sync_posko
            ;;
        feeds)
            check_api || exit 1
            sync_feeds
            ;;
        faskes)
            check_api || exit 1
            sync_faskes
            ;;
        photos)
            check_api || exit 1
            sync_photos
            sync_feed_photos
            sync_faskes_photos
            ;;
        photos-posko)
            check_api || exit 1
            sync_photos
            ;;
        photos-feed)
            check_api || exit 1
            sync_feed_photos
            ;;
        photos-faskes)
            check_api || exit 1
            sync_faskes_photos
            ;;
        hard)
            check_api || exit 1
            hard_sync_posko
            hard_sync_feeds
            hard_sync_faskes
            sync_photos
            sync_feed_photos
            sync_faskes_photos
            print_header "Hard Sync Complete!"
            ;;
        hard-posko)
            check_api || exit 1
            hard_sync_posko
            ;;
        hard-feeds)
            check_api || exit 1
            hard_sync_feeds
            ;;
        hard-faskes)
            check_api || exit 1
            hard_sync_faskes
            ;;
        status)
            check_api || exit 1
            show_status
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            print_error "Unknown command: $1"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

main "$@"
