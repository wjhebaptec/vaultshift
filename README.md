# vaultshift

> A CLI tool for rotating and syncing secrets across multiple cloud secret managers (AWS, GCP, Vault).

---

## Installation

```bash
go install github.com/yourusername/vaultshift@latest
```

Or download a pre-built binary from the [releases page](https://github.com/yourusername/vaultshift/releases).

---

## Usage

### Sync secrets from AWS Secrets Manager to HashiCorp Vault

```bash
vaultshift sync --from aws --to vault --secret my/app/secret
```

### Rotate a secret across all configured providers

```bash
vaultshift rotate --secret my/app/db-password --providers aws,gcp,vault
```

### Configuration

Create a `vaultshift.yaml` file in your working directory:

```yaml
providers:
  aws:
    region: us-east-1
  gcp:
    project: my-gcp-project
  vault:
    address: https://vault.example.com
    token: $VAULT_TOKEN
```

Then run:

```bash
vaultshift sync --config vaultshift.yaml
```

### Available Commands

| Command   | Description                                      |
|-----------|--------------------------------------------------|
| `sync`    | Sync secrets between two providers               |
| `rotate`  | Rotate a secret and propagate to all providers   |
| `list`    | List secrets available in a provider             |
| `diff`    | Show differences in secrets between providers    |

---

## Requirements

- Go 1.21+
- Credentials configured for your target cloud provider(s)

---

## Contributing

Pull requests are welcome. Please open an issue first to discuss any major changes.

---

## License

[MIT](LICENSE)