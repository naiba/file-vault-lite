# File Vault Lite

For guys who have exceeded their GitHub storage quota.

## Deploy

```yaml
version: '3.4'
services:
  file_vault_lite:
    image: ghcr.io/naiba/file-vault-lite:latest
    restart: always
    environment:
      - "FV_USERNAME=username"
      - "FV_PASSWORD=password"
    ports:
      - "8080:8080"
    volumes:
        - ./uploads:/file-vault-lite/uploads
```

## Upload

```shell
curl -u username:password -F "file=@/path/to/file" -F "filename=myfile.txt" http://localhost:8080/upload
```

## Download

```shell
curl -u username:password http://localhost:8080/download?filename=myfile.txt -o myfile.txt
```
