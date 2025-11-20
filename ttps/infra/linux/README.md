# Async Worker SSH Production Access Test

This TTP validates whether async workers can establish unauthorized SSH access to other compute resources, helping identify overprivileged service accounts and potential lateral movement paths.

## Test Scenarios

This TTP tests lateral movement capabilities by attempting SSH connections from `async_tier` workers to:

1. **Async worker to itself (localhost)** - Tests if async worker can SSH to itself
2. **Async worker to Sandcastle workers** - Tests lateral movement to sandcastle compute surface
3. **Async worker to other Async workers** - Tests lateral movement between async workers
4. **Async worker to Chronos workers** - Tests lateral movement to chronos job scheduler
5. **Async worker to DevVMs** - Tests lateral movement to development environments
6. **Async worker to production hosts** - Tests unauthorized access to production infrastructure

## How It Works

The TTP uses default SSH configuration with BatchMode enabled (non-interactive):
- Attempts connection with any available SSH keys in `~/.ssh/`
- Uses standard SSH without passwords or explicit credentials
- Tests if SSH daemon is running and accessible
- Verifies if async worker service account has SSH permissions

## Arguments
- **target_host**: The target host to SSH into (default: `localhost`)
  - Examples: `localhost`, `sandcastle-worker.fbinfra.net`, `chronos-worker.facebook.com`, `prod-host.facebook.com`
- **username**: The username to use for SSH connection (default: `twsvcscm`)
- **test_file_path**: Path for test marker file on target (default: `/tmp/async_ssh_test_marker_$$_$RANDOM.txt`)
- **sleep_time**: Time to sleep before cleanup in seconds (default: `5`)
- **connection_timeout**: SSH connection timeout in seconds (default: `10`)

## Steps
1. **Test SSH Access**: Attempt SSH connection and create test marker file
2. **Sleep Before Cleanup**: Wait for configured time before cleanup
3. **Generate Test Report**: Display summary report of test results

## Expected Results
- ** ACCESS BLOCKED** (Secure): Async worker cannot SSH to target - correct security posture
- ** ACCESS GRANTED** (Risk): Async worker can SSH to target - potential security misconfiguration

### Quick Test (Localhost)
```bash
bas run --group-id="async_tier" ttps/infra/linux/async_worker_ssh_to_prod.yaml
```

### Test Against Specific Host
```bash
bas run --group-id="async_tier" \
  ttps/infra/linux/async_worker_ssh_to_prod.yaml \
  -a target_host=sandcastle-worker.fbinfra.net
```

This will test async_tier workers attempting SSH to:
- localhost (self)
- Sandcastle workers
- Other async workers
- Chronos workers
- DevVM instances

## MITRE ATT&CK Mapping
- **Tactics**:
  - TA0001: Initial Access
  - TA0008: Lateral Movement
- **Techniques**:
  - T1078: Valid Accounts
  - T1021.004: SSH
- **Sub-Techniques**:
  - T1021.004: SSH

## Reference
- Related Work: SSH lateral movement detection and prevention
- Security Best Practice: Principle of least privilege for service accounts
