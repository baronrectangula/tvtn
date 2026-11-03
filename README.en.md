# TVTN: Tìm việc, tìm người

[Tiếng Việt](README.md) | English

## Contents

- [Setup and running](#setup-and-running)
  - [Docker Compose](#docker-compose)
  - [Manually](#manually)
- [Configuration](#configuration)
- [Verification and tests](#verification-and-tests)
- [Git hooks](#git-hooks)
  - [Commit messages](#commit-messages)
  - [Configuration change confirmation](#configuration-change-confirmation)
- [License](#license)

## Setup and running

Run the commands below from the project root. Local settings and credentials
are provided. To change them, see [Configuration](#configuration).

### Docker Compose

This is the preferred way to run the application and its infrastructure services.
Install Docker and Docker Compose; Make is optional:

```sh
docker compose --profile rest-server up -d --build
# Or:
make compose profile=rest-server
```

The REST server runs at `http://127.0.0.1:8080`. To view logs or stop the stack:

```sh
docker compose logs -f rest-server
docker compose --profile "*" down   # Keep persisted data
```

### Manually

Requires Go 1.27+. Docker Compose provides the infrastructure services.
Start them and wait until they are ready:

```sh
docker compose up -d --wait
```

Then load the credentials in Bash/Zsh and start the server:

```sh
set -a
source .env.sample
set +a
go run ./cmd/rest-server -config config.toml -addr 127.0.0.1:8080
```

With Make installed, you can replace `go run` with `make run`. To select a
program or use a different port:

```sh
make run cmd=rest-server args="-addr 127.0.0.1:9090"
```

## Configuration

Both methods use the same [config.toml](config.toml).

| Setting | Where to change it |
| --- | --- |
| PostgreSQL connection | `host`, `port`, `database`, and `ssl_mode` in the TOML `[postgres]` section |
| PostgreSQL host by environment | `POSTGRES_HOST` overrides TOML; Compose sets it to `postgres`, while manual runs default to `127.0.0.1` |
| Username and password | `POSTGRES_USER` and `POSTGRES_PASSWORD` environment variables; the application does not read these from TOML |
| HTTP address | The `-addr` flag for manual runs; the `rest-server` service's `ports` mapping for Compose |

To use your own credentials, copy the sample file and edit the values in `.env`:

```sh
cp .env.sample .env
```

Compose reads `.env` automatically; exported shell variables take precedence.
For manual runs, replace `source .env.sample` with `source .env` in the steps above.
Git ignores `.env`; keep `.env.sample` as the local example.

Restart the application after changing settings. For Compose, rerun the stack
startup command. If you change the database name, update both `database` in TOML
and `POSTGRES_DB` in Compose. If PostgreSQL already contains data, update the
database or credentials in PostgreSQL to match; changing environment variables
does not update an initialized database.

## Verification and tests

The server exposes `/livez` to check HTTP liveness and `/readyz` to check the
PostgreSQL connection.

```sh
curl --fail http://127.0.0.1:8080/livez
curl --fail http://127.0.0.1:8080/readyz
```

Run tests with Go 1.27+:

```sh
go test ./...
# Or:
make test
```

## Git hooks

Run `make hooks` once after cloning.

### Commit messages

The hook and CI check Conventional Commit style: `type(scope): description`, with lowercase types
`feat`, `chore`, `fix`, `test`, `doc`, `refactor`, or `perf`. The scope is optional. Separate the body
and footers from the subject with a blank line. A `!` requires a footer describing
the incompatible change; the footer can also mark a breaking change without `!`:

```text
feat(api)!: change the response format

BREAKING CHANGE: The user_id field is now named id.
```

`BREAKING-CHANGE` is also accepted in place of `BREAKING CHANGE`.

### Configuration change confirmation

When committing staged changes to `config.toml`
or `.env.sample`, review them and add either footer after a blank line:

```text
Config-Change: I confirm these settings and credentials are for local development only.
```

The footer must match exactly, including capitalization. This hook prompts a review
of local settings; it does not detect secrets and can be bypassed with `--no-verify`.

CI runs lint, tests, and commit footer checks on pushes and same-repository pull requests.
All jobs, including notifications, skip fork pull requests and forked copies of the repository.
The first push of a branch checks its entire history.

When configuration changes lack a valid confirmation footer, a bot comments on the
PR (or commit if no PR is open), mentioning the personal repository owner and admins
with links to each offending commit and the check. CI execution errors do not send this notification. The notification workflow must be on the default branch; delivery
to each recipient depends on their GitHub notification settings.

## License

TVTN is licensed under [AGPL-3.0-only](LICENSE). Dependencies in `vendor/` retain their own licenses.
