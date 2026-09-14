$SERVICE     = "garage"
$ZONE        = "dc1"
$CAPACITY    = "10G"
$BUCKET_NAME = "iot-images"
$KEY_NAME    = "app-key"

Write-Host "==> Starting $SERVICE container..." -ForegroundColor Cyan
docker compose up -d $SERVICE

Write-Host "==> Waiting for Garage daemon to become ready..." -ForegroundColor Cyan
while ($true) {
    cmd /c "docker compose exec -T $SERVICE /garage status >nul 2>&1"
    if ($LASTEXITCODE -eq 0) { break }
    Start-Sleep -Seconds 1
}

Write-Host "==> Fetching Node ID..." -ForegroundColor Cyan
$rawNodeOutput = docker compose exec -T $SERVICE /garage node id 2>$null | Out-String

# Extract exactly the 64-character hex ID before the @ symbol
if ($rawNodeOutput -match '([0-9a-fA-F]{64})(@|\s|$)') {
    $NODE_ID = $matches[1]
    Write-Host "Node ID: $NODE_ID" -ForegroundColor Green
} else {
    Write-Host "Failed to parse Node ID. Raw output:" -ForegroundColor Red
    Write-Host $rawNodeOutput
    exit 1
}

Write-Host "==> Configuring cluster layout..." -ForegroundColor Cyan
docker compose exec -T $SERVICE /garage layout assign -z $ZONE -c $CAPACITY $NODE_ID

Write-Host "==> Applying layout..." -ForegroundColor Cyan
$layoutShow = docker compose exec -T $SERVICE /garage layout show 2>$null | Out-String
$match = [regex]::Match($layoutShow, 'Current layout version:\s+(\d+)')
$currentVer = if ($match.Success) { [int]$match.Groups[1].Value } else { 0 }
$targetVer  = $currentVer + 1

docker compose exec -T $SERVICE /garage layout apply --version $targetVer

Write-Host "==> Creating bucket: $BUCKET_NAME..." -ForegroundColor Cyan
docker compose exec -T $SERVICE /garage bucket create $BUCKET_NAME 2>$null

Write-Host "==> Handling access key: $KEY_NAME..." -ForegroundColor Cyan
# Check if key already exists
cmd /c "docker compose exec -T $SERVICE /garage key info $KEY_NAME >nul 2>&1"

if ($LASTEXITCODE -ne 0) {
    # Key doesn't exist; create it and capture output
    $KEY_OUTPUT = docker compose exec -T $SERVICE /garage key create $KEY_NAME
    
    docker compose exec -T $SERVICE /garage bucket allow --read --write $BUCKET_NAME --key $KEY_NAME 2>$null | Out-Null

    Write-Host "`n==========================================" -ForegroundColor Green
    Write-Host " COPY YOUR CREDENTIALS (SECRET SHOWN ONCE)" -ForegroundColor Yellow
    Write-Host "==========================================" -ForegroundColor Green
    $KEY_OUTPUT | ForEach-Object { Write-Host $_ }
    Write-Host "==========================================`n" -ForegroundColor Green
} else {
    Write-Host "==> Key '$KEY_NAME' already exists." -ForegroundColor Yellow
    Write-Host "Note: Garage cannot re-print a secret key once created." -ForegroundColor Yellow
    docker compose exec -T $SERVICE /garage bucket allow --read --write $BUCKET_NAME --key $KEY_NAME 2>$null | Out-Null
}