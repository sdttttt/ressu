// Ressu is a free desktop RSS reader: the Go side fetches and parses feeds,
// keeps them in a bbolt database under the user data directory, and hands
// them to the native reader interface built on MyGO.
package main

import "github.com/sdttttt/ressu/mygoui"

func main() {
	mygoui.Run()
}
