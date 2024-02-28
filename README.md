# Gazelle Extension for `rules_oci`

This repository hosts a Gazelle extension tailored for enhancing the capabilities
of `rules_oci` in your Bazel projects. Designed to streamline container image
building processes and improve developer productivity, this extension provides
additional functionalities and optimizations.

## Overview

The Gazelle extension for `rules_oci` empowers developers with enhanced features
for managing container images efficiently within Bazel projects. By seamlessly
integrating with `rules_oci`, it offers automatic dependency detection,
efficient layer management, and flexible configuration options to optimize
image building workflows.

## Features

- **Automatic Dependency Detection**: Automatically detects dependencies within
  your Bazel workspace to optimize image building processes.

- **Efficient Layer Management**: Intelligently manages image layers to
  minimize redundancy and optimize caching for faster builds and smaller image sizes.

- **Configuration Flexibility**: Provides fine-grained control over image
  configuration through Gazelle's configuration mechanisms, allowing easy
  customization to meet project requirements.

- **Seamless Integration**: Integrates seamlessly with existing Bazel projects
  using `rules_oci`, requiring minimal setup and configuration changes.

## Installation

To use this extension, follow these steps:

1. Clone this repository into your `MODULE.bazel`:

   ```python
   bazel_dep(name = "gazelle_oci")
   git_override(
       module_name = "gazelle_oci",
       commit = "<SHA>",
       remote = "https://github.com/pedrobarco/gazelle_oci",
   )
   ```

2. Configure Gazelle to use the extension. Add the following lines to your `BUILD.bazel`:

   ```python
   load("@gazelle//:def.bzl", "DEFAULT_LANGUAGES", "gazelle", "gazelle_binary")

   gazelle(name = "gazelle")

   gazelle_binary(
       name = "gazelle_binary",
       languages = DEFAULT_LANGUAGES + [
           # add other extensions before gazelle_oci
           "@gazelle_oci//gazelle"
       ],
   )
   ```

3. That's it! You're ready to leverage the enhanced capabilities provided by
   this extension.

## Usage

Once installed, you can use this extension just like you would use any other
gazelle extension.

<!-- TODO: directives table -->

For detailed usage instructions examples, refer to the documentation
provided in the [`examples/`](examples/) directory of this repository.

## Contributing

Contributions are welcome! If you encounter any issues, have suggestions for
improvements, or would like to contribute new features, please open an issue or
submit a pull request. Be sure to follow the contribution guidelines outlined
in the [CONTRIBUTING.md](CONTRIBUTING.md) file.

<!-- TODO: create contribution guidelines -->

## License

This project is licensed under the Apache License. See the [LICENSE](LICENSE)
file for details.

---

**Note**: For specific details and up-to-date information, refer to the
documentation and source code provided within the repository.
