# sas

**Somehow another Sudo** — an AUR helper for Arch Linux, written in Go, built
around privilege separation: one password prompt per run and no reusable sudo
credential.

sas wraps pacman for official-repository packages and adds full AUR support:
PKGBUILD downloading, cross-source dependency resolution, makepkg-based
building, devel/VCS package tracking, and AUR voting. A single
pacman-compatible CLI for both sources.

## Install

Build from source (requires `go`, `git` and `base-devel`):

```bash
git clone https://github.com/FinzoYFunzo/sas.git
cd sas && make
sudo install -Dm755 sas /usr/local/bin/sas
```

## Documentation

- [Manual — sas(8)](man.html) — complete command and option reference
- [Lua API](lua.html) — init.lua hooks, autocmds, and configuration
- [init.lua template](init-lua.html) — ready-to-copy configuration template

## Quick reference

```bash
sas foo                     # search and install (yogurt mode)
sas -Syu                    # full upgrade: repo + AUR
sas -Syua                   # AUR-only upgrade
sas -G foo                  # download PKGBUILD
sas -Ps                     # system statistics and health check
```

## Source

[github.com/FinzoYFunzo/sas](https://github.com/FinzoYFunzo/sas) — a fork of [yay](https://github.com/Jguer/yay).
