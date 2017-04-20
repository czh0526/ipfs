package commands

import (
	cmds "github.com/czh0526/ipfs/commands"
)

const (
	ApiOption = "api"
)

var Root = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:  "Global p2p merkle-day filesystem.",
		Synopsis: "ipfs [--config=<config> | -c] [ -c] [--debug=<debug>]",
		Subcommands: `
      BASIC COMMANDS
        init          Initialize ipfs local configuration
        add <path>    Add a file to IPFS
        cat <ref>     Show IPFS object data
        get <ref>     Download IPFS objects
        ls <ref>      List links from an object
        refs <ref>    List hashes of links from an object

      DATA STRUCTURE COMMANDS
        block         Interact with raw blocks in the datastore
        object        Interact with raw dag nodes
        files         Interact with objects as if they were a unix filesystem
        dag           Interact with IPLD documents (experimental)

      ADVANCED COMMANDS
        daemon        Start a long-running daemon process
        mount         Mount an IPFS read-only mountpoint
        resolve       Resolve any type of name
        name          Publish and resolve IPNS names
        key           Create and list IPNS name keypairs
        dns           Resolve DNS links
        pin           Pin objects to local storage
        repo          Manipulate the IPFS repository
        stats         Various operational stats
        filestore     Manage the filestore (experimental)

      NETWORK COMMANDS
        id            Show info about IPFS peers
        bootstrap     Add or remove bootstrap peers
        swarm         Manage connections to the p2p network
        dht           Query the DHT for values or peers
        ping          Measure the latency of a connection
        diag          Print diagnostics

      TOOL COMMANDS
        config        Manage configuration
        version       Show ipfs version information
        update        Download and apply go-ipfs updates
        commands      List all available commands

      Use 'ipfs <command> --help' to learn more about each command.

      ipfs uses a repository in the local file system. By default, the repo is located
      at ~/.ipfs. To change the repo location, set the $IPFS_PATH environment variable:

        export IPFS_PATH=/path/to/ipfsrepo

      EXIT STATUS

      The CLI will exit with one of the following values:

      0     Successful execution.
      1     Failed executions.
    `,
	},
	Options: []cmds.Option{
		cmds.StringOption("config", "c", "Path to the configuration file to use."),
		cmds.BoolOption("debug", "D", "Operate in debug mode.").Default(false),
		cmds.BoolOption("help", "Show the full command help text.").Default(false),
		cmds.BoolOption("h", "Show a short version of the command help text.").Default(false),
		cmds.BoolOption("local", "L", "Run the command locally, instead of using daemon.").Default(false),
		cmds.StringOption(ApiOption, "Use a specific API instance (defaults to /ip4/127.0.0.1/tcp/5001)"),
	},
}
