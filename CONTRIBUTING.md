## How to be a contributor to this project

### v2.0.0 Migration Notes for Contributors

Nancy v2.0.0 is a breaking change release. Key changes relevant to contributors:

- The vulnerability data backend has moved from **OSS Index** to **Sonatype Guide**. Use a `GUIDE_TOKEN` for authentication instead of `OSSI_USERNAME`/`OSSI_TOKEN`.
- The `nancy iq` command has been renamed to `nancy lifecycle`. The `nancy iq` alias is retained but deprecated.
- The `--iq-*` flags are deprecated aliases for the new `--lifecycle-*` flags.
- Gopkg.lock scanning (`-p Gopkg.lock`) has been removed. Go modules are the only supported input format.
- Running `nancy sleuth` with no piped input now auto-invokes `go list -json -deps ./...`.

When writing tests or updating examples, use `GUIDE_TOKEN` / `--guide-token` rather than OSS Index credentials.

### Are you submitting a pull request?

* Make sure to fill out an issue for your PR, so that we have traceability as to what you are trying to fix,
versus how you fixed it.
* Sign the [Sonatype CLA](https://sonatypecla.herokuapp.com/sign-cla)
* Try to fix one thing per pull request! Many people work on this code, so the more focused your changes are, the less
of a headache other people will have when they merge their work in.
* Ensure your Pull Request passes tests either locally or via CI (it will run automatically on your PR)
* Make sure to add yourself or your organization to CONTRIBUTORS.md as a part of your PR, if you are new to the project!
* If you're stuck, reach out by opening an issue on GitHub!

### Are you new and looking to dive in?

* Check our issues to see if there is something you can dive in to.
* Come hang out with us on GitHub Discussions or open an issue to introduce yourself.
