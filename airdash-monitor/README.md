# AirDash Monitor - Air Quality Alert System

Monitors your AirGradient sensor and sends push notifications to your iPhone when PM2.5 levels exceed a threshold.

## Setup Instructions

### 1. Get Pushover Credentials

1. Purchase the Pushover app from the iOS App Store ($5 one-time)
2. Create a free account at https://pushover.net
3. After logging in, you'll see your **User Key** on the dashboard
4. Create a new application at https://pushover.net/apps/build
   - Name it "AirDash" or whatever you prefer
   - Copy the **API Token/Key** that's generated

### 2. Configure the Monitor

Edit `~/.airdash/monitor-config.yaml` and add your credentials:

```yaml
pushoverUserKey: YOUR_USER_KEY_HERE      # From Pushover dashboard
pushoverApiToken: YOUR_API_TOKEN_HERE    # From app you created
pm25Threshold: 10                        # Alert when PM2.5 exceeds this
checkInterval: 5                         # Check every 5 minutes
```

### 3. Test the Monitor

Run it manually to test:
```bash
cd /Users/work/Repos/airdash/airdash-monitor
./airdash-monitor
```

You should see output like:
```
AirDash Monitor starting...
Monitoring PM2.5 levels. Threshold: 10.0 μg/m³
Checking every 5 minutes
Current PM2.5: 0.0 μg/m³, CO2: 689 ppm
```

To test notifications, temporarily set `pm25Threshold: 0` in the config, then run it again. You should get an alert notification immediately.

### 4. Enable Auto-Start

Load the Launch Agent to run automatically:
```bash
launchctl load /Users/work/Library/LaunchAgents/com.airdash.monitor.plist
```

Verify it's running:
```bash
launchctl list | grep airdash.monitor
```

### 5. View Logs

Check the logs to see what's happening:
```bash
tail -f /tmp/airdash-monitor.log
```

## How It Works

- Checks AirGradient API every 5 minutes (configurable)
- When PM2.5 exceeds threshold: sends high-priority alert notification
- When PM2.5 returns to normal: sends "all clear" notification
- Prevents spam by only alerting once when crossing threshold

## Notifications

**Alert notification** (high priority):
```
Air Quality Alert
PM2.5 is elevated at 15.2 μg/m³ (threshold: 10.0)

Current readings:
• PM2.5: 15.2 μg/m³
• CO2: 850 ppm
```

**All clear notification** (normal priority):
```
Air Quality Normal
PM2.5 has returned to safe levels at 8.5 μg/m³

Current readings:
• PM2.5: 8.5 μg/m³
• CO2: 720 ppm
```

## PM2.5 Reference Levels

- **0-12**: Good air quality
- **12-35**: Moderate
- **35-55**: Unhealthy for sensitive groups
- **55+**: Unhealthy

Default threshold of 10 gives you a heads up before air quality becomes "moderate".

## Troubleshooting

**No notifications?**
- Check config file has correct Pushover credentials
- Test Pushover by logging into https://pushover.net and using "Send a Notification"
- Check logs: `cat /tmp/airdash-monitor.error.log`

**Monitor not running?**
```bash
launchctl unload /Users/work/Library/LaunchAgents/com.airdash.monitor.plist
launchctl load /Users/work/Library/LaunchAgents/com.airdash.monitor.plist
```

## Disable Auto-Start

```bash
launchctl unload /Users/work/Library/LaunchAgents/com.airdash.monitor.plist
```
