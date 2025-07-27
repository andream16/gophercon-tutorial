package main

import (
	"fmt"

	"github.com/andream16/gophercon-tutorial/tools/goenum/complete/currency"
)

func main() {
	fmt.Println(currency.SymbolEur, currency.SymbolEur.IsValid())
	fmt.Println(currency.SymbolGbp, currency.SymbolGbp.String())
	fmt.Println(currency.SymbolUsd)
}
