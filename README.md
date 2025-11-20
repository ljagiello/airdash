# AirDash

Complete air quality monitoring system for macOS with AirGradient sensors.

![screenshot](screenshot.png)

## Features

- **Menu Bar Display**: Real-time PM2.5 and CO2 readings in your Mac menu bar
- **iPhone Alerts**: Push notifications when air quality degrades (PM2.5 or CO2 exceed thresholds)
- **Auto-Start**: Both services run automatically on login
- **Independent Alerts**: Separate notifications for PM2.5 and CO2 levels

## Architecture

AirDash consists of two independent services:

1. **Menu Bar App** (`airdash`): Displays real-time air quality in your Mac menu bar
   - Shows PM2.5 and CO2 levels
   - Updates every 60 seconds

2. **Monitor Service** (`airdash-monitor`): Background monitoring with push notifications
   - Checks air quality every 5 minutes
   - Sends iPhone alerts when thresholds are exceeded
   - Sends "all clear" notifications when levels return to normal

## Setup

### Prerequisites

- macOS (tested on Darwin 25.x)
- Go 1.21+ (for building from source)
- AirGradient sensor with API access
- Pushover account (for iPhone notifications)

### 1. Configure AirGradient

Create `~/.airdash/config.yaml`:

```yaml
token: <your-airgradient-api-token>
locationId: <your-location-id>
interval: 60        # Update interval in seconds
tempUnit: F         # F or C (not used in display)
```

Get your token and locationId from https://app.airgradient.com/

### 2. Build the Applications

```bash
cd /Users/work/Repos/airdash

# Build menu bar app
go build -o airdash .

# Build monitor
cd airdash-monitor
go build -o airdash-monitor .
```

### 3. Set Up Pushover (for iPhone Alerts)

1. Purchase Pushover app from iOS App Store ($5 one-time)
2. Create account at https://pushover.net
3. Get your **User Key** from the dashboard
4. Create a new application at https://pushover.net/apps/build
   - Name: "AirDash"
   - Copy the **API Token**

### 4. Configure Monitor

Create `~/.airdash/monitor-config.yaml`:

```yaml
pushoverUserKey: <your-pushover-user-key>
pushoverApiToken: <your-pushover-api-token>
pm25Threshold: 10      # Alert when PM2.5 > 10 μg/m³
co2Threshold: 750      # Alert when CO2 > 750 ppm
checkInterval: 5       # Check every 5 minutes
```

### 5. Test the Applications

Test menu bar app:
```bash
cd /Users/work/Repos/airdash
./airdash
```

Test monitor:
```bash
cd /Users/work/Repos/airdash/airdash-monitor
./airdash-monitor
```

You should see output like:
```
AirDash Monitor starting...
Monitoring PM2.5 levels. Threshold: 10.0 μg/m³
Monitoring CO2 levels. Threshold: 750 ppm
Checking every 5 minutes
Current PM2.5: 0.0 μg/m³, CO2: 711 ppm
```

### 6. Enable Auto-Start

The Launch Agents are already created at:
- `~/Library/LaunchAgents/com.airdash.plist`
- `~/Library/LaunchAgents/com.airdash.monitor.plist`

Load them to start on login:
```bash
launchctl load ~/Library/LaunchAgents/com.airdash.plist
launchctl load ~/Library/LaunchAgents/com.airdash.monitor.plist
```

Verify they're running:
```bash
launchctl list | grep airdash
```

## Air Quality Reference Levels

### PM2.5 (μg/m³)
- **0-12**: Good air quality
- **12-35**: Moderate
- **35-55**: Unhealthy for sensitive groups
- **55+**: Unhealthy

Default threshold of 10 alerts you before air quality becomes "moderate".

### CO2 (ppm)
- **400-600**: Outdoor/fresh air
- **600-800**: Good indoor air
- **800-1000**: Adequate ventilation
- **1000+**: Poor ventilation - open windows!

Default threshold of 750 alerts you before air quality becomes stuffy.

## Notifications

### PM2.5 Alert (high priority)
```
PM2.5 Alert
PM2.5 is elevated at 15.2 μg/m³ (threshold: 10.0)

Current readings:
• PM2.5: 15.2 μg/m³
• CO2: 720 ppm
```

### PM2.5 All Clear (normal priority)
```
PM2.5 Normal
PM2.5 has returned to safe levels at 8.5 μg/m³

Current readings:
• PM2.5: 8.5 μg/m³
• CO2: 720 ppm
```

### CO2 Alert (high priority)
```
CO2 Alert
CO2 is elevated at 850 ppm (threshold: 750)

Current readings:
• PM2.5: 5.0 μg/m³
• CO2: 850 ppm
```

### CO2 All Clear (normal priority)
```
CO2 Normal
CO2 has returned to safe levels at 720 ppm

Current readings:
• PM2.5: 5.0 μg/m³
• CO2: 720 ppm
```

## How Alerts Work

- Checks AirGradient API every 5 minutes (configurable)
- PM2.5 and CO2 are monitored independently
- Sends **high-priority alert** when threshold is exceeded
- Sends **all-clear notification** when levels return to normal
- Prevents spam by only alerting once when crossing thresholds
- Both alerts can trigger simultaneously

## Testing

Run the unit tests to verify alert logic:
```bash
cd /Users/work/Repos/airdash/airdash-monitor
go test -v
```

All tests should pass:
- ✓ PM2.5 alert triggers when exceeding threshold
- ✓ PM2.5 all-clear when returning to normal
- ✓ CO2 alert triggers when exceeding threshold
- ✓ CO2 all-clear when returning to normal
- ✓ Both alerts can trigger simultaneously
- ✓ No spam when already active
- ✓ No alerts when levels are normal

## Troubleshooting

### No notifications?
- Check config file has correct Pushover credentials
- Test Pushover by logging into https://pushover.net and using "Send a Notification"
- Check logs: `cat /tmp/airdash-monitor.error.log`
- Verify thresholds are set appropriately for your current air quality

### Monitor not running?
```bash
launchctl unload ~/Library/LaunchAgents/com.airdash.monitor.plist
launchctl load ~/Library/LaunchAgents/com.airdash.monitor.plist
```

### Want to test alerts?
Temporarily lower thresholds in `~/.airdash/monitor-config.yaml`:
```bash
# Set pm25Threshold: 0.1 and/or co2Threshold: 500
launchctl unload ~/Library/LaunchAgents/com.airdash.monitor.plist
launchctl load ~/Library/LaunchAgents/com.airdash.monitor.plist
# You should get alerts immediately
# Don't forget to set thresholds back to normal!
```

### View logs
```bash
# Menu bar app logs
tail -f /tmp/airdash.error.log

# Monitor logs
tail -f /tmp/airdash-monitor.error.log
```

## Disable Auto-Start

```bash
launchctl unload ~/Library/LaunchAgents/com.airdash.plist
launchctl unload ~/Library/LaunchAgents/com.airdash.monitor.plist
```

## API Reference

- AirGradient API - https://api.airgradient.com/public/docs/api/v1/
- Pushover API - https://pushover.net/api
