# sas

**Somehow another Sudo** — an AUR helper for Arch Linux built around privilege
separation.

sas wraps pacman for official-repository packages and adds full AUR support
(dependency resolution, PKGBUILD downloading, makepkg builds, devel package
tracking, voting). What makes it different is **how it gets root**: sas never
calls `sudo pacman` over and over, and it never keeps a sudo credential alive in
the background. Instead it starts a small privileged broker **once**, and every
pacman operation goes through it.

- **One password prompt per run.** You authenticate at the start, while you are
  at the keyboard. Long builds, suspend/resume or `timestamp_timeout=0` cannot
  trigger new prompts, so unattended upgrades never lock your account through
  `pam_faillock`.
- **No reusable credential.** sas clears the sudo timestamp right after the
  broker starts, and runs makepkg with `no_new_privs`: a PKGBUILD cannot use
  `sudo` to get root during the build.
- **A narrow, exclusive channel.** The broker only talks to the sas process that
  launched it (checked with `SO_PEERCRED` on both ends), only runs allowlisted
  pacman operations, and exits as soon as sas does.
- **Fails closed.** If the broker cannot start, sas stops. It never falls back to
  plain `sudo pacman`. Privilege separation cannot be turned off.

How it works, in depth: [doc/sas.md](doc/sas.md).

## Features

Everything you expect from an AUR helper:

- Advanced dependency solving
- PKGBUILD downloading from ABS or AUR
- Completions for AUR packages
- Query user up-front for all input (prior to starting builds)
- Narrow search (`sas linux header` will first search `linux` and then narrow on `header`)
- Find matching package providers during search and allow selection
- Remove make dependencies at the end of the build process
- Build local PKGBUILDs with AUR dependencies
- Un/Vote for packages
- Audit trail of every privileged operation in the journal (`journalctl -t sas-broker`)

## Installation

sas is built from source. You need `go`, `git` and `base-devel`:

```sh
sudo pacman -S --needed go git base-devel
git clone https://github.com/FinzoYFunzo/sas.git
cd sas
make
sudo install -Dm755 sas /usr/local/bin/sas
```

Install the binary in a root-owned location such as `/usr/local/bin`: the
broker runs this same binary as root.

sas works with `sudo`, `doas` or `run0` as privilege elevator. `su` is not
supported.

## First Use

sas keeps its own configuration in `~/.config/sas` and its cache in
`~/.cache/sas`. To reuse an existing yay configuration:

```sh
cp -r ~/.config/yay ~/.config/sas
```

#### Development packages upgrade

- Use `sas -Y --gendb` to generate a development package database for `*-git`
  packages that were installed without sas.
  This command should only be run once.

- `sas -Syu --devel` will then check for development package updates

- Use `sas -Y --devel --save` to make development package updates permanently
  enabled (`sas` and `sas -Syu` will then always check dev packages)

## Examples of Custom Operations

| Command                           | Description                                                                                   |
| --------------------------------- | --------------------------------------------------------------------------------------------- |
| `sas`                             | Alias to `sas -Syu`.                                                                          |
| `sas <Search Term>`               | Present package-installation selection menu.                                                  |
| `sas -Bi <dir>`                   | Install dependencies and build a local PKGBUILD.                                              |
| `sas -G <AUR Package>`            | Download PKGBUILD from ABS or AUR.                                                            |
| `sas -Gp <AUR Package>`           | Print to stdout PKGBUILD from ABS or AUR.                                                     |
| `sas -Ps`                         | Print system statistics.                                                                      |
| `sas -Syu --devel`                | Perform system upgrade, but also check for development package updates.                       |
| `sas -Wu <AUR Package>`           | Unvote for package (Requires setting `AUR_USERNAME` and `AUR_PASSWORD` environment variables) |
| `sas -Wv <AUR Package>`           | Vote for package (Requires setting `AUR_USERNAME` and `AUR_PASSWORD` environment variables)   |
| `sas -Y --combinedupgrade --save` | Make combined upgrade the default mode.                                                       |
| `sas -Y --gendb`                  | Generate development package database used for devel update.                                  |
| `sas -Yc`                         | Clean unneeded dependencies.                                                                  |

## Frequently Asked Questions

- **How many times will sas ask for my password?**

  Once per run, at the start. If a command only turns out to need root later
  (for example `sas -B`), sas asks at that point instead, and still only once.

- **What happened to `--sudoloop`?**

  It is not needed: there is no sudo credential to keep alive. sas accepts the
  flag and ignores it with a warning.

- **sas does not display colored output. How do I fix it?**

  Make sure you have the `Color` option in your `/etc/pacman.conf`.

- **How do I change the pager used for PKGBUILD diffs?**

  Diffs for all selected packages are collected and shown in a single pager
  session. Set `pager` in `config.json` or `sas.opt.pager` in `init.lua`, or
  set `PAGER` to override the default (`less`, or `cat` if less is unavailable).
  Example: `sas.opt.pager = "delta"` or `export PAGER=less`.

- **sas is not asking me to edit PKGBUILDS, and I don't like the diff menu! What can I do?**

  `sas --editmenu --diffmenu=false --save`

- **How can I tell sas to act only on AUR packages, or only on repo packages?**

  `sas -{OPERATION} --aur`
  `sas -{OPERATION} --repo`

- **A `Flagged Out Of Date AUR Packages` message is displayed. Why doesn't sas update them?**

  This message does not mean that updated AUR packages are available. It means
  the packages have been flagged out of date on the AUR, but
  their maintainers have not yet updated the `PKGBUILD`s
  (see [outdated AUR packages](https://wiki.archlinux.org/index.php/Arch_User_Repository#Foo_in_the_AUR_is_outdated.3B_what_should_I_do.3F)).

- **sas doesn't install dependencies added to a PKGBUILD during installation.**

  sas resolves all dependencies ahead of time. You are free to edit the
  PKGBUILD in any way, but any problems you cause are your own and should not be
  reported unless they can be reproduced with the original PKGBUILD.

- **I know my `-git` package has updates but sas doesn't offer to update it**

  sas uses a hash cache for development packages. Normally it is updated at the end of the package install with the message `Found git repo`.
  If you transition between AUR helpers and did not install the devel package using sas at some point, it is possible it never got added to the cache. `sas -Y --gendb` will fix the current version of every devel package and start checking from there.

## Support

Report issues on [GitHub](https://github.com/FinzoYFunzo/sas/issues). sas is not
officially supported by Arch Linux, so support should not be sought out on the
forums, AUR comments or other official channels.

A broken AUR package should be reported as a comment on the package's AUR page,
and only if it fails to build with plain makepkg.

## Credits and License

sas is a fork of [yay](https://github.com/Jguer/yay) by Jguer and its
contributors, and keeps its feature set. The privilege separation design and
the broker are new in sas. Like yay, sas is distributed under the
[GNU General Public License v3.0](LICENSE).

### Other AUR helpers/tools

- [yay](https://github.com/Jguer/yay)
- [paru](https://github.com/morganamilo/paru)
- [aurutils](https://github.com/AladW/aurutils)
- [pikaur](https://github.com/actionless/pikaur)
