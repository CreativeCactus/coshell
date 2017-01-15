# coshell v0.1.5

A no-frills dependency-free replacement for GNU parallel, perfect for initramfs usage.

Licensed under GNU/GPL v2.

# How it works

An ``sh -c ...`` command is started for each of the input commands; environment and current working directory are preserved.
**NOTE:** file descriptors are not

All commands will be executed, no matter which one fails.
Return value will be the sum of exit values of each command.

It is suggested to use `exec` if you want the shell-spawned process to subsitute each of the wrapping shells and be able to handle signals.
See also http://tldp.org/LDP/abs/html/process-sub.html

# Self-contained

    $ ldd coshell
    	not a dynamic executable

Thanks to [Go language](https://golang.org/), this is a self-contained executable thus a perfect match for inclusion in an initramfs or any other project where you would prefer not to have too many dependencies.

# Installation

Once you run:

    go get github.com/gdm85/coshell

The binary will be available in your ``$GOPATH/bin``; alternatively, build it with:

    make

Then copy the output ``bin/coshell`` binary to your `$PATH`, ``~/bin``, ``/usr/local/bin`` or any of your option.

# Usage

Specify each command on a single line as standard input.

Example:

    echo -e "echo test1\necho test2\necho test3" | coshell

Output:

    test3
    test1
    test2

Each line may also be prefixed with a custom string, and have substitutions applied to it.

Example:

    echo -e "a\nb\nc" | coshell -p "echo ?{3#10} : " -e "?"

Output:

    010 : a
    011 : b
    012 : c

## deinterlace option

Order is not deterministic by default, but with option ``--deinterlace`` or ``-d`` all output will be buffered and afterwards
printed in the same chronological order as process termination.

## halt-all option

If `--halt-all` or `-a` option is specified then first process to terminate unsuccessfully (with non-zero exit code) will cause 
all processes to immediately exit (including coshell) with the exit code of such process.

## master option

The `--master=n` or `-m=n` option takes a positive integer number as the index of specified command lines to identify
which process "leads" the pack: when the process exits all neighbour processes will be terminated as well and its exit code
will be adopted as coshell exit code.

## prefix option

With `-prefix "some_cmd"` or `-p "some_cmd"`, each line of input will be prepended with some_cmd. This can be used to 
check the commands before they run, for instance, with -p "echo ".

## escape option

The `-escape` or `-e` option allows substitution of certain patterns in the prefix string for each line. If the option
is given with a value like `-e "#"` then the substitution escape character will be set to "#". "?" by default. 
Any string may be used.

### escape patterns

Assuming the default escape string "?" and `-e` is provided, using `-p "some_prefix "` will yield the following, given:

    `" ?{} "`        ?{} will be substituted with the line that is given. If no ?{} is given, it is appended to the prefix.
    `" ?{#} "`       ?{#} will be substituted with the number of the line, starting with 0.
    `" ?{2#} "`      ?{2#} will be substituted with the number of the line, padded but not limited, to 2 characters. eg. 08
    `" ?{#1} "`      ?{#1} will be substituted with the number of the line, starting with 1.
    `" ?{5#5} "`     ?{5#5} will be substituted with the number of the line, starting with 5, padded to 5 or more characters.

## Examples

See [examples/](examples/) directory.
