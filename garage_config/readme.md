# Garage S3 Docker Setup

### Setup
1) Go to garage.toml 
2) Gen new key hex-32 with 64 char at rpc_secret for init garage


### Script Run
Run Script init (window .ps1 / linux .sh) for auto create and log out ACCESS KEY INFORMATION


### Manual Setup Garage Docker
```bash
docker compose up -d garage

docker compose exec garage /garage node id

# Replace <NODE_ID> with the actual ID from step 2
docker compose exec garage /garage layout assign -z dc1 -c 10G <NODE_ID>

docker compose exec garage /garage layout apply --version 1

# iot-images by env set bucket
docker compose exec garage /garage bucket create iot-images

# make this bucket user as app-key
docker compose exec garage /garage key create app-key

docker compose exec garage /garage bucket allow \
  --read \
  --write \
  iot-images \
  --key iot-app-key
```