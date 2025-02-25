# TTPForge Attacker Action Types

TTPForge supports the following types of actions:

- [inline:](actions/inline.md) Run Shell Commands
- [create_file:](actions/create_file.md) Create Files on Disk
- [copy_path:](actions/copy_path.md) Copy File or Directory on Disk
- [edit_file:](actions/edit_file.md) Append/Delete/Replace Lines in Files
- [remove_path:](actions/remove_path.md) Delete Files/Directories
- [fetch_uri:](actions/fetch_uri.md) Downloads a File from URL to Disk.
- [print_str:](actions/print_str.md) Print Strings to the Screen
- [file:](actions/file.md) Execute an External Program (No Shell)
- [ttp:](chaining.md) Chain Multiple TTPForge TTPs together

There is no limit on how many `steps:` a TTP can have and no restrictions on the
mix of action types that you can use in a given TTP. However, each step must map
to one and only one action type - for example, if you specify both `inline:` and
`create_file:`, you'll get an error pointing out that your step has an ambiguous
action type.
