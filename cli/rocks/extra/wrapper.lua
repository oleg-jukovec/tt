local function exec(bin, ...)
    local cfg = require("luarocks.core.cfg")
    cfg.init()
    local util = require("luarocks.util")
    local cmd = require("luarocks.cmd")

    -- Tweak help messages.
    util.see_help = function(command, program) -- luacheck: no unused args
        return "\nRun \"" .. bin .. " rocks " .. command .. " --help\" for help."
    end
    util.this_program = function(default) -- luacheck: no unused args
        return bin .. " rocks"
    end

    --[[ Disabled: path, upload,
    -- init: luarocks init command generates a project, including local
    dependency management. It also creates two wrapper scripts that can be used to run
    lua & luarocks from inside the project. tt rocks init is unable to create correct luarocks
    wrapper because luarocks script, that must be wrapped, is missing.
    So, rocks init is disabled for now.
    --]]
    local rocks_commands = {
        build = "luarocks.cmd.build",
        config = "luarocks.cmd.config",
        doc = "luarocks.cmd.doc",
        download = "luarocks.cmd.download",
        help = "luarocks.cmd.help",
        install = "luarocks.cmd.install",
        lint = "luarocks.cmd.lint",
        list = "luarocks.cmd.list",
        make = "luarocks.cmd.make",
        make_manifest = "luarocks.admin.cmd.make_manifest",
        new_version = "luarocks.cmd.new_version",
        pack = "luarocks.cmd.pack",
        purge = "luarocks.cmd.purge",
        remove = "luarocks.cmd.remove",
        search = "luarocks.cmd.search",
        show = "luarocks.cmd.show",
        test = "luarocks.cmd.test",
        unpack = "luarocks.cmd.unpack",
        which = "luarocks.cmd.which",
        write_rockspec = "luarocks.cmd.write_rockspec",
    }

    cmd.run_command('LuaRocks package manager', rocks_commands, 'luarocks.cmd.external', ...)
end

return {
    exec = exec
}
