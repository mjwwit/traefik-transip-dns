# Traefik TransIP DNS Provider
A TransIP Let's Encrypt DNS challenge provider for Traefik.
Pretty much all credits for this should go to [mpdroog](https://github.com/mpdroog) and [alexflint](https://github.com/alexflint) for creating such beautiful and easy to use libraries.

## Configuration
This project relies on the traefik `exec` DNS provider. For it to work you'll need to run the `alpine` version of Traefik.
Also, the following environment variables need to be configured:

```
EXEC_PATH={path to the binary created from this project}
TRANSIP_USERNAME={your transip username}
TRANSIP_PRIVATE_KEY_PATH={path to the private key file}
```

And, of course, your `traefik.toml` file needs:
```toml
[acme.dnsChallenge]
provider = "exec"
```

## Notes
This project has only been tested on non-wildcard subdomains. If anyone has the option and the balls to try this, let me know how it goes.

## Building
This project uses [dep](https://golang.github.io/dep/).

Build a binary using `build.sh`.