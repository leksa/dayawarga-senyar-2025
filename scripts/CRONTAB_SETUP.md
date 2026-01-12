# Crontab Setup for Auto-Approve & Sync

## Quick Setup on Server

1. **Copy environment file:**
```bash
sudo mkdir -p /etc/dayawarga
sudo cp scripts/cron.env.example /etc/dayawarga/cron.env
sudo chmod 600 /etc/dayawarga/cron.env
sudo nano /etc/dayawarga/cron.env  # Edit with real credentials
```

2. **Install Python dependencies:**
```bash
pip3 install requests
```

3. **Add to crontab:**
```bash
sudo crontab -e
```

Add this line (runs every 5 minutes):
```
*/5 * * * * . /etc/dayawarga/cron.env && /home/deploy/dayawarga-senyar-2025/scripts/cron-autoapprove-sync.sh >> /var/log/dayawarga-sync.log 2>&1
```

4. **Create log file with proper permissions:**
```bash
sudo touch /var/log/dayawarga-sync.log
sudo chmod 644 /var/log/dayawarga-sync.log
```

## Alternative: Using Systemd Timer

1. **Create service file `/etc/systemd/system/dayawarga-sync.service`:**
```ini
[Unit]
Description=Dayawarga Auto-Approve and Sync
After=network.target

[Service]
Type=oneshot
EnvironmentFile=/etc/dayawarga/cron.env
ExecStart=/home/deploy/dayawarga-senyar-2025/scripts/cron-autoapprove-sync.sh
User=deploy
WorkingDirectory=/home/deploy/dayawarga-senyar-2025/scripts
StandardOutput=append:/var/log/dayawarga-sync.log
StandardError=append:/var/log/dayawarga-sync.log
```

2. **Create timer file `/etc/systemd/system/dayawarga-sync.timer`:**
```ini
[Unit]
Description=Run Dayawarga sync every 5 minutes

[Timer]
OnBootSec=1min
OnUnitActiveSec=5min
AccuracySec=1min

[Install]
WantedBy=timers.target
```

3. **Enable and start:**
```bash
sudo systemctl daemon-reload
sudo systemctl enable dayawarga-sync.timer
sudo systemctl start dayawarga-sync.timer
```

4. **Check status:**
```bash
sudo systemctl status dayawarga-sync.timer
sudo systemctl list-timers
```

## Manual Test

```bash
# Test the script manually
source /etc/dayawarga/cron.env
/home/deploy/dayawarga-senyar-2025/scripts/cron-autoapprove-sync.sh
```

## Logs

View sync logs:
```bash
tail -f /var/log/dayawarga-sync.log
```
