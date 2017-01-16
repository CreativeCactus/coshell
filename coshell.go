/*
 * coshell v0.1.5 - a no-frills dependency-free replacement for GNU parallel
 * Copyright (C) 2014-2015 gdm85 - https://github.com/gdm85/coshell/

This program is free software; you can redistribute it and/or
modify it under the terms of the GNU General Public License
as published by the Free Software Foundation; either version 2
of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program; if not, write to the Free Software
Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.
*/

package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/gdm85/coshell/cosh"
)

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "coshell: %s\n", err)
	os.Exit(1)
}

func main() {
	os.Exit(coshell())
}
func coshell() int {
	var deinterlace, halt, nextMightBeMasterId, expectCommandPrefix, expectEscapeChar bool
	prefix := ""
	escape := ""
	masterId := -1
	if len(os.Args) > 1 {
		for i := 1; i < len(os.Args); i++ {
			if nextMightBeMasterId {
				nextMightBeMasterId = false
				if len(os.Args[i]) == 0 {
					fmt.Fprintf(os.Stderr, "Empty master command index specified\n")
					//					os.Exit(1)
					return 1
				}

				// clearly not a number
				if os.Args[i][0] == '-' {
					// use default
					masterId = 0
					continue
				}

				// argument is not starting with dash, number expected
				i, err := strconv.Atoi(os.Args[i])
				if err != nil || i < 0 {
					fmt.Fprintf(os.Stderr, "Invalid master command index specified: %s\n", os.Args[i])
					//					os.Exit(1)
					return 1
				}
				masterId = i
				continue
			}

			// check parameter of --master option
			var remainder string
			var found bool
			if strings.HasPrefix(os.Args[i], "--master") {
				remainder = os.Args[i][len("--master"):]
				found = true
			} else if strings.HasPrefix(os.Args[i], "-m") {
				remainder = os.Args[i][len("-m"):]
				found = true
			}
			if found {
				if len(remainder) == 0 {
					nextMightBeMasterId = true
					continue
				}
				if remainder[0] == '=' {
					remainder = remainder[1:]
				}
				i, err := strconv.Atoi(remainder)
				if err != nil || i < 0 {
					fmt.Fprintf(os.Stderr, "Invalid master command index specified: %s\n", remainder)
					//					os.Exit(1)
					return 1
				}
				masterId = i
				continue
			}

			if expectCommandPrefix {
				if len(os.Args[i]) < 1 || rune(os.Args[i][0]) == '-' {
					continue
				}
				prefix = os.Args[i]
				expectCommandPrefix = false
				continue
			}

			if expectEscapeChar {
				if len(os.Args[i]) < 1 || rune(os.Args[i][0]) == '-' {
					continue
				}
				escape = os.Args[i]
				expectEscapeChar = false
				continue
			}

			switch os.Args[i] {
			case "--help", "-h":
				fmt.Printf("coshell v0.1.4 by gdm85 - Licensed under GNU GPLv2\n")
				fmt.Printf("Usage:\n\tcoshell [--help|-h] [--deinterlace|-d] [--halt-all|-a] [--cmd|-c \"prefix-to-each-command\"] [--parse|-p[=\"$\"]] < list-of-commands\n")
				fmt.Printf("\t\t--deinterlace | -d\t\tShow individual output of processes in blocks, second order of termination\n\n")
				fmt.Printf("\t\t--halt-all | -a\t\t\tTerminate neighbour processes as soon as any has failed, using its exit code\n\n")
				fmt.Printf("\t\t--master=0 | -m=0\t\tTerminate neighbour processes as soon as command from specified input line exits and use its exit code; if no id is specified, 0 is assumed\n\n")
				fmt.Printf("\t\t--prefix \"echo \" | -p \"echo \"\tPrefix each line of input with the given string. Use -c=\"echo \", for instance, to check each line rather than execute.\n\n")
				fmt.Printf("\t\t--escape \"?\" | -e \"?\"\t\tSubstitute parts of the prefix string according to the following rules, optionally using an escape character (? by default).\n")
				fmt.Printf("\t\t\t!{}\t\t\t\tSubstitute the input line. This will remove it as the suffix unless explicitly used at the end of the prefix. May be used multiple times in the prefix.\n")
				fmt.Printf("\t\t\t!{#}\t\t\t\tSubstitute the input line number, starting with 0.\n")
				fmt.Printf("\t\t\t!{p#n}\t\t\t\tSubstitute the input line number, where p and n are both optional integers. P indicates the padded length of the substituded line number. N indicates starting value.\n")
				fmt.Printf("\nBy default, each line read from standard input will be run as a command via `sh -c`\n")
				//				os.Exit(0)
				return 0
			case "--halt-all", "-a":
				halt = true
				continue
			case "--deinterlace", "-d":
				deinterlace = true
				continue
			case "--prefix", "-p":
				expectCommandPrefix = true
				continue
			case "--escape", "-e":
				expectEscapeChar = true
				continue
			default:
				fmt.Fprintf(os.Stderr, "Invalid parameter specified: %s\n", os.Args[i])
			}
			//			os.Exit(1)
			return 1
		}
	}

	// if we never get that escape string we expected, use default.
	if expectEscapeChar {
		escape = "?"
	}

	// collect all commands to run from stdin
	var commandLines []string

	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		line = strings.TrimSuffix(line, "\n")

		if err != nil {
			if err == io.EOF {
				break
			}

			// crash in case of other errors
			fatal(err)
			return -1
		}

		// ensure that parsing is enabled and we have something to parse
		if escape != "" && prefix != "" {
			lineNumber := len(commandLines)
			line = parseLine(escape, prefix, line, lineNumber)
		} else { // if there is no parsing but still have a prefix
			line = prefix + line
		}

		commandLines = append(commandLines, line)
	}

	if len(commandLines) == 0 {
		fatal(errors.New("please specify at least 1 command in standard input"))
		return -1
	}
	if masterId != -1 && masterId >= len(commandLines) {
		fatal(errors.New("specified master command index is beyond last specified command"))
		return -1
	}

	cg := cosh.NewCommandGroup(deinterlace, halt, masterId)

	err := cg.Add(commandLines...)
	if err != nil {
		fatal(err)
		return -1
	}

	err = cg.Start()
	if err != nil {
		fatal(err)
		return -1
	}

	err, exitCode := cg.Join()
	if err != nil {
		fatal(err)
		return -1
	}

	//	os.Exit(exitCode)
	return exitCode
}

