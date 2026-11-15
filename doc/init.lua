-- Example sas init.lua
--
-- This file is a complete template for sas.opt. Copy entries you need,
-- or keep all of them and tune values. Command-line flags still override
-- these values.

sas.opt.aururl = "https://aur.archlinux.org"
sas.opt.aurrpcurl = ""
sas.opt.build_dir = os.getenv("HOME") .. "/.cache/sas"
sas.opt.editor = os.getenv("EDITOR") or os.getenv("VISUAL") or "vi"
sas.opt.editor_flags = ""
sas.opt.pager = os.getenv("PAGER") or ""
sas.opt.makepkg_bin = "makepkg"
sas.opt.makepkg_conf = ""
sas.opt.pacman_bin = "pacman"
sas.opt.pacman_conf = "/etc/pacman.conf"
sas.opt.redownload = "no"
sas.opt.git_bin = "git"
sas.opt.gpg_bin = "gpg"
sas.opt.gpg_flags = ""
sas.opt.mflags = ""
sas.opt.sort_by = ""
sas.opt.search_by = "name-desc"
sas.opt.git_flags = ""
sas.opt.remove_make = "ask"
sas.opt.sudo_bin = "sudo"
sas.opt.sudo_flags = ""
sas.opt.rebuild = "no"
sas.opt.answer_clean = ""
sas.opt.answer_diff = ""
sas.opt.answer_edit = ""

sas.opt.request_split_n = 150
sas.opt.completion_refresh_time = 7
sas.opt.max_concurrent_downloads = 1

sas.opt.bottom_up = true
sas.opt.sudo_loop = false
sas.opt.devel = false
sas.opt.clean_after = false
sas.opt.keep_src = false
sas.opt.provides = true
sas.opt.pgp_fetch = true
sas.opt.clean_menu = true
sas.opt.diff_menu = true
sas.opt.edit_menu = false
sas.opt.combined_upgrade = true
sas.opt.use_ask = false
sas.opt.batch_install = false
sas.opt.single_line_results = false
sas.opt.separate_sources = true
sas.opt.debug = false
sas.opt.rpc = true
sas.opt.double_confirm = true

-- Hooks
-- Run Lua before sas prints the upgrade exclusion menu. Return package names
-- from event.data.upgrades to pre-exclude them. Set skip_menu = false, or omit
-- it, to show the native menu after these exclusions are applied.
--
-- sas.create_autocmd("UpgradeSelect", {
--   desc = "skip recently modified AUR upgrades",
--   callback = function(event)
--     local exclude = {}
--     local recent_cutoff = os.time() - (3 * 24 * 60 * 60)
--     for _, pkg in ipairs(event.data.upgrades) do
--       if pkg.repository == "aur" and pkg.last_modified >= recent_cutoff then
--         sas.log.warn("pre-excluding recently modified AUR package:", pkg.name)
--         table.insert(exclude, pkg.name)
--       end
--     end
--
--     return { exclude = exclude, skip_menu = false }
--   end,
-- })
--
-- Run Lua after AUR PKGBUILD repos are downloaded/merged and before the
-- clean/diff/edit menus or source downloads.
--
-- sas.create_autocmd("AURPreInstall", {
--   desc = "inspect or modify AUR package files",
--   callback = function(event)
--     if event.data.pkgbuild:match("forbidden.example") then
--       sas.log.warn(event.match .. ": forbidden source URL")
--       sas.abort(event.match .. ": forbidden source URL")
--     end
--
--     -- File edits are picked up by later menus and build steps.
--     -- local path = event.data.pkgbuild_path
--     -- local f = assert(io.open(path, "a"))
--     -- f:write("\n# edited by sas init.lua\n")
--     -- f:close()
--   end,
-- })
--
-- Run Lua after sas downloads/verifies package sources and before builds or
-- installs. AURPostDownload receives the same payload shape as AURPreInstall.
--
-- sas.create_autocmd("AURPostDownload", {
--   desc = "block forbidden source URLs after download",
--   callback = function(event)
--     if event.data.pkgbuild:match("forbidden.example") then
--       sas.abort(event.match .. ": forbidden source URL")
--     end
--   end,
-- })
--
-- Run Lua once after a successful install/upgrade transaction (skipped on
-- --downloadonly). The callback is fire-and-forget; returning has no effect.
--
-- sas.create_autocmd("PostInstall", {
--   desc = "log every package sas installed",
--   callback = function(event)
--     for _, pkg in ipairs(event.data.packages) do
--       sas.log.info(pkg.name .. " " .. pkg.version .. " (" .. pkg.source .. ")")
--     end
--   end,
-- })
--
-- Run Lua during -Ss / -S number menu after ranking, before display. Return
-- an ordered array of {source=, name=} to filter/reorder; nil = unchanged.
--
-- sas.create_autocmd("SearchFilter", {
--   desc = "show only AUR results",
--   callback = function(event)
--     local out = {}
--     for _, r in ipairs(event.data.results) do
--       if r.source == "aur" then
--         out[#out + 1] = { source = r.source, name = r.name }
--       end
--     end
--     return out
--   end,
-- })
--
-- Render a custom string for each AUR result row during -Ss / -S number menu.
-- Return a string to use it; return nil (or nothing) to fall back to default.
--
-- sas.create_autocmd("RenderAUR", {
--   desc = "compact AUR search lines",
--   callback = function(event)
--     local d = event.data
--     local tag = d.local_version ~= "" and (" [installed: " .. d.local_version .. "]") or ""
--     return string.format("aur/%s %s (+%d)%s\n    %s",
--       d.name, d.version, d.votes, tag, d.description)
--   end,
-- })
--
-- Render a custom string for each sync/repo result row during -Ss / -S number menu.
-- Return a string to use it; return nil (or nothing) to fall back to default.
--
-- sas.create_autocmd("RenderSync", {
--   desc = "compact sync search lines",
--   callback = function(event)
--     local d = event.data
--     local tag = d.local_version ~= "" and (" [installed: " .. d.local_version .. "]") or ""
--     return string.format("%s/%s %s%s\n    %s",
--       d.repository, d.name, d.version, tag, d.description)
--   end,
-- })
