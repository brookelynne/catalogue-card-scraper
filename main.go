package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	tcn := getTitleControlNumber()
	resp, err := http.Get(fmt.Sprintf("https://iucat.iu.edu/catalog/%s/librarian_view", tcn))
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
	fmt.Println("a" + tcn)
	if err = parse(body); err != nil {
		fmt.Println("Error parsing item:", err)
		os.Exit(1)
	}
}

func getTitleControlNumber() string {
	args := os.Args
	if len(args) < 2 {
		fmt.Print(getHelpText())
		os.Exit(1)
	}
	if args[1] == "help" || args[1] == "-h" || args[1] == "--help" {
		fmt.Print(getHelpText())
		os.Exit(0)
	}
	tcn := args[1]
	// title control numbers begin with a lower-case 'a'; we'll strip that for use in the URL
	if tcn[0] == 'a' {
		tcn = tcn[1:]
	}
	return tcn
}

func getHelpText() string {
	return `Missing title control number. Please supply a title control number after the program name, example:
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