func parseLine(escape, prefix, line string, lineNumber int) string {
	// double verifying that prefix and escape are longer than 0
	if len(prefix) < 1 || len(escape) < 1 {
		return line
	}

	// if we don't have any line substitution (!{}) in the prefix, append one
	if !strings.Contains(prefix, escape+"{}") {
		prefix = prefix + escape + "{}"
	}

	// find parseable substitutions
	parts := strings.Split(prefix, escape+"{")

	for i := len(parts) - 1; i > 0; i-- {

		// check if this is just !{}, if so replace and move on
		// would be nice to just .Replace earlier, but may lead to recursion
		if parts[i][0] == '}' {
			remain := strings.TrimPrefix(parts[i], "}")
			parts[i] = line + remain
			continue
		}

		segments := strings.SplitN(parts[i], "}", 2)
		if len(segments) != 2 {
			fatal(errors.New("Malformed prefx string! Use without -e or check that substitutions match."))
		}

		subCmd := strings.TrimFunc(segments[0], func(r rune) bool {
			// return true to remove each digit character
			// NOTE: only trimming from sides. 123#3#321 => #3#
			return strings.IndexRune("0123456789", r) >= 0
		})

		// the following will give ["", ""] if no args
		args := strings.Split(segments[0], subCmd)

		if args[0] == "" {
			args[0] = "1"
		}
		if args[1] == "" {
			args[1] = "0"
		}

		// we can now check which substitution command was used
		// note we can also define ranges in the prefix using segments[1] to format uppercase, for example. Perhaps allow all but !{} in input lines?
		switch subCmd {
		case "#":
			padding := args[0]
			offset, err := strconv.Atoi(args[1])
			if err != nil {
				fatal(err) //errors.New("Malformed substitution! Use without -e or check that format is correct."))
			}
			parts[i] = fmt.Sprintf("%0"+padding+"d%s", offset+lineNumber, segments[1])
			continue
		}
	}

	return strings.Join(parts, "")
}
