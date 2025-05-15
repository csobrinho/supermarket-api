# supermarket-api
Supermarket API service

This CLI can easily be called from a CronJob to clip all deals the user has in their account.

## Run
```sh
# Export your ENV variables or set the flags.
go run cmd/supermarket/main.go
```

```log
INFO : 2025/05/15 13:43:55.139077 main.go:60: main: registering safeway...
INFO : 2025/05/15 13:43:55.139321 logger.go:344: Info verbosity set to 0
INFO : 2025/05/15 13:43:55.139366 factory.go:44: supermarket: registered "safeway" provider
INFO : 2025/05/15 13:43:55.139402 main.go:83: main: getting an access token...
INFO : 2025/05/15 13:43:55.520381 main.go:90: main: getting all promotions...
INFO : 2025/05/15 13:43:55.914610 promotion.go:67: promotion: found 440 clip deals
INFO : 2025/05/15 13:43:55.915179 promotion.go:81: promotion:   - C: 440
INFO : 2025/05/15 13:43:55.915190 promotion.go:81: promotion:   - U: 0
INFO : 2025/05/15 13:43:55.915203 main.go:113: main: clip stats:
    - already: 440
    - newly:   0
    - deleted: 0
    - ignored: 0
    - errors:  0
INFO : 2025/05/15 13:43:55.915206 main.go:172: main: done
```