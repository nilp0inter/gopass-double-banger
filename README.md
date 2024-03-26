# gopass-double-banger

Gopass integration for nested secrets

## Installation

### From Source

```bash
go install github.com/gopasspw/gopass-double-banger@latest
```

## bootstrap-yk-oath

This script loads a Yubikey with OATH secrets from a gopass store subfolder
which secrets are stored with gopass-double-banger.

Requires the `ykman` command line tool to be installed.

```console
$ nix run .#bootstrap-yk-oath
Usage: bootstrap-yk-oath <YK_DEVICE> <TOTP_PATH>
```

Where `YK_DEVICE` is your Yubikey serial number and `TOTP_PATH` is the path to the gopass store subfolder
containing the secrets.

You can find the Yubikey serial number by running `ykman list`.
