# Catalogue Card Scraper
This catalogue card scraper was written for Indiana University cataloguers, to quickly generate catalogue cards to be
printed and filed as a redundancy for the electronic system. It pulls from the Indiana University
[online catalogue](https://iucat.iu.edu/catalog), on any given item's
[Librarian View page](https://iucat.iu.edu/catalog/19858379/librarian_view), generating a card with the following fields:
- 090
- 100
- 240
- 245
- 260
- 264
- 300
- 500 if contains `5|`
- 590
- 591
- 700s (any 7XX) if contains `5|`

## Usage

`$ go build .` to get the binary.

Provide a title control number as the first argument, e.g.:

`$ ./catalogue-card-scraper a20972946` should print
```
a20972946
M312.4 .G35 op.1, no.7-12 (1757)
Geminiani, Francesco, 1687-1762, composer.
Sonatas, violins (2), continuo, no. 7-12 (1757)
London : Printed for the author by J. Johnson, [1757]
4 parts ; 33 cm
With Geminiani's signature on title-page of each part, and with former owner signature ("R Lukyn") on cello part.
Sewn in modern ivory wrappers; laid in half tan calf clamshell with marbled boards.
BY: J&J Lubrano 10-28-24 $3500 law
```

`$ ./catalogue-card-scraper a19858379` should print
```
a19858379
PS3563.E747 S75 1995
Michaels, Barbara, 1927-2013. ^A101009
London : Piatkus, 1995.
307 pages ; 25 cm
From the library of Barbara Mertz, with bookplate. InU
In red boards with dark, illustrated dust jacket.
BY: Barbara Mertz library Quill & Brush en bloc 2017 law
Mertz, Barbara, former owner. InU ^A970484
```

The "a" prefix on the title control number is optional.