# Install #

    go get github.com/tgulacsi/mantis-soap/cmd/mantiscli

# Usage #
`mantiscli --mantis=https://my.mantis.host issue get 3`
```
$ mantiscli --help
usage: mantiscli [<flags>] <command> [<args> ...]

Mantis Command-Line Interface

Flags:
      --help            Show context-sensitive help (also try --help-long and --help-man).
      --mantis=MANTIS   Mantis URL
  -u, --user="$USER"  Mantis user name
      --password-env="MC_PASSWORD"
                        Environment variable's name for the password
      --config="/home/$USER/.config/mantiscli.json"
                        config file with the stored password

Commands:
  help [<command>...]
    Show help.

  issue exists [<issueid>...]
    check the existence of issues

  issue get [<issueid>...]
    get

  issue monitors [<issueid>...]
    get monitors

  issue addmonitor [<issueid>] [<plus_monitors>...]
    add monitor

  issue attach [<issueid>] [<file>]
    attach a file to the issue

  issue attachments [<issueid>]
    list attachments

  issue download [<issueid>]
    download attachemnts of the issue

  attachment add [<issueid>] [<file>]
    add attachment

  attachment list* [<issueid>]
    list attachments

  attachment download [<issueid>]
    download attachments

  monitor list* [<issueid>]
    list monitors

  monitor add [<issueid>] [<plus_monitors>...]
    add monitor

  note add* [<issueid>] [<text>...]
    add a note to an issue

  project list*
    list projects

  users list* [<flags>] [<project>]
    list users

```
