# Network Setup Guide (China)

## Problem: Cannot Download Go Modules

If you see errors like:
```
dial tcp 142.250.73.113:443: connectex: A connection attempt failed...
```

This means the default Go proxy (proxy.golang.org) is blocked.

## Solution: Use China Mirror

### Option 1: Quick Setup (Recommended)

**Windows:**
```bash
cd server
setup.bat
```

**Linux/Mac:**
```bash
cd server
chmod +x setup.sh
./setup.sh
```

### Option 2: Manual Setup

**Set Go proxy permanently:**

Windows (PowerShell):
```powershell
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=sum.golang.google.cn
```

Windows (CMD):
```cmd
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=sum.golang.google.cn
```

Linux/Mac:
```bash
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=sum.golang.google.cn
```

**Then download modules:**
```bash
cd server
go mod download
```

### Option 3: Temporary (One-time)

**Windows (PowerShell):**
```powershell
$env:GOPROXY="https://goproxy.cn,direct"
go mod download
```

**Windows (CMD):**
```cmd
set GOPROXY=https://goproxy.cn,direct
go mod download
```

**Linux/Mac:**
```bash
export GOPROXY=https://goproxy.cn,direct
go mod download
```

## Alternative Proxies

If `goproxy.cn` doesn't work, try these:

### Aliyun Mirror
```bash
go env -w GOPROXY=https://mirrors.aliyun.com/goproxy/,direct
```

### Tencent Mirror
```bash
go env -w GOPROXY=https://mirrors.tencent.com/go/,direct
```

### 7niu Mirror
```bash
go env -w GOPROXY=https://goproxy.io,direct
```

## Verify Setup

Check your current proxy:
```bash
go env GOPROXY
```

Should show:
```
https://goproxy.cn,direct
```

## Test Download

```bash
cd server
go mod download
```

Should complete without errors.

## Start Server

After successful setup:
```bash
go run main.go
```

## Troubleshooting

### Still can't download?

1. **Check internet connection**
   ```bash
   ping goproxy.cn
   ```

2. **Try different proxy**
   ```bash
   go env -w GOPROXY=https://mirrors.aliyun.com/goproxy/,direct
   go mod download
   ```

3. **Clear module cache**
   ```bash
   go clean -modcache
   go mod download
   ```

4. **Use VPN** (if available)
   - Connect to VPN
   - Use default proxy:
   ```bash
   go env -w GOPROXY=https://proxy.golang.org,direct
   ```

### Verify modules downloaded

```bash
go list -m all
```

Should show all dependencies.

## Permanent Configuration

Add to your shell profile (~/.bashrc, ~/.zshrc, or PowerShell profile):

**Linux/Mac:**
```bash
export GOPROXY=https://goproxy.cn,direct
export GOSUMDB=sum.golang.google.cn
```

**Windows PowerShell Profile:**
```powershell
$env:GOPROXY="https://goproxy.cn,direct"
$env:GOSUMDB="sum.golang.google.cn"
```

## Build Without Network

Once modules are downloaded, you can build offline:
```bash
go build -o stock-server main.go
```

The binary will work without internet connection.

## Common Errors

### Error: "unknown revision"
```bash
go clean -modcache
go mod download
```

### Error: "checksum mismatch"
```bash
go env -w GOSUMDB=off
go mod download
```

### Error: "timeout"
- Try different proxy
- Check firewall settings
- Use VPN

## Success Indicators

✅ `go mod download` completes without errors  
✅ `go run main.go` starts server  
✅ No network errors in console  

## Quick Reference

```bash
# Set proxy
go env -w GOPROXY=https://goproxy.cn,direct

# Download modules
go mod download

# Run server
go run main.go

# Build binary
go build -o stock-server main.go
```

## Need More Help?

1. Check Go version: `go version`
2. Check proxy setting: `go env GOPROXY`
3. Test network: `ping goproxy.cn`
4. Clear cache: `go clean -modcache`

---

**After setup, you should be able to run the server without issues!** 🚀
