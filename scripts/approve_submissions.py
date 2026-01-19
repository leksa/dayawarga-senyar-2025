#!/usr/bin/env python3
"""
Script to approve pending submissions in ODK Central
Supports multiple forms: posko, feeds, faskes
Usage: python approve_submissions.py [--dry-run] [--limit N] [--form-id FORM]
"""

import os
import sys
import argparse
import requests
from pathlib import Path

# Load .env file if exists
ENV_FILE = Path(__file__).parent.parent / ".env"
if ENV_FILE.exists():
    with open(ENV_FILE) as f:
        for line in f:
            line = line.strip()
            if line and not line.startswith('#') and '=' in line:
                key, value = line.split('=', 1)
                value = value.strip().strip('"').strip("'")
                os.environ.setdefault(key.strip(), value)

# Configuration
ODK_BASE_URL = os.getenv('ODK_CENTRAL_URL', 'https://data.dayawarga.com')
ODK_PROJECT_ID = os.getenv('ODK_PROJECT_ID', '3')
ODK_EMAIL = os.getenv('ODK_EMAIL', '')
ODK_PASSWORD = os.getenv('ODK_PASSWORD', '')

# All forms to process
ALL_FORMS = [
    'form_posko_v1',
    'form_feed_v1',
    'form_faskes_v1',
]

# Submitters to skip from auto-approve (comma-separated in env, or default list)
# These submissions require manual review
SKIP_SUBMITTERS_ENV = os.getenv('ODK_SKIP_SUBMITTERS', '')
SKIP_SUBMITTERS = [
    'Laporan Warga - DW',  # Default: skip submissions from public reports
]
if SKIP_SUBMITTERS_ENV:
    SKIP_SUBMITTERS = [s.strip() for s in SKIP_SUBMITTERS_ENV.split(',') if s.strip()]


class ODKCentralClient:
    def __init__(self, base_url: str, email: str, password: str):
        self.base_url = base_url.rstrip('/')
        self.email = email
        self.password = password
        self.session = requests.Session()
        self.token = None

    def authenticate(self) -> bool:
        url = f"{self.base_url}/v1/sessions"
        response = self.session.post(url, json={
            'email': self.email,
            'password': self.password
        })

        if response.status_code == 200:
            data = response.json()
            self.token = data.get('token')
            self.session.headers['Authorization'] = f'Bearer {self.token}'
            return True
        else:
            print(f"Authentication failed: {response.status_code} - {response.text}")
            return False

    def get_submissions(self, project_id: str, form_id: str) -> list:
        url = f"{self.base_url}/v1/projects/{project_id}/forms/{form_id}/submissions"
        response = self.session.get(url)
        if response.status_code == 200:
            return response.json()
        return []

    def set_review_state(self, project_id: str, form_id: str, instance_id: str, state: str) -> dict:
        """
        Set review state for a submission
        Valid states: null, hasIssues, edited, approved, rejected
        """
        url = f"{self.base_url}/v1/projects/{project_id}/forms/{form_id}/submissions/{instance_id}"

        # PATCH request to update review state
        response = self.session.patch(url, json={
            'reviewState': state
        })

        if response.status_code == 200:
            return {'success': True, 'data': response.json()}
        else:
            return {'success': False, 'error': response.text, 'status': response.status_code}


def should_skip_submitter(submission):
    """Check if submission should be skipped based on submitter name"""
    submitter = submission.get('submitter', {})
    display_name = submitter.get('displayName', '')
    return display_name in SKIP_SUBMITTERS


