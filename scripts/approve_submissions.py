#!/usr/bin/env python3
"""
Script to approve pending submissions in ODK Central
Usage: python approve_submissions.py [--dry-run] [--limit N]
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
ODK_FORM_ID = os.getenv('ODK_FORM_ID', 'form_posko_v1')
ODK_EMAIL = os.getenv('ODK_EMAIL', '')
ODK_PASSWORD = os.getenv('ODK_PASSWORD', '')


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


def main():
    parser = argparse.ArgumentParser(description='Approve pending submissions in ODK Central')
    parser.add_argument('--dry-run', action='store_true', help='Show what would be approved without doing it')
    parser.add_argument('--limit', type=int, default=0, help='Limit number of approvals (0 = all)')
    parser.add_argument('--form-id', type=str, default=ODK_FORM_ID, help='Form ID to process')
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
    print(f"Project: {args.project_id}, Form: {args.form_id}")

    client = ODKCentralClient(ODK_BASE_URL, ODK_EMAIL, ODK_PASSWORD)

    if not client.authenticate():
        print("Failed to authenticate!")
        sys.exit(1)

    print("Authenticated successfully!")

    # Get submissions
    print("\nFetching submissions...")
    submissions = client.get_submissions(args.project_id, args.form_id)
    print(f"Total submissions: {len(submissions)}")

    # Filter pending (null review state) and optionally edited
    pending = [s for s in submissions if s.get('reviewState') is None]
    edited = [s for s in submissions if s.get('reviewState') == 'edited']
    print(f"Pending (null): {len(pending)}")
    print(f"Edited: {len(edited)}")

    # Combine based on flags
    to_approve = pending.copy()
    if args.include_edited:
        to_approve.extend(edited)
        print(f"Including edited submissions: {len(edited)}")

    if len(to_approve) == 0:
        print("\nNo submissions to approve!")
        return

    # Determine how many to process
    to_process = to_approve if args.limit == 0 else to_approve[:args.limit]
    print(f"\nWill process: {len(to_process)} submissions")

    if args.dry_run:
        print("\n[DRY RUN] Would approve:")
        for i, sub in enumerate(to_process[:10]):
            print(f"  {i+1}. {sub.get('instanceId')}")
        if len(to_process) > 10:
            print(f"  ... and {len(to_process) - 10} more")
        print("\nTo approve for real, remove --dry-run flag")
        return

    # Approve submissions
    print("\nApproving submissions...")
    success_count = 0
    error_count = 0

    for i, sub in enumerate(to_process):
        instance_id = sub.get('instanceId')
        result = client.set_review_state(args.project_id, args.form_id, instance_id, 'approved')

        if result.get('success'):
            success_count += 1
            print(f"✓ [{i+1}/{len(to_process)}] {instance_id[:36]}...")
        else:
            error_count += 1
            print(f"✗ [{i+1}/{len(to_process)}] {instance_id[:36]}... - {result.get('error', 'Unknown error')[:50]}")

    # Summary
    print("\n" + "=" * 60)
    print(f"SUMMARY: {success_count} approved, {error_count} errors")
    print("=" * 60)


if __name__ == '__main__':
    main()
