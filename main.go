package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	tcn := getFirstArgument()
	// title control numbers begin with a lower-case 'a'; we'll strip that for use in the URL
	resp, err := http.Get(fmt.Sprintf("https://iucat.iu.edu/catalog/%s/librarian_view", stripPrefixedA(tcn)))
	if err != nil {
		fmt.Println("Error fetching item:", err)
		os.Exit(1)
	}

	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusInternalServerError {
		fmt.Println("No record found for title control number", tcn)
		os.Exit(1)
	}

	body, _ := io.ReadAll(resp.Body)
	fmt.Println(tcn)
	if err = parse(body); err != nil {
		fmt.Println("Error parsing item:", err)
		os.Exit(1)
	}
}

func getFirstArgument() string {
	args := os.Args
	if len(args) < 2 {
		fmt.Println("Missing title control number.")
		fmt.Print(getHelpText())
		os.Exit(1)
	}
	if args[1] == "help" || args[1] == "-h" || args[1] == "--help" {
		fmt.Print(getHelpText())
		os.Exit(0)
	}
	return args[1]
}
func stripPrefixedA(s string) string {
	if s[0] == 'a' {
		return s[1:]
	}
	return s
}

func getHelpText() string {
	return `Please supply a title control number after the program name, example:
    catalogue-card-scraper.exe a19858379
The "a" prefix on the title control number is optional.
For help or change requests, please email brooke.weaver@gmail.com

             .--.           .---.        .-.
         .---|--|   .-.     | L |  .---. |~|    .--.
      .--|===|  |---|_|--.__| I |--|:::| |~|-==-|==|---.
      |%%|RDA|  |===| |~~|%%| L |--|   |_|~|CATS|  |___|-.
      |  |   |  |===| |==|  | L |  |:::|=| |    |BX|---|=|
      |  |   |  |   |_|__|  | Y |__|   | | |    |  |___| |
      |~~|===|--|===|~|~~|%%|~~~|--|:::|=|~|----|==|---|=|
      ^--^---'--^---^-^--^--^---'--^---^-^-^-==-^--^---^-'

`
	// ASCII art credit: https://www.asciiart.eu/books/books
}