def process_form(client, project_id, form_id, dry_run=False, limit=0, include_edited=False):
    """Process a single form and approve pending submissions"""
    print(f"\n{'─' * 50}")
    print(f"Form: {form_id}")
    print('─' * 50)

    # Get submissions
    submissions = client.get_submissions(project_id, form_id)

    if not submissions:
        print(f"  No submissions found (or form doesn't exist)")
        return 0, 0, 0

    # Filter pending (null review state) and optionally edited
    pending = [s for s in submissions if s.get('reviewState') is None]
    edited = [s for s in submissions if s.get('reviewState') == 'edited']

    print(f"  Total: {len(submissions)}, Pending: {len(pending)}, Edited: {len(edited)}")

    # Combine based on flags
    to_approve = pending.copy()
    if include_edited:
        to_approve.extend(edited)

    if len(to_approve) == 0:
        print(f"  ✓ No submissions to approve")
        return 0, 0, 0

    # Filter out submissions from skip list
    skipped_submitters = [s for s in to_approve if should_skip_submitter(s)]
    to_approve = [s for s in to_approve if not should_skip_submitter(s)]

    if skipped_submitters:
        print(f"  ⊘ Skipped {len(skipped_submitters)} (submitter in skip list)")

    if len(to_approve) == 0:
        print(f"  ✓ No submissions to approve after filtering")
        return 0, 0, len(skipped_submitters)

    # Determine how many to process
    to_process = to_approve if limit == 0 else to_approve[:limit]

    if dry_run:
        print(f"  [DRY RUN] Would approve {len(to_process)} submissions")
        return len(to_process), 0, len(skipped_submitters)

    # Approve submissions
    success_count = 0
    error_count = 0

    for sub in to_process:
        instance_id = sub.get('instanceId')
        result = client.set_review_state(project_id, form_id, instance_id, 'approved')

        if result.get('success'):
            success_count += 1
        else:
            error_count += 1
            print(f"  ✗ {instance_id[:36]}... - {result.get('error', 'Unknown')[:50]}")

    if success_count > 0:
        print(f"  ✓ Approved {success_count} submissions")
    if error_count > 0:
        print(f"  ✗ Failed {error_count} submissions")

    return success_count, error_count, len(skipped_submitters)


def main():
    parser = argparse.ArgumentParser(description='Approve pending submissions in ODK Central')
    parser.add_argument('--dry-run', action='store_true', help='Show what would be approved without doing it')
    parser.add_argument('--limit', type=int, default=0, help='Limit number of approvals per form (0 = all)')
    parser.add_argument('--form-id', type=str, default=None, help='Process specific form only (default: all forms)')
    parser.add_argument('--project-id', type=str, default=ODK_PROJECT_ID, help='Project ID')
    parser.add_argument('--include-edited', action='store_true', help='Also approve submissions with edited state')
    args = parser.parse_args()

    print("=" * 60)
    print("ODK CENTRAL - APPROVE PENDING SUBMISSIONS")
    print("=" * 60)

    # Check credentials
    if not ODK_EMAIL or not ODK_PASSWORD:
        print("\nERROR: ODK credentials not set!")
        print("Set environment variables:")
        print("  export ODK_EMAIL='your-email'")
        print("  export ODK_PASSWORD='your-password'")
        sys.exit(1)

    # Connect
    print(f"\nConnecting to: {ODK_BASE_URL}")
    print(f"Project: {args.project_id}")

    client = ODKCentralClient(ODK_BASE_URL, ODK_EMAIL, ODK_PASSWORD)

    if not client.authenticate():
        print("Failed to authenticate!")
        sys.exit(1)

    print("Authenticated successfully!")

    # Show skip list if any
    if SKIP_SUBMITTERS:
        print(f"Skip list: {', '.join(SKIP_SUBMITTERS)}")

    # Determine which forms to process
    forms_to_process = [args.form_id] if args.form_id else ALL_FORMS

    # Process each form
    total_success = 0
    total_errors = 0
    total_skipped = 0

    for form_id in forms_to_process:
        success, errors, skipped = process_form(
            client,
            args.project_id,
            form_id,
            dry_run=args.dry_run,
            limit=args.limit,
            include_edited=args.include_edited
        )
        total_success += success
        total_errors += errors
        total_skipped += skipped

    # Summary
    print("\n" + "=" * 60)
    print(f"TOTAL: {total_success} approved, {total_skipped} skipped, {total_errors} errors")
    print("=" * 60)


if __name__ == '__main__':
    main()
