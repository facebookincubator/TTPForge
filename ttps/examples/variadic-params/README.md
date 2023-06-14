# delta.yaml

Test a Meta account for the Delta 2FA mechanism.

## Arguments

- target: target URL (default: <https://auth.meta.com>)
- ignoreCertErrors: ignore certificate errors (default: false)
- headless: run the TTP in headless mode (default: true)
- keeperRecord: record ID of credential in Keeper (required if no user or password provided)
- user: email address for the user (required if no keeperRecord is present)
- password: password (required if no keeperRecord is present)

## Examples

Test an account for the delta 2FA gating mechanism using credentials from keeper:

```bash
KEEPER_RECORD=WPlgP-G2xlR2B_MV-6XRaA # Non 2FA that had no delta at the time of creation - 6/8/2023
KEEPER_RECORD=r0H-B6_g3PdEfVI4AAKMOw # Account with delta activated

./ttpforge \
    -c config.yaml \
    -l ttpforge.log \
    run ttps/examples/variadic-params/delta.yaml \
    --arg headless=false \
    --arg keeperRecord=$KEEPER_RECORD
```
