package shell_test

import (
	"fmt"

	"github.com/taskcluster/shell"
)

func out(args ...string) {
	fmt.Println(shell.Escape(args...))
}

func ExampleEscape_basic() {
	out(`echo`, `hello!`, `how are you doing $USER`, `"double"`, `'single'`)
	out(`'''`)
	out(`'>''`)
	out(`curl`, `-v`, `-H`, `Location;`, `-H`, `User-Agent: dave#10`, `http://www.daveeddy.com/?name=dave&age=24`)
	out(`echo`, `hello\\nworld`)
	out(`echo`, `hello:world`)
	out(`echo`, `--hello=world`)
	out(`echo`, `hello\\tworld`)
	out(`echo`, `\thello\nworld'`)
	out(`echo`, `hello  world`)
	out(`echo`, `hello`, `world`)
	out(`echo`, "hello\\\\'", "'\\\\'world")
	out(`echo`, "hello", "world\\")

	// Output:
	// echo 'hello!' 'how are you doing $USER' '"double"' \'single\'
	// \'\'\'
	// \''>'\'\'
	// curl -v -H 'Location;' -H 'User-Agent: dave#10' 'http://www.daveeddy.com/?name=dave&age=24'
	// echo 'hello\\nworld'
	// echo hello:world
	// echo --hello=world
	// echo 'hello\\tworld'
	// echo '\thello\nworld'\'
	// echo 'hello  world'
	// echo hello world
	// echo 'hello\\'\' \''\\'\''world'
	// echo hello 'world\'
}
