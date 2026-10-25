# hatch

TODO

## Commands

- `hatch up`: apply the hatch definition to the server
- `hatch down`: revert the changes made by hatch
- `hatch status`: view information about hatch-managed resources
- `hatch config`: configuration files
    - `hatch config add <source-path> <dest-path> [--id]`
    - `hatch config rm <id>`
- `hatch service`: manage systemd definitions
    - `hatch service enable <service>`: enables a service
    - `hatch service disable <service>`: disables a service
    - `hatch service add <service-file>`: adds a service file
    - `hatch service rm <service-file>`: removes a service file
    - These commands do not make any changes to the host system. To deploy changes, use `hatch up`.
- `hatch cmd`: manage setup commands
    - `hatch cmd add [command]`: adds a command to the definition
    - `hatch run [command]`: runs the given command and adds it to the definition
    - `hatch cmd rm [command-id]`: removes a previously added command from the definition
