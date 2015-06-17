package main

import (
	"io/ioutil"
)

// Creates a powershell script file with the given filename, that can be used
// to execute a command as a different user. The powershell script takes 4
// positional arguments: the windows username, password, the script to run, and
// the working directory to run from. Returns an error if there is a problem
// creating the script.
func createRunAsUserScript(filename string) error {
	scriptContents := `$username = $args[0]\r
$password = $args[1]\r
$script = $args[2]\r
$dir = $args[3]\r
\r
$credentials = New-Object System.Management.Automation.PSCredential -ArgumentList @($username,(ConvertTo-SecureString -String $password -A\r
sPlainText -Force))\r
\r
Start-Process $script -WorkingDirectory $dir -Credential ($credentials)\r
`
	return ioutil.WriteFile(filename, []byte(scriptContents), 0755)
}
