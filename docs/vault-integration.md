# HashiCorp Vault Integration

Lobstertank supports HashiCorp Vault as a secrets provider for secure credential storage.

## Overview

When using Vault, gateway credentials (tokens, API keys) are stored in Vault KV v2 instead of the built-in encrypted store. This enables:

- Centralized secret management across multiple Lobstertank instances
- Dynamic secret rotation without redeploying
- Audit logging of secret access
- Integration with existing Vault infrastructure

## Configuration

### 1. Enable Vault in Helm Values

```yaml
config:
  secretsProvider: vault

vault:
  enabled: true
  addr: "https://vault.example.com"
  mountPath: "secret"
  token: "hvs.xxxxxx"  # Or use existingSecret
```

### 2. Use Existing Secret (Production)

For production, store the Vault token in a pre-existing secret:

```yaml
config:
  secretsProvider: vault

vault:
  enabled: true
  addr: "https://vault.example.com"
  mountPath: "secret"
  existingSecret: "vault-credentials"
  existingSecretKey: "token"
```

Create the secret separately:

```bash
kubectl create secret generic vault-credentials \
  --from-literal=token=hvs.xxxxxx \
  -n lobstertank
```

### 3. Vault KV Structure

Secrets are stored under the configured mount path with this structure:

```
secret/
  lobstertank/
    gateways/
      {gateway-id}/
        token
        api-key
        other-creds
```

Example:
```bash
# Store a gateway token
vault kv put secret/lobstertank/gateways/gw-123/token value="my-secret-token"

# Store API credentials
vault kv put secret/lobstertank/gateways/gw-123/openai-api-key value="sk-..."
```

## Migration from Built-in Provider

To migrate existing secrets from the built-in provider to Vault:

1. Export existing secrets:
   ```bash
   # From Lobstertank with builtin provider
   curl -H "Authorization: Bearer $TOKEN" \
     http://lobstertank-backend:8080/api/v1/gateways | jq .
   ```

2. Import to Vault:
   ```bash
   for gw in $(curl -s ... | jq -r '.[].id'); do
     token=$(curl -s ... | jq -r '.[].auth.token')
     vault kv put secret/lobstertank/gateways/$gw/token value="$token"
   done
   ```

3. Update Helm values to use Vault provider
4. Redeploy Lobstertank

## Security Considerations

- **Token Rotation**: Use Vault's dynamic secrets or short-lived tokens
- **Network Policy**: Restrict egress from Lobstertank to Vault only
- **Audit Logging**: Enable Vault audit logs for compliance
- **mTLS**: Configure Vault with mTLS for additional security

## Troubleshooting

### "vault returned HTTP 403"
- Check Vault token has read/write access to the KV path
- Verify token hasn't expired

### "secret not found in vault"
- Ensure secret exists at the expected path
- Check `vault.mountPath` matches your KV mount

### Connection refused
- Verify `vault.addr` is reachable from the cluster
- Check network policies allow egress to Vault
